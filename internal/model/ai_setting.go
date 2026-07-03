package model

import "time"

// AISetting 表示某个用户自己的 AI API 配置。
//
// API Key 只保存在后端数据库里，小程序读取配置时不会返回完整 key。
// 当前学习版使用数据库保存配置；正式部署时还应进一步配合数据库权限、备份加密等措施。
type AISetting struct {
	ID uint64 `json:"id" gorm:"primaryKey"`

	// UserID 表示这套 AI 配置属于哪个用户。
	UserID uint64 `json:"user_id" gorm:"uniqueIndex;not null"`

	BaseURL  string `json:"base_url" gorm:"size:255;not null"`
	Protocol string `json:"protocol" gorm:"size:30;not null;default:chat_completions"`
	Endpoint string `json:"endpoint" gorm:"size:100;not null;default:/chat/completions"`
	Model    string `json:"model" gorm:"size:100;not null"`

	// APIKey 保存真实 key。接口读取配置时不会返回这个字段。
	APIKey string `json:"-" gorm:"type:text;not null"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
