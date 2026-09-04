package controllers

import (
	"net/http"
	"time"

	"aiops/internal/inspection"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InspectionController 巡检任务 & 巡检报告 API
type InspectionController struct {
	db       *gorm.DB
	scheduler *inspection.Scheduler // 运行时注入
}

var inspectionScheduler *inspection.Scheduler

// SetInspectionScheduler 在 main.go 初始化时注入全局 scheduler
func SetInspectionScheduler(s *inspection.Scheduler) {
	inspectionScheduler = s
}

// NewInspectionController 创建 controller
func NewInspectionController(db *gorm.DB) *InspectionController {
	return &InspectionController{db: db, scheduler: inspectionScheduler}
}

// ---------- 巡检任务 ----------

func (c *InspectionController) ListTasks(ctx *gin.Context) {
	var tasks []models.InspectionTask
	if err := c.db.Order("created_at desc").Find(&tasks).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询任务失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": tasks})
}

func (c *InspectionController) GetTask(ctx *gin.Context) {
	var task models.InspectionTask
	if err := c.db.First(&task, ctx.Param("id")).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "任务不存在"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": task})
}

func (c *InspectionController) CreateTask(ctx *gin.Context) {
	var task models.InspectionTask
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 校验 cron 表达式
	if _, err := inspection.ParseCron(task.CronExpr, 1); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Cron 表达式不合法: " + err.Error()})
		return
	}

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	if err := c.db.Create(&task).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建任务失败"})
		return
	}

	// 注册到调度器
	if c.scheduler != nil && task.Enabled {
		_ = c.scheduler.Add(task)
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": task})
}

func (c *InspectionController) UpdateTask(ctx *gin.Context) {
	var task models.InspectionTask
	if err := c.db.First(&task, ctx.Param("id")).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "任务不存在"})
		return
	}

	var req struct {
		Name           string `json:"name"`
		CronExpr       string `json:"cron_expr"`
		Prompt         string `json:"prompt"`
		NotifyMediaIDs string `json:"notify_media_ids"`
		Enabled        *bool  `json:"enabled"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if req.Name != "" {
		task.Name = req.Name
	}
	if req.CronExpr != "" {
		if _, err := inspection.ParseCron(req.CronExpr, 1); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Cron 表达式不合法: " + err.Error()})
			return
		}
		task.CronExpr = req.CronExpr
	}
	if req.Prompt != "" {
		task.Prompt = req.Prompt
	}
	// 通知媒介允许清空，因此始终赋值
	task.NotifyMediaIDs = req.NotifyMediaIDs
	if req.Enabled != nil {
		task.Enabled = *req.Enabled
	}
	task.UpdatedAt = time.Now()

	if err := c.db.Save(&task).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	if c.scheduler != nil {
		_ = c.scheduler.Update(task)
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": task})
}

func (c *InspectionController) DeleteTask(ctx *gin.Context) {
	var task models.InspectionTask
	if err := c.db.First(&task, ctx.Param("id")).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "任务不存在"})
		return
	}

	if err := c.db.Delete(&task).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	if c.scheduler != nil {
		c.scheduler.Remove(task.ID)
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

func (c *InspectionController) ToggleTask(ctx *gin.Context) {
	var task models.InspectionTask
	if err := c.db.First(&task, ctx.Param("id")).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "任务不存在"})
		return
	}
	task.Enabled = !task.Enabled
	task.UpdatedAt = time.Now()

	if err := c.db.Save(&task).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	if c.scheduler != nil {
		_ = c.scheduler.Update(task)
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": task})
}

func (c *InspectionController) TriggerTask(ctx *gin.Context) {
	if c.scheduler == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "调度器未初始化"})
		return
	}

	report, err := c.scheduler.TriggerNow(ctx.Request.Context(), parseUint(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "执行失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": report})
}

func (c *InspectionController) PreviewCron(ctx *gin.Context) {
	var req struct {
		CronExpr string `json:"cron_expr" binding:"required"`
		Count    int    `json:"count"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	count := req.Count
	if count <= 0 {
		count = 5
	}
	times, err := inspection.ParseCron(req.CronExpr, count)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Cron 表达式不合法: " + err.Error()})
		return
	}

	result := make([]string, len(times))
	for i, t := range times {
		result[i] = t.Format("2006-01-02 15:04:05")
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// ---------- 巡检报告 ----------

func (c *InspectionController) ListReports(ctx *gin.Context) {
	taskID := ctx.Query("task_id")
	status := ctx.Query("status")

	db := c.db.Model(&models.InspectionReport{})
	if taskID != "" {
		db = db.Where("task_id = ?", taskID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}

	var reports []models.InspectionReport
	if err := db.Order("created_at desc").Find(&reports).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": reports})
}

func (c *InspectionController) GetReport(ctx *gin.Context) {
	var report models.InspectionReport
	if err := c.db.First(&report, ctx.Param("id")).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "报告不存在"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": report})
}

func (c *InspectionController) DeleteReport(ctx *gin.Context) {
	if err := c.db.Delete(&models.InspectionReport{}, ctx.Param("id")).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

func parseUint(s string) uint {
	var v uint
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		v = v*10 + uint(c-'0')
	}
	return v
}
