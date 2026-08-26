package main

import (
	"path/filepath"
	"runtime"

	"aiops/configs"
	"aiops/controllers"
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
		panic("连接数据库失败")
	}

	router := gin.Default()

	// 认证相关路由（公开）
	authCtrl := controllers.NewAuthController(db)
	router.POST("/api/auth/login", authCtrl.Login)
	router.POST("/api/auth/logout", authCtrl.Logout)
	router.GET("/api/auth/info", authCtrl.GetUserInfo)

	// 需要认证的 API
	api := router.Group("/api")
	api.Use(middleware.AuthRequired())

	// LLM 配置相关路由
	llmCtrl := controllers.NewLLMConfigController(db)
	llmGroup := api.Group("/llm-config")
	{
		llmGroup.GET("", llmCtrl.List)
		llmGroup.GET("/suppliers", llmCtrl.SupplierCategoryList)
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

	// 自动迁移（如表不存在则创建）
	db.AutoMigrate(
		&models.SysRole{},
		&models.SysAuth{},
		&models.SysRoleAuthRelation{},
		&models.SysUserRoleRelation{},
		&models.ChatSession{},
		&models.ChatMessage{},
	)

	// 启动服务器
	router.Run(cfg.Server.Port)
}
