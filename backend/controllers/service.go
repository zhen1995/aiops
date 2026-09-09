package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"aiops/internal/datasource"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// codeRegexp 服务编码格式：小写字母/数字/中划线，不能以中划线开头结尾
var codeRegexp = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// promLabelKeyRegexp 合法的 Prom 标签名：[a-zA-Z_][a-zA-Z0-9_]*
var promLabelKeyRegexp = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// buildPromSelector 将标签选择器构造成 {k1="v1",k2="v2"} 形式，供 PromQL 查询使用
func buildPromSelector(labels map[string]string) string {
	return datasource.BuildPromSelector(labels)
}

// ServiceController 服务注册控制器
type ServiceController struct {
	DB *gorm.DB
}

// NewServiceController 创建控制器
func NewServiceController(db *gorm.DB) *ServiceController {
	return &ServiceController{DB: db}
}

// serviceForm 服务注册入参（索引模式/标签以结构化形式接收，入库时序列化为 JSON 字符串）
type serviceForm struct {
	Name             string            `json:"name"`
	Code             string            `json:"code"`
	ESDatasourceID   string            `json:"es_datasource_id"`
	ESIndexPatterns  []string          `json:"es_index_patterns"`
	PromDatasourceID string            `json:"prom_datasource_id"`
	PromLabels       map[string]string `json:"prom_labels"`
	ParentID         string            `json:"parent_id"`
	PyroscopeApp     string            `json:"pyroscope_app"`
	Owner            string            `json:"owner"`
	Description      string            `json:"description"`
	Status           int               `json:"status"`
}

