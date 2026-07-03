package model

import "time"

// Category 表示记账分类。
//
// 第一版分类属于系统默认分类，不区分用户。
// 后面如果要支持用户自定义分类，可以再加 user_id 字段。
type Category struct {
	ID uint64 `json:"id" gorm:"primaryKey"`

	// Type 表示分类类型。
	// expense 表示支出分类，income 表示收入分类。
	Type string `json:"type" gorm:"size:20;not null;uniqueIndex:idx_category_type_name"`

	// Name 是分类名称，例如餐饮、交通、工资。
	Name string `json:"name" gorm:"size:50;not null;uniqueIndex:idx_category_type_name"`

	// Sort 表示排序值，越小越靠前。
	Sort int `json:"sort" gorm:"not null;default:0"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
