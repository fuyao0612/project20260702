package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config 保存后端启动时需要用到的配置。
//
// 现在包括：
// 1. HTTPAddr：Gin 服务监听地址，例如 :8080。
// 2. MySQL：连接 MySQL 需要的主机、端口、库名、用户名、密码。
type Config struct {
	AppEnv    string
	HTTPAddr  string
	MySQL     MySQLConfig
	JWTSecret string
	WeChat    WeChatConfig
}

// MySQLConfig 保存 MySQL 连接配置。
type MySQLConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

// WeChatConfig 保存微信小程序登录需要用到的配置。
type WeChatConfig struct {
	AppID     string
	AppSecret string

	// DevOpenID 用于本地开发。
	// 当没有配置真实 AppSecret 时，后端会用这个固定 openid 模拟微信登录。
	DevOpenID string
}

// Load 读取项目配置。
//
// godotenv.Load() 会尝试读取项目根目录下的 .env 文件。
// 如果 .env 不存在，也不会直接报错退出，因为服务器环境通常会直接注入环境变量。
func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppEnv:    getEnv("APP_ENV", "local"),
		HTTPAddr:  getEnv("HTTP_ADDR", ":8080"),
		JWTSecret: getEnv("JWT_SECRET", "project20260702_local_jwt_secret"),
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "127.0.0.1"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			Database: getEnv("MYSQL_DATABASE", "project20260702"),
			Username: getEnv("MYSQL_USERNAME", "root"),
			Password: getEnv("MYSQL_PASSWORD", ""),
		},
		WeChat: WeChatConfig{
			AppID:     getEnv("WECHAT_APP_ID", "touristappid"),
			AppSecret: getEnv("WECHAT_APP_SECRET", ""),
			DevOpenID: getEnv("WECHAT_DEV_OPENID", "dev_openid_001"),
		},
	}
}

// DSN 生成 GORM 连接 MySQL 时需要的连接字符串。
//
// 这里把拼接逻辑放在配置结构体里，是为了避免数据库包知道太多环境变量细节。
func (c MySQLConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)
}

// getEnv 读取环境变量。
//
// 如果环境变量不存在，就使用 defaultValue。
// 这样本地开发时配置不完整也能有明确的默认行为。
func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}