// serviceView 服务注册出参（索引模式/标签反序列化为结构化形式）
type serviceView struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Code             string            `json:"code"`
	ESDatasourceID   string            `json:"es_datasource_id"`
	ESIndexPatterns  []string          `json:"es_index_patterns"`
	PromDatasourceID string            `json:"prom_datasource_id"`
	PromLabels       map[string]string `json:"prom_labels"`
	ParentID         *string           `json:"parent_id"`
	PyroscopeApp     string            `json:"pyroscope_app"`
	Owner            string            `json:"owner"`
	Description      string            `json:"description"`
	Status           int               `json:"status"`
	Verified         int               `json:"verified"`
	VerifyMessage    string            `json:"verify_message"`
	LastVerifiedAt   *time.Time        `json:"last_verified_at"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// toView 将模型转换为出参结构，JSON 字符串字段反序列化
func toView(s *models.Service) serviceView {
	view := serviceView{
		ID:               s.ID,
		Name:             s.Name,
		Code:             s.Code,
		ESDatasourceID:   s.ESDatasourceID,
		PromDatasourceID: s.PromDatasourceID,
		ParentID:         s.ParentID,
		PyroscopeApp:     s.PyroscopeApp,
		Owner:            s.Owner,
		Description:      s.Description,
		Status:           s.Status,
		Verified:         s.Verified,
		VerifyMessage:    s.VerifyMessage,
		LastVerifiedAt:   s.LastVerifiedAt,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
	// 索引模式反序列化，失败时按空数组返回
	if s.ESIndexPatterns != "" {
		var patterns []string
		if err := json.Unmarshal([]byte(s.ESIndexPatterns), &patterns); err == nil {
			view.ESIndexPatterns = patterns
		}
	}
	if view.ESIndexPatterns == nil {
		view.ESIndexPatterns = []string{}
	}
	// 标签反序列化，失败时按空对象返回
	if s.PromLabels != "" {
		var labels map[string]string
		if err := json.Unmarshal([]byte(s.PromLabels), &labels); err == nil {
			view.PromLabels = labels
		}
	}
	if view.PromLabels == nil {
		view.PromLabels = map[string]string{}
	}
	return view
}

// validateService 校验服务注册入参，并检查数据源引用是否存在且类型匹配
func (c *ServiceController) validateService(form *serviceForm, excludeID string) error {
	if strings.TrimSpace(form.Name) == "" {
		return errors.New("服务名称不能为空")
	}
	if strings.TrimSpace(form.Code) == "" {
		return errors.New("服务编码不能为空")
	}
	if !codeRegexp.MatchString(form.Code) {
		return errors.New("服务编码格式不正确：仅限小写字母、数字和中划线，且不能以中划线开头或结尾")
	}
	// 编码唯一性校验（排除自身与已软删除记录）
	var count int64
	q := c.DB.Model(&models.Service{}).Where("code = ?", form.Code)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return fmt.Errorf("校验服务编码失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("服务编码 %s 已存在", form.Code)
	}
	// ES 数据源引用校验：配置了数据源则索引模式必填
	if form.ESDatasourceID != "" {
		if len(form.ESIndexPatterns) == 0 {
			return errors.New("已配置 ElasticSearch 数据源，索引模式不能为空")
		}
		var ds models.Datasource
		if err := c.DB.First(&ds, "id = ?", form.ESDatasourceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("ElasticSearch 数据源不存在")
			}
			return fmt.Errorf("查询数据源失败: %w", err)
		}
		if ds.Type != datasource.TypeElasticsearch {
			return fmt.Errorf("数据源 %s 不是 ElasticSearch 类型", ds.Name)
		}
	}
	// Prometheus 数据源引用校验：配置了数据源则标签选择器必填且键名合法
	if form.PromDatasourceID != "" {
		if len(form.PromLabels) == 0 {
			return errors.New("已配置 Prometheus 数据源，标签选择器不能为空")
		}
		for k := range form.PromLabels {
			if !promLabelKeyRegexp.MatchString(k) {
				return fmt.Errorf("Prometheus 标签名 %s 不合法：需匹配 [a-zA-Z_][a-zA-Z0-9_]*", k)
			}
		}
		var ds models.Datasource
		if err := c.DB.First(&ds, "id = ?", form.PromDatasourceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("Prometheus 数据源不存在")
			}
			return fmt.Errorf("查询数据源失败: %w", err)
		}
		if ds.Type != datasource.TypePrometheus {
			return fmt.Errorf("数据源 %s 不是 Prometheus 类型", ds.Name)
		}
	}
	return nil
}

// applyForm 将入参序列化后写入模型
func applyForm(s *models.Service, form *serviceForm) {
	s.Name = strings.TrimSpace(form.Name)
	s.Code = form.Code
	s.ESDatasourceID = form.ESDatasourceID
	s.PromDatasourceID = form.PromDatasourceID
	// 父服务：空串置 nil，非空写入指针值
	if strings.TrimSpace(form.ParentID) == "" {
		s.ParentID = nil
	} else {
		pid := strings.TrimSpace(form.ParentID)
		s.ParentID = &pid
	}
	s.PyroscopeApp = strings.TrimSpace(form.PyroscopeApp)
	s.Owner = strings.TrimSpace(form.Owner)
	s.Description = strings.TrimSpace(form.Description)
	if form.Status == 1 || form.Status == 0 {
		s.Status = form.Status
	}
	// 索引模式序列化为 JSON 数组字符串，无值时存空串
	if len(form.ESIndexPatterns) > 0 {
		if b, err := json.Marshal(form.ESIndexPatterns); err == nil {
			s.ESIndexPatterns = string(b)
		}
	} else {
		s.ESIndexPatterns = ""
	}
	// 标签序列化为 JSON 对象字符串，无值时存空串
	if len(form.PromLabels) > 0 {
		if b, err := json.Marshal(form.PromLabels); err == nil {
			s.PromLabels = string(b)
		}
	} else {
		s.PromLabels = ""
	}
}

// List 分页查询服务列表，支持关键字与状态过滤
func (c *ServiceController) List(ctx *gin.Context) {
	page, err := parseInt(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "page 参数必须是正整数"})
		return
	}
	pageSize, err := parseInt(ctx.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "page_size 参数必须是正整数"})
		return
	}

	db := c.DB.Model(&models.Service{})
	if q := strings.TrimSpace(ctx.Query("keyword")); q != "" {
		like := "%" + q + "%"
		db = db.Where("name LIKE ? OR code LIKE ? OR owner LIKE ?", like, like, like)
	}
	if s := ctx.Query("status"); s != "" {
		status, err := parseInt(s)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "status 参数必须是整数"})
			return
		}
		db = db.Where("status = ?", status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询服务列表失败", "error": err.Error()})
		return
	}

	var list []models.Service
	if err := db.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询服务列表失败", "error": err.Error()})
		return
	}

	views := make([]serviceView, 0, len(list))
	for i := range list {
		views = append(views, toView(&list[i]))
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": views, "total": total}})
}

// Get 根据 ID 获取单个服务
func (c *ServiceController) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	var svc models.Service
	if err := c.DB.First(&svc, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "服务不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": toView(&svc)})
}

// Create 创建服务
func (c *ServiceController) Create(ctx *gin.Context) {
	var form serviceForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	if err := c.validateService(&form, ""); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	var svc models.Service
	applyForm(&svc, &form)
	if err := c.DB.Create(&svc).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": toView(&svc)})
}

// Update 更新服务
func (c *ServiceController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var svc models.Service
	if err := c.DB.First(&svc, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "服务不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败", "error": err.Error()})
		return
	}

	var form serviceForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	if err := c.validateService(&form, id); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	// 防止把父服务设成自身（简单排除自身，不做深度环校验）
	if strings.TrimSpace(form.ParentID) != "" && strings.TrimSpace(form.ParentID) == id {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "父服务不能选择服务自身"})
		return
	}

	applyForm(&svc, &form)
	// 数据位置变更后旧探测结论失效，重置探测状态
	svc.Verified = 0
	svc.VerifyMessage = ""
	svc.LastVerifiedAt = nil
	if err := c.DB.Save(&svc).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": toView(&svc)})
}

// Delete 删除服务（软删除）
func (c *ServiceController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	var svc models.Service
	if err := c.DB.First(&svc, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "服务不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败", "error": err.Error()})
		return
	}

	if err := c.DB.Delete(&svc).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// Toggle 切换服务启用状态
func (c *ServiceController) Toggle(ctx *gin.Context) {
	id := ctx.Param("id")

	var svc models.Service
	if err := c.DB.First(&svc, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "服务不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败", "error": err.Error()})
		return
	}

	newStatus := 0
	if svc.Status == 0 {
		newStatus = 1
	}
	if err := c.DB.Model(&svc).Update("status", newStatus).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "操作失败", "error": err.Error()})
		return
	}

	if err := c.DB.First(&svc, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": toView(&svc)})
}

// Options 返回启用服务的精简列表（供下拉选择）
func (c *ServiceController) Options(ctx *gin.Context) {
	var list []models.Service
	if err := c.DB.Where("status = ?", 1).Order("name ASC").Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询服务列表失败", "error": err.Error()})
		return
	}
	options := make([]gin.H, 0, len(list))
	for _, svc := range list {
		options = append(options, gin.H{"id": svc.ID, "name": svc.Name, "code": svc.Code})
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": options})
}

// previewJobsForm PreviewJobs 入参（按标签选择器预览命中的 job 列表）
type previewJobsForm struct {
	PromDatasourceID string            `json:"prom_datasource_id"`
	PromLabels       map[string]string `json:"prom_labels"`
}

// PreviewJobs 根据标签选择器预览其命中的 Prometheus job 列表：
// 执行 count by (job) (up{selector})，收集样本中的 job 去重排序返回；
// 探测失败也返回 200，由前端直接展示 message。
func (c *ServiceController) PreviewJobs(ctx *gin.Context) {
	var form previewJobsForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	if strings.TrimSpace(form.PromDatasourceID) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "prom_datasource_id 不能为空"})
		return
	}
	if len(form.PromLabels) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "prom_labels 不能为空"})
		return
	}
	for k := range form.PromLabels {
		if !promLabelKeyRegexp.MatchString(k) {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("Prometheus 标签名 %s 不合法：需匹配 [a-zA-Z_][a-zA-Z0-9_]*", k)})
			return
		}
	}
	var ds models.Datasource
	if err := c.DB.First(&ds, "id = ?", form.PromDatasourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Prometheus 数据源不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询数据源失败", "error": err.Error()})
		return
	}
	if ds.Type != datasource.TypePrometheus || ds.IsEnabled != 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("数据源 %s 不是已启用的 Prometheus 类型", ds.Name)})
		return
	}

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client := datasource.NewPrometheusClient(&ds)
	query := "count by (job) (up" + buildPromSelector(form.PromLabels) + ")"
	res, err := client.QueryInstant(reqCtx, query)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"jobs": []string{}, "message": "探测失败: " + err.Error()}})
		return
	}
	jobSet := make(map[string]bool)
	for _, sample := range res.Samples {
		if job := sample.Labels["job"]; job != "" {
			jobSet[job] = true
		}
	}
	jobs := make([]string, 0, len(jobSet))
	for job := range jobSet {
		jobs = append(jobs, job)
	}
	sort.Strings(jobs)
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"jobs":    jobs,
		"message": fmt.Sprintf("查询 %s 命中 %d 个 job", query, len(jobs)),
	}})
}

// Verify 探测服务配置的数据源是否真正有数据：
// ES 对每个索引模式执行 count（至少一个有数据即通过），Prometheus 执行 count(up{标签选择器}) 即时查询。
// 未配置的数据源视为跳过、不判失败；探测失败只记录不报错。
func (c *ServiceController) Verify(ctx *gin.Context) {
	id := ctx.Param("id")

	var svc models.Service
	if err := c.DB.First(&svc, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "服务不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败", "error": err.Error()})
		return
	}

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Elasticsearch 探测：任一索引模式有文档即通过；未配置则跳过
	esOK := true
	esMessage := "未配置 ElasticSearch 数据源，跳过探测"
	if svc.ESDatasourceID != "" {
		var ds models.Datasource
		if err := c.DB.First(&ds, "id = ?", svc.ESDatasourceID).Error; err != nil {
			esOK = false
			esMessage = "ElasticSearch 数据源不存在或已删除"
		} else {
			var patterns []string
			if err := json.Unmarshal([]byte(svc.ESIndexPatterns), &patterns); err != nil {
				esOK = false
				esMessage = "索引模式配置不是合法的 JSON 数组"
			} else {
				client := datasource.NewElasticsearchClient(&ds)
				hitPatterns := make([]string, 0, len(patterns))
				var lastErr error
				for _, pattern := range patterns {
					n, err := client.Count(reqCtx, pattern)
					if err != nil {
						lastErr = err
						continue
					}
					if n > 0 {
						hitPatterns = append(hitPatterns, fmt.Sprintf("%s(%d条)", pattern, n))
					}
				}
				if len(hitPatterns) > 0 {
					esOK = true
					esMessage = "索引有数据: " + strings.Join(hitPatterns, ", ")
				} else if lastErr != nil {
					esOK = false
					esMessage = "探测失败: " + lastErr.Error()
				} else {
					esOK = false
					esMessage = "所有索引模式均无数据"
				}
			}
		}
	}

	// Prometheus 探测：count(up{标签选择器}) 有结果即通过；未配置或标签反序列化失败按未配置处理
	promOK := true
	promMessage := "未配置 Prometheus 数据源，跳过探测"
	if svc.PromDatasourceID != "" {
		var labels map[string]string
		if err := json.Unmarshal([]byte(svc.PromLabels), &labels); err != nil || len(labels) == 0 {
			promMessage = "标签选择器未配置，跳过探测"
		} else {
			var ds models.Datasource
			if err := c.DB.First(&ds, "id = ?", svc.PromDatasourceID).Error; err != nil {
				promOK = false
				promMessage = "Prometheus 数据源不存在或已删除"
			} else {
				client := datasource.NewPrometheusClient(&ds)
				query := fmt.Sprintf("count(up%s)", buildPromSelector(labels))
				res, err := client.QueryInstant(reqCtx, query)
				if err != nil {
					promOK = false
					promMessage = "探测失败: " + err.Error()
				} else if res.Hit {
					promOK = true
					promMessage = fmt.Sprintf("查询 %s 命中 %d 个实例", query, len(res.Samples))
				} else {
					promOK = false
					promMessage = fmt.Sprintf("查询 %s 无结果", query)
				}
			}
		}
	}

	verified := 0
	if esOK && promOK {
		verified = 1
	}
	verifyMessage := esMessage
	if svc.PromDatasourceID != "" {
		verifyMessage = esMessage + "；Prometheus: " + promMessage
	}

	now := time.Now()
	updates := map[string]interface{}{
		"verified":         verified,
		"verify_message":   verifyMessage,
		"last_verified_at": &now,
	}
	if err := c.DB.Model(&svc).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存探测结果失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"es_ok":        esOK,
			"es_message":   esMessage,
			"prom_ok":      promOK,
			"prom_message": promMessage,
			"verified":     verified,
		},
	})
}
