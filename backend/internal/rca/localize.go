package rca

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"aiops/models"

	"gorm.io/gorm"
)

// localizeWindow 拓扑定位考察的告警时间窗
const localizeWindow = 6 * time.Hour

// svcNode 拓扑定位用的服务节点（告警均为未恢复事件，不含本次触发事件自身之外的噪声恢复记录）
type svcNode struct {
	id       string
	name     string
	parentID string
	labels   map[string]string // Prometheus 标签选择器
	ownRules map[string]bool   // 近窗口内该服务自身 firing 的规则集合
	ownCount int               // 近窗口内该服务自身 firing 事件数
}

// localizeByTopology 拓扑打分定位（确定性程序侧计算，不进模型执行）：
// 1. 近窗口内 firing 告警按"事件标签 ⊇ 服务 PromLabels"归属到服务；
// 2. 沿 Service.parent_id 聚合子孙告警数得服务分；
// 3. 候选故障域 = 有告警（自身或子孙）且祖先无自身告警的"级联顶端"，按分数降序输出 Top-N。
// 无候选时返回说明文本；结果作为 [拓扑定位] 段注入综合分析上下文。
func localizeByTopology(db *gorm.DB, event models.AlertEvent, window time.Duration) string {
	// 1. 加载服务拓扑
	var services []models.Service
	if err := db.Find(&services).Error; err != nil {
		return ""
	}
	nodes := make(map[string]*svcNode, len(services))
	children := map[string][]string{}
	for _, svc := range services {
		n := &svcNode{id: svc.ID, name: svc.Name, ownRules: map[string]bool{}, labels: map[string]string{}}
		if svc.ParentID != nil {
			n.parentID = *svc.ParentID
			children[n.parentID] = append(children[n.parentID], svc.ID)
		}
		if svc.PromLabels != "" {
			_ = json.Unmarshal([]byte(svc.PromLabels), &n.labels)
		}
		nodes[svc.ID] = n
	}
	if len(nodes) == 0 {
		return ""
	}

	// 2. 近窗口 firing 告警归属到服务
	var events []models.AlertEvent
	if err := db.Where("status = ? AND trigger_time >= ?", models.AlertEventStatusFiring, time.Now().Add(-window)).
		Find(&events).Error; err != nil {
		return ""
	}
	for _, ev := range events {
		tagSet := tagStringSet(ev.Tags)
		for _, n := range nodes {
			if len(n.labels) == 0 {
				continue
			}
			if labelsSubset(n.labels, tagSet) {
				n.ownRules[ev.RuleName] = true
				n.ownCount++
			}
		}
	}

	// 3. 聚合子孙得分并找级联顶端
	score := map[string]int{}
	var collect func(id string) int
	collect = func(id string) int {
		if s, ok := score[id]; ok {
			return s
		}
		n := nodes[id]
		s := n.ownCount
		for _, c := range children[id] {
			s += collect(c)
		}
		score[id] = s
		return s
	}
	hasOwnAncestor := func(id string) bool {
		for p := nodes[id].parentID; p != ""; {
			pn, ok := nodes[p]
			if !ok {
				break
			}
			if pn.ownCount > 0 {
				return true
			}
			p = pn.parentID
		}
		return false
	}

	type candidate struct {
		name   string
		score  int
		rules  []string
		ownCnt int
	}
	var cands []candidate
	for id, n := range nodes {
		if collect(id) == 0 {
			continue
		}
		if hasOwnAncestor(id) {
			continue // 级联非顶端，由祖先代表
		}
		rules := make([]string, 0, len(n.ownRules))
		for r := range n.ownRules {
			rules = append(rules, r)
		}
		sort.Strings(rules)
		cands = append(cands, candidate{name: n.name, score: score[id], rules: rules, ownCnt: n.ownCount})
	}
	if len(cands) == 0 {
		return fmt.Sprintf("近 %s 内无其他 firing 告警，拓扑定位无候选故障域", window)
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].score > cands[j].score })
	if len(cands) > 5 {
		cands = cands[:5]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "近 %s 内 firing 告警按服务拓扑聚合的候选故障域（含子孙告警数降序）：", window)
	for _, c := range cands {
		fmt.Fprintf(&b, "\n- %s：共 %d 条告警（自身 %d 条", c.name, c.score, c.ownCnt)
		if len(c.rules) > 0 {
			fmt.Fprintf(&b, "，涉及规则：%s", strings.Join(c.rules, "、"))
		}
		b.WriteString("）")
	}
	return b.String()
}

// tagStringSet "k=v,k=v" → 集合
func tagStringSet(tags string) map[string]bool {
	set := map[string]bool{}
	for _, t := range strings.Split(tags, ",") {
		if t = strings.TrimSpace(t); t != "" {
			set[t] = true
		}
	}
	return set
}

// labelsSubset 服务的 PromLabels 是否全部包含在事件标签集合中
func labelsSubset(labels map[string]string, tagSet map[string]bool) bool {
	for k, v := range labels {
		if !tagSet[fmt.Sprintf("%s=%s", k, v)] {
			return false
		}
	}
	return true
}
