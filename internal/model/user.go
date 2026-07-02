package model

import "time"

// User 表示一个小程序用户。
//
// 第一版只保存微信 openid。
// 昵称、头像可以后面通过 wx.getUserProfile 或用户主动填写再补。
type User struct {
	// ID 是我们自己数据库里的用户主键。
	ID uint64 `json:"id" gorm:"primaryKey"`

	// OpenID 是微信给同一个小程序下用户分配的唯一标识。
	// 同一个用户在同一个小程序里 openid 基本保持不变。
	OpenID string `json:"openid" gorm:"uniqueIndex;size:128;not null"`

	// CreatedAt 表示用户第一次登录并创建记录的时间。
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 表示用户记录最后更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}
