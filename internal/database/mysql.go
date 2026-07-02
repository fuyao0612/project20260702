package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"project20260702/internal/config"
)

// OpenMySQL 创建一个 MySQL 数据库连接。
//
// DSN 是 Data Source Name 的缩写，可以理解成“数据库连接地址”。
// 格式大致是：
// 用户名:密码@tcp(主机:端口)/数据库名?参数
//
// parseTime=True 很重要，它让 MySQL 的 DATETIME 能正确转换成 Go 的 time.Time。
// loc=Local 表示按本地时区解析时间。
func OpenMySQL(mysqlConfig config.MySQLConfig) (*gorm.DB, error) {
	// gorm.Open 会建立一个数据库连接对象。
	// 后面的查询、新增、修改、删除都会通过返回的 db 来操作数据库。
	return gorm.Open(mysql.Open(mysqlConfig.DSN()), &gorm.Config{})
}
