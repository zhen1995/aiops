package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"aiops/internal/alerting"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AlertRuleController 告警规则控制器
type AlertRuleController struct {
	DB     *gorm.DB
	engine *alerting.Engine
}

// NewAlertRuleController 创建控制器
func NewAlertRuleController(db *gorm.DB, engine *alerting.Engine) *AlertRuleController {
	return &AlertRuleController{DB: db, engine: engine}
}

// ruleWithGroup 规则响应体：附加业务分组名称（冗余展示字段，同 AlertEvent.GroupName 模式）
type ruleWithGroup struct {
	models.AlertRule
	GroupName string `json:"group_name"`
}

// attachGroupNames 批量查询分组名称，为规则列表附加 group_name 响应字段
func (c *AlertRuleController) attachGroupNames(list []models.AlertRule) []ruleWithGroup {
	items := make([]ruleWithGroup, 0, len(list))
	groupIDs := make([]string, 0, len(list))
	for _, r := range list {
		if r.GroupID != "" {
			groupIDs = append(groupIDs, r.GroupID)
		}
	}
	nameMap := map[string]string{}
	if len(groupIDs) > 0 {
		var groups []models.BusinessGroup
		if err := c.DB.Where("id IN ?", groupIDs).Find(&groups).Error; err == nil {
			for _, g := range groups {
				nameMap[g.ID] = g.Name
			}
		}
	}
	for _, r := range list {
		items = append(items, ruleWithGroup{AlertRule: r, GroupName: nameMap[r.GroupID]})
	}
	return items
}

// List 获取告警规则列表（分页）
// 支持 ?page=&limit=&group_id=&query=（规则名称/PromQL 模糊过滤）；响应中附加 group_name 便于前端展示
func (c *AlertRuleController) List(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "page 参数必须是正整数"})
		return
	}
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 200 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "limit 参数必须是 1-200 的整数"})
		return
	}

	db := c.DB.Model(&models.AlertRule{})
	if gid := ctx.Query("group_id"); gid != "" {
		db = db.Where("group_id = ?", gid)
	}
	if q := strings.TrimSpace(ctx.Query("query")); q != "" {
		like := "%" + q + "%"
		db = db.Where("name LIKE ? OR prom_ql LIKE ?", like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	var list []models.AlertRule
	if err := db.Order("created_at desc").
		Offset((page - 1) * limit).Limit(limit).
		Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  c.attachGroupNames(list),
			"total": total,
		},
	})
}

// Get 根据 ID 获取单个告警规则
func (c *AlertRuleController) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	var rule models.AlertRule
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "告警规则不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": c.attachGroupNames([]models.AlertRule{rule})[0],
	})
}

// validateRule 校验规则必填字段；excludeID 用于更新时排除自身做名称唯一性校验
func (c *AlertRuleController) validateRule(rule *models.AlertRule, excludeID string) error {
	if rule.Name == "" {
		return errors.New("规则名称不能为空")
	}
	if rule.PromQL == "" {
		return errors.New("PromQL 不能为空")
	}
	if rule.Severity < 1 || rule.Severity > 3 {
		return errors.New("告警级别必须是 1(P1紧急)、2(P2警告) 或 3(P3提醒)")
	}
	if rule.Duration < 0 {
		return errors.New("持续时间不能为负数")
	}
	if rule.EvalInterval <= 0 {
		return errors.New("执行频率必须是正整数（秒）")
	}
	if rule.GroupID != "" {
		var count int64
		if err := c.DB.Model(&models.BusinessGroup{}).Where("id = ?", rule.GroupID).Count(&count).Error; err != nil {
			return errors.New("校验业务分组失败")
		}
		if count == 0 {
			return errors.New("业务分组不存在")
		}
	}
	// 规则名称唯一性校验（软删除记录不占用名称）
	var dup int64
	q := c.DB.Model(&models.AlertRule{}).Where("name = ?", rule.Name)
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&dup).Error; err != nil {
		return errors.New("校验规则名称失败")
	}
	if dup > 0 {
		return errors.New("规则名称已存在")
	}
	return nil
}

// Create 创建告警规则
func (c *AlertRuleController) Create(ctx *gin.Context) {
	var rule models.AlertRule
	if err := ctx.ShouldBindJSON(&rule); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	if err := c.validateRule(&rule, ""); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	if err := c.DB.Create(&rule).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建失败",
			"error":   err.Error(),
		})
		return
	}

	// 启用状态的规则立即纳入评估引擎
	c.engine.Add(rule)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": rule,
	})
}

