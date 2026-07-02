package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"project20260702/internal/handler"
)

// New 创建并配置 Gin 路由。
//
// 路由可以理解成“接口地址和处理函数的对应表”。
// 例如 GET /api/transactions 会交给 TransactionHandler.List 处理。
func New(db *gorm.DB) *gin.Engine {
	// gin.Default() 创建一个 Gin 路由引擎。
	// 它默认带有日志和异常恢复中间件，适合新手和项目早期使用。
	r := gin.Default()

	// GET /ping 是一个健康检查接口。
	// 我们用它判断后端服务是否启动成功。
	r.GET("/ping", func(c *gin.Context) {
		// c.JSON 会返回 JSON 数据给请求方。
		// http.StatusOK 就是 HTTP 200。
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	transactionHandler := handler.NewTransactionHandler(db)
	statisticsHandler := handler.NewStatisticsHandler(db)

	// /api 这一组路由是给小程序或前端调用的业务接口。
	api := r.Group("/api")
	{
		// 查询账单列表。
		api.GET("/transactions", transactionHandler.List)

		// 新增一条账单。
		api.POST("/transactions", transactionHandler.Create)

		// 查询一条账单详情。:id 表示这里是动态参数，例如 /transactions/1。
		api.GET("/transactions/:id", transactionHandler.Get)

		// 修改一条账单。:id 表示这里是动态参数，例如 /transactions/2。
		api.PUT("/transactions/:id", transactionHandler.Update)

		// 删除一条账单。
		api.DELETE("/transactions/:id", transactionHandler.Delete)

		// 查询月度统计数据。
		api.GET("/statistics/monthly", statisticsHandler.Monthly)
	}

	return r
}
