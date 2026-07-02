// package main 表示这是一个可以直接运行的 Go 程序。
// Go 程序的入口函数固定叫 main()。
package main

import (
	"log"

	"project20260702/internal/database"
	"project20260702/internal/model"
	"project20260702/internal/router"
)

func main() {
	// 连接 MySQL。
	// 如果数据库连不上，后端继续运行也没有意义，所以这里直接退出。
	db, err := database.OpenMySQL()
	if err != nil {
		log.Fatal("connect mysql failed: ", err)
	}

	// AutoMigrate 会根据 Go 结构体自动同步数据库表结构。
	// 现在它会确保 MySQL 里存在 transactions 表。
	// 项目早期用它很方便；后期上线后，我们会再学习更严谨的数据库迁移工具。
	if err := db.AutoMigrate(&model.Transaction{}); err != nil {
		log.Fatal("migrate database failed: ", err)
	}

	// 创建 Gin 路由，并把数据库连接传进去。
	// 后面的接口处理函数就可以通过这个 db 查询或写入数据。
	r := router.New(db)

	// 启动 HTTP 服务，监听 8080 端口。
	// 启动后可以通过 http://127.0.0.1:8080 访问这个后端。
	if err := r.Run(":8080"); err != nil {
		log.Fatal("start server failed: ", err)
	}
}
