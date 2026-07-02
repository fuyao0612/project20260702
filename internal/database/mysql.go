package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// OpenMySQL 创建一个 MySQL 数据库连接。
//
// DSN 是 Data Source Name 的缩写，可以理解成“数据库连接地址”。
// 格式大致是：
// 用户名:密码@tcp(主机:端口)/数据库名?参数
//
// parseTime=True 很重要，它让 MySQL 的 DATETIME 能正确转换成 Go 的 time.Time。
// loc=Local 表示按本地时区解析时间。
func OpenMySQL() (*gorm.DB, error) {
	dsn := "root:060612cjh@tcp(127.0.0.1:3306)/project20260702?charset=utf8mb4&parseTime=True&loc=Local"

	// gorm.Open 会建立一个数据库连接对象。
	// 后面的查询、新增、修改、删除都会通过返回的 db 来操作数据库。
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
