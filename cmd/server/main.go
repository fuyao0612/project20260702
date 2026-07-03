// package main 表示这是一个可以直接运行的 Go 程序。
// Go 程序的入口函数固定叫 main()。
package main

import (
	"log"

	"project20260702/internal/config"
	"project20260702/internal/database"
	"project20260702/internal/model"
	"project20260702/internal/router"
)

func main() {
	// 读取配置。
	// 本地开发时主要来自 .env；部署到服务器后也可以来自系统环境变量。
	cfg := config.Load()

	// 连接 MySQL。
	// 如果数据库连不上，后端继续运行也没有意义，所以这里直接退出。
	db, err := database.OpenMySQL(cfg.MySQL)
	if err != nil {
		log.Fatal("connect mysql failed: ", err)
	}

	// AutoMigrate 会根据 Go 结构体自动同步数据库表结构。
	// 现在它会确保 MySQL 里存在 transactions 表。
	// 项目早期用它很方便；后期上线后，我们会再学习更严谨的数据库迁移工具。
	if err := db.AutoMigrate(&model.User{}, &model.Category{}, &model.Transaction{}); err != nil {
		log.Fatal("migrate database failed: ", err)
	}

	// 初始化默认分类。
	// 这样新环境启动后，小程序分类选择器就有基础分类可用。
	if err := database.SeedDefaultCategories(db); err != nil {
		log.Fatal("seed default categories failed: ", err)
	}

	// 创建 Gin 路由，并把数据库连接传进去。
	// 后面的接口处理函数就可以通过这个 db 查询或写入数据。
	r := router.New(db, cfg)

	// 启动 HTTP 服务，监听 8080 端口。
	// 启动后可以通过 http://127.0.0.1:8080 访问这个后端。
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal("start server failed: ", err)
	}
}
