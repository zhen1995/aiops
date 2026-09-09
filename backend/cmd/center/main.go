package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"aiops/configs"
	"aiops/controllers"
	"aiops/internal/agent"
	"aiops/internal/alerting"
	"aiops/internal/denoise"
	"aiops/internal/inspection"
	"aiops/internal/knowledge"
	"aiops/middleware"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// getConfigPath 获取配置文件路径（自动定位到项目根目录）
func getConfigPath() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	dir = filepath.Join(dir, "..", "..")
	absDir, err := filepath.Abs(dir)
	if err != nil {
		panic("获取项目根目录失败: " + err.Error())
	}
	return filepath.Join(absDir, "configs", "config.yaml")
}

func main() {
	// 加载配置文件
	cfg, err := configs.LoadConfig(getConfigPath())
	if err != nil {
		panic("加载配置文件失败: " + err.Error())
	}
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败" + err.Error())
	}

	// 自动迁移（如表不存在则创建，字段变更自动加列）
	if err := db.AutoMigrate(
		// 用户与权限
		&models.SysUser{},
		&models.SysRole{},
		&models.SysAuth{},
		&models.SysRoleAuthRelation{},
		&models.SysUserRoleRelation{},
		// 对话
		&models.ChatSession{},
		&models.ChatMessage{},
		// AI 配置与数据源
		&models.LLMConfig{},
		&models.Datasource{},
		&models.AlertEngineConfig{},
		// 服务注册
		&models.Service{},
		// 告警
		&models.AlertRule{},
		&models.AlertEvent{},
		&models.RootCauseAnalysis{},
		// 巡检与通知
		&models.InspectionTask{},
		&models.InspectionReport{},
		&models.NotifyMedia{},
		&models.NotifyTemplate{},
		&models.NotifyRule{},
		// 知识库
		&models.KBDocument{},
		&models.KBChunk{},
		// 夜莺告警引擎
		&models.N9eConfig{},
		// 告警降噪
		&models.DenoisePolicy{},
		&models.DenoiseRecord{},
	); err != nil {
		panic("数据库自动迁移失败: " + err.Error())
	}

	// 历史告警规则补齐执行频率默认值（GORM 自动加列后为 0）
	db.Model(&models.AlertRule{}).Where("eval_interval = 0 OR eval_interval IS NULL").Update("eval_interval", 30)

	if err := models.SeedSysAuth(db); err != nil {
		fmt.Println("初始化权限失败:", err)
	}
	if err := models.SeedAdminRoleAuth(db); err != nil {
		fmt.Println("为 admin 角色分配权限失败:", err)
	}
	if err := controllers.EnsureDefaultTemplate(db); err != nil {
		fmt.Println("初始化默认通知模板失败:", err)
	}

	// 启动告警规则评估引擎
	alertEngine := alerting.NewEngine(db, cfg.App.FrontendBaseURL)
	if err := alertEngine.Start(); err != nil {
		fmt.Println("启动告警引擎失败:", err)
	}
	controllers.SetAlertingEngine(alertEngine)

	// 告警降噪服务：注入引擎，并在策略表为空时种子插入两条默认策略
	denoiseSvc := denoise.NewService(db)
	alertEngine.SetDenoise(denoiseSvc)
	if err := seedDenoisePolicies(db); err != nil {
		fmt.Println("初始化降噪策略失败:", err)
	}

	// 启动巡检调度器
	sched := inspection.NewScheduler(db, cfg.App.FrontendBaseURL)
	if err := sched.Start(); err != nil {
		fmt.Println("启动巡检调度器失败:", err)
	}
	controllers.SetInspectionScheduler(sched)

	router := gin.Default()

	// 认证相关路由（公开）
	authCtrl := controllers.NewAuthController(db)
	router.POST("/api/auth/login", authCtrl.Login)
	router.POST("/api/auth/logout", authCtrl.Logout)
	router.GET("/api/auth/info", authCtrl.GetUserInfo)

	// 需要认证的 API
	api := router.Group("/api")
	api.Use(middleware.AuthRequired())

	// 知识库服务（对话 Agent 工具与知识库路由共用）
	kbClient := knowledge.NewClient(cfg.Knowledge.PythonBaseURL)
	kbSvc := knowledge.NewService(db, kbClient, cfg.Knowledge.UploadDir)

	// Agent 工具注册表：注入知识库服务后再用于对话控制器
	agentRegistry := agent.NewRegistry(db)
	agentRegistry.SetKnowledge(kbSvc)

	// Chat 对话相关路由
	chatCtrl := controllers.NewChatController(db, agentRegistry)
	chatGroup := api.Group("/chat")
	{
		chatGroup.GET("/sessions", chatCtrl.ListSessions)
		chatGroup.POST("/sessions", chatCtrl.CreateSession)
		chatGroup.DELETE("/sessions/:id", chatCtrl.DeleteSession)
		chatGroup.GET("/sessions/:id/messages", chatCtrl.ListMessages)
		chatGroup.GET("/sessions/:id/stream", chatCtrl.StreamChat)
	}

	// LLM 配置相关路由
	llmCtrl := controllers.NewLLMConfigController(db)
	llmGroup := api.Group("/llm-config")
	{
		llmGroup.GET("", llmCtrl.List)
		llmGroup.GET("/suppliers", llmCtrl.SupplierCategoryList)
		llmGroup.GET("/model-types", llmCtrl.ModelTypeList)
		llmGroup.GET("/:id", llmCtrl.Get)
		llmGroup.POST("", llmCtrl.Create)
		llmGroup.PUT("/:id", llmCtrl.Update)
		llmGroup.DELETE("/:id", llmCtrl.Delete)
		llmGroup.PATCH("/:id/toggle", llmCtrl.ToggleEnabled)
	}

	// 数据源相关路由
	dsCtrl := controllers.NewDatasourceController(db)
	dsGroup := api.Group("/datasources")
	{
		dsGroup.GET("", dsCtrl.List)
		dsGroup.GET("/types", dsCtrl.TypeList)
		dsGroup.GET("/:id", dsCtrl.Get)
		dsGroup.POST("", dsCtrl.Create)
		dsGroup.PUT("/:id", dsCtrl.Update)
		dsGroup.DELETE("/:id", dsCtrl.Delete)
		dsGroup.PATCH("/:id/toggle", dsCtrl.ToggleEnabled)
	}

	// 服务注册相关路由
	serviceCtrl := controllers.NewServiceController(db)
	serviceGroup := api.Group("/services")
	{
		serviceGroup.GET("", serviceCtrl.List)
		serviceGroup.GET("/options", serviceCtrl.Options)
		serviceGroup.POST("/preview-jobs", serviceCtrl.PreviewJobs)
		serviceGroup.GET("/:id", serviceCtrl.Get)
		serviceGroup.POST("", serviceCtrl.Create)
		serviceGroup.PUT("/:id", serviceCtrl.Update)
		serviceGroup.DELETE("/:id", serviceCtrl.Delete)
		serviceGroup.PATCH("/:id/toggle", serviceCtrl.Toggle)
		serviceGroup.POST("/:id/verify", serviceCtrl.Verify)
	}

	// 告警引擎配置相关路由
	engineCtrl := controllers.NewAlertEngineController(db)
	engineGroup := api.Group("/alert-engines")
	{
		engineGroup.GET("", engineCtrl.List)
		engineGroup.GET("/:id", engineCtrl.Get)
		engineGroup.POST("", engineCtrl.Create)
		engineGroup.PUT("/:id", engineCtrl.Update)
		engineGroup.DELETE("/:id", engineCtrl.Delete)
		engineGroup.PATCH("/:id/toggle", engineCtrl.ToggleEnabled)
		engineGroup.GET("/:id/test", engineCtrl.Test)
	}

	// 告警规则相关路由（系统自建）
	ruleCtrl := controllers.NewAlertRuleController(db, alertEngine)
	ruleGroup := api.Group("/alert-rules")
	{
		ruleGroup.GET("", ruleCtrl.List)
		ruleGroup.GET("/:id", ruleCtrl.Get)
		ruleGroup.POST("", ruleCtrl.Create)
		ruleGroup.PUT("/:id", ruleCtrl.Update)
		ruleGroup.DELETE("/:id", ruleCtrl.Delete)
		ruleGroup.PATCH("/:id/toggle", ruleCtrl.ToggleEnabled)
	}

	// 告警事件代理路由
	eventCtrl := controllers.NewAlertEventController(db)
	api.GET("/alert-events", eventCtrl.List)
	api.DELETE("/alert-events/batch", eventCtrl.BatchDelete)
	api.DELETE("/alert-events/:id", eventCtrl.Delete)

	// 告警降噪相关路由
	denoiseCtrl := controllers.NewDenoiseController(db, denoiseSvc)
	denoiseGroup := api.Group("/denoise")
	{
		denoiseGroup.GET("/policies", denoiseCtrl.ListPolicies)
		denoiseGroup.PUT("/policies/:strategy", denoiseCtrl.UpdatePolicy)
		denoiseGroup.GET("/stats", denoiseCtrl.Stats)
	}

	// 总览大盘相关路由
	dashboardCtrl := controllers.NewDashboardController(db)
	api.GET("/dashboard/overview", dashboardCtrl.Overview)

	// 全局搜索聚合路由
	searchCtrl := controllers.NewSearchController(db)
	api.GET("/search", searchCtrl.Search)


	// 夜莺（Nightingale）引擎配置与告警代理
	n9eCtrl := controllers.NewN9eController(db)
	n9eGroup := api.Group("/n9e")
	{
		// 引擎配置 CRUD
		n9eGroup.GET("/configs", n9eCtrl.ListConfig)
		n9eGroup.POST("/configs", n9eCtrl.SaveConfig)
		n9eGroup.DELETE("/configs/:id", n9eCtrl.DeleteConfig)
		n9eGroup.POST("/configs/test", n9eCtrl.TestConfig)
		// 代理到夜莺的告警接口
		n9eGroup.GET("/alert-rules", n9eCtrl.GetAlertRules)
		n9eGroup.GET("/busi-groups", n9eCtrl.GetBusiGroups)
		n9eGroup.GET("/alert-cur-events", n9eCtrl.GetCurEvents)
		n9eGroup.GET("/alert-his-events", n9eCtrl.GetHisEvents)
	}

	// 根因分析相关路由（针对告警事件触发 AI 根因分析）
	rootCauseCtrl := controllers.NewRootCauseController(db)
	api.POST("/alert-events/:id/root-cause", rootCauseCtrl.Trigger)
	api.GET("/root-cause-analyses", rootCauseCtrl.List)
	api.GET("/root-cause-analyses/:id", rootCauseCtrl.Detail)

	// 巡检任务 & 巡检报告相关路由
	insCtrl := controllers.NewInspectionController(db)
	insGroup := api.Group("/inspection")
	{
		// 任务 CRUD
		insGroup.GET("/tasks", insCtrl.ListTasks)
		insGroup.GET("/tasks/:id", insCtrl.GetTask)
		insGroup.POST("/tasks", insCtrl.CreateTask)
		insGroup.PUT("/tasks/:id", insCtrl.UpdateTask)
		insGroup.DELETE("/tasks/:id", insCtrl.DeleteTask)
		insGroup.PATCH("/tasks/:id/toggle", insCtrl.ToggleTask)
		insGroup.POST("/tasks/:id/trigger", insCtrl.TriggerTask)
		// cron 表达式预览
		insGroup.POST("/cron-preview", insCtrl.PreviewCron)
		// 报告查询
		insGroup.GET("/reports", insCtrl.ListReports)
		insGroup.GET("/reports/:id", insCtrl.GetReport)
		insGroup.DELETE("/reports/:id", insCtrl.DeleteReport)
	}

	// 通知媒介相关路由
	mediaCtrl := controllers.NewNotifyMediaController(db)
	mediaGroup := api.Group("/notify-media")
	{
		mediaGroup.GET("", mediaCtrl.List)
		mediaGroup.GET("/types", mediaCtrl.TypeList)
		mediaGroup.GET("/:id", mediaCtrl.Get)
		mediaGroup.POST("", mediaCtrl.Create)
		mediaGroup.PUT("/:id", mediaCtrl.Update)
		mediaGroup.DELETE("/:id", mediaCtrl.Delete)
		mediaGroup.PATCH("/:id/toggle", mediaCtrl.ToggleEnabled)
		mediaGroup.POST("/:id/test", mediaCtrl.Test)
	}

	// 消息模板相关路由
	tplCtrl := controllers.NewNotifyTemplateController(db, cfg.App.FrontendBaseURL)
	tplGroup := api.Group("/notify-templates")
	{
		tplGroup.GET("", tplCtrl.List)
		tplGroup.GET("/sample-event", tplCtrl.SampleEvent)
		tplGroup.POST("", tplCtrl.Create)
		tplGroup.PUT("/:id", tplCtrl.Update)
		tplGroup.DELETE("/:id", tplCtrl.Delete)
		tplGroup.POST("/validate", tplCtrl.Validate)
		tplGroup.POST("/preview", tplCtrl.Preview)
	}

	// 通知规则相关路由
	notifyRuleCtrl := controllers.NewNotifyRuleController(db)
	notifyRuleGroup := api.Group("/notify-rules")
	{
		notifyRuleGroup.GET("", notifyRuleCtrl.List)
		notifyRuleGroup.POST("", notifyRuleCtrl.Create)
		notifyRuleGroup.PUT("/:id", notifyRuleCtrl.Update)
		notifyRuleGroup.DELETE("/:id", notifyRuleCtrl.Delete)
		notifyRuleGroup.PATCH("/:id/toggle", notifyRuleCtrl.Toggle)
	}

	// 运维知识库
	kbCtrl := controllers.NewKnowledgeController(db, kbSvc)
	kbGroup := api.Group("/knowledge-base")
	{
		kbGroup.GET("", kbCtrl.List)
		kbGroup.POST("", kbCtrl.Upload)
		kbGroup.POST("/retrieve", kbCtrl.Retrieve)
		kbGroup.DELETE("/:id", kbCtrl.Delete)
		kbGroup.POST("/:id/reindex", kbCtrl.Reindex)
	}

	// 用户管理相关路由
	userCtrl := controllers.NewUserController(db)
	userGroup := api.Group("/users")
	{
		userGroup.GET("", userCtrl.List)
		userGroup.GET("/:id", userCtrl.Get)
		userGroup.POST("", userCtrl.Create)
		userGroup.PUT("/:id", userCtrl.Update)
		userGroup.DELETE("/:id", userCtrl.Delete)
		userGroup.PATCH("/:id/reset-password", userCtrl.ResetPassword)
	}

	// 角色管理相关路由
	roleCtrl := controllers.NewRoleController(db)
	roleGroup := api.Group("/roles")
	{
		roleGroup.GET("", roleCtrl.List)
		roleGroup.GET("/simple", roleCtrl.ListRolesSimple)
		roleGroup.GET("/auths", roleCtrl.ListAuths)
		roleGroup.GET("/:id", roleCtrl.Get)
		roleGroup.POST("", roleCtrl.Create)
		roleGroup.PUT("/:id", roleCtrl.Update)
		roleGroup.DELETE("/:id", roleCtrl.Delete)
		roleGroup.PUT("/:id/auths", roleCtrl.SetAuths)
	}

	// 启动服务器
	router.SetTrustedProxies(nil)
	router.Run(cfg.Server.Port)
}

// seedDenoisePolicies 降噪策略表为空时种子插入两条默认策略（幂等）
func seedDenoisePolicies(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.DenoisePolicy{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	policies := []models.DenoisePolicy{
		{
			Strategy:    models.DenoiseStrategyWindowAggregation,
			Name:        "时间窗口聚合",
			Description: "5 分钟内相同告警合并为一条",
			Enabled:     true,
			Config:      `{"window_minutes":5}`,
		},
		{
			Strategy:    models.DenoiseStrategyTopologySuppression,
			Name:        "拓扑抑制",
			Description: "父节点故障抑制子节点告警",
			Enabled:     true,
			Config:      `{}`,
		},
	}
	for _, p := range policies {
		if err := db.Create(&p).Error; err != nil {
			return err
		}
	}
	return nil
}