// Update 更新告警规则
func (c *AlertRuleController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var rule models.AlertRule
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "告警规则不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	var updateData models.AlertRule
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	if err := c.validateRule(&updateData, id); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	updates := map[string]interface{}{
		"name":                    updateData.Name,
		"prom_ql":                 updateData.PromQL,
		"eval_interval":           updateData.EvalInterval,
		"duration":                updateData.Duration,
		"severity":                updateData.Severity,
		"group_id":                updateData.GroupID,
		"notify_rule_id":          updateData.NotifyRuleID,
		"repeat_interval_minutes": updateData.RepeatIntervalMinutes,
		"max_send_count":          updateData.MaxSendCount,
		"is_enabled":              updateData.IsEnabled,
	}

	if err := c.DB.Model(&rule).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败",
			"error":   err.Error(),
		})
		return
	}

	// 重新查询返回最新数据
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	// 重启该规则的评估循环
	c.engine.Update(rule)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": c.attachGroupNames([]models.AlertRule{rule})[0],
	})
}

// BatchDelete 批量删除告警规则（软删除）：逐条停止评估并将未恢复事件置为已恢复
func (c *AlertRuleController) BatchDelete(ctx *gin.Context) {
	var payload struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil || len(payload.IDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误：ids 不能为空"})
		return
	}

	deleted := 0
	for _, id := range payload.IDs {
		var rule models.AlertRule
		if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
			continue
		}
		if err := c.DB.Delete(&rule).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败", "error": err.Error()})
			return
		}
		c.engine.Remove(id)
		deleted++
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功", "data": gin.H{"deleted": deleted}})
}

// CloneToGroup 将选中的告警规则克隆到指定业务分组：逐条复制为新规则，名称冲突时自动追加“-副本”后缀
func (c *AlertRuleController) CloneToGroup(ctx *gin.Context) {
	var payload struct {
		IDs     []string `json:"ids" binding:"required"`
		GroupID string   `json:"group_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil || len(payload.IDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误：ids 与 group_id 不能为空"})
		return
	}

	// 目标业务分组必须存在
	var group models.BusinessGroup
	if err := c.DB.First(&group, "id = ?", payload.GroupID).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "目标业务分组不存在"})
		return
	}

	// 生成未被占用的规则名称：原名 → 原名-副本 → 原名-副本2 …
	uniqueName := func(base string) (string, error) {
		candidate := base + "-副本"
		for i := 2; ; i++ {
			var count int64
			if err := c.DB.Model(&models.AlertRule{}).Where("name = ?", candidate).Count(&count).Error; err != nil {
				return "", err
			}
			if count == 0 {
				return candidate, nil
			}
			if i > 99 {
				return "", errors.New("无法生成可用副本名称（重名过多）")
			}
			candidate = fmt.Sprintf("%s-副本%d", base, i)
		}
	}

	cloned := 0
	var firstErr error
	for _, id := range payload.IDs {
		var src models.AlertRule
		if err := c.DB.First(&src, "id = ?", id).Error; err != nil {
			continue
		}
		name, err := uniqueName(src.Name)
		if err != nil {
			firstErr = err
			break
		}
		copy := models.AlertRule{
			Name:                  name,
			PromQL:                src.PromQL,
			Severity:              src.Severity,
			Duration:              src.Duration,
			EvalInterval:          src.EvalInterval,
			GroupID:               payload.GroupID,
			NotifyRuleID:          src.NotifyRuleID,
			RepeatIntervalMinutes: src.RepeatIntervalMinutes,
			MaxSendCount:          src.MaxSendCount,
			IsEnabled:             src.IsEnabled,
		}
		if err := c.DB.Create(&copy).Error; err != nil {
			firstErr = err
			break
		}
		// 启用状态的副本直接纳入评估引擎
		if copy.IsEnabled == 1 {
			c.engine.Update(copy)
		}
		cloned++
	}

	if firstErr != nil && cloned == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "克隆失败", "error": firstErr.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "克隆成功", "data": gin.H{"cloned": cloned}})
}

// Delete 删除告警规则（软删除）
func (c *AlertRuleController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	var rule models.AlertRule
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "告警规则不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	if err := c.DB.Delete(&rule).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败",
			"error":   err.Error(),
		})
		return
	}

	// 停止该规则的评估，并将其未恢复事件置为已恢复
	c.engine.Remove(id)

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ToggleEnabled 切换启用状态
func (c *AlertRuleController) ToggleEnabled(ctx *gin.Context) {
	id := ctx.Param("id")

	var rule models.AlertRule
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "告警规则不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	newStatus := 0
	if rule.IsEnabled == 0 {
		newStatus = 1
	}

	if err := c.DB.Model(&rule).Update("is_enabled", newStatus).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "操作失败",
			"error":   err.Error(),
		})
		return
	}

	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	// 启停该规则的评估循环
	c.engine.Update(rule)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": c.attachGroupNames([]models.AlertRule{rule})[0],
	})
}
