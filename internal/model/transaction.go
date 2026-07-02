package model

import "time"

// Transaction 表示一条账单记录。
//
// 这个结构体对应 MySQL 里的 transactions 表。
// GORM 默认会把结构体名 Transaction 转成表名 transactions。
type Transaction struct {
	// ID 是主键，也就是每条账单的唯一编号。
	// json:"id" 表示返回给前端时字段名叫 id。
	// gorm:"primaryKey" 告诉 GORM 这个字段是数据库主键。
	ID uint64 `json:"id" gorm:"primaryKey"`

	// Type 表示账单类型。
	// 目前我们约定 expense 表示支出，income 表示收入。
	Type string `json:"type"`

	// Amount 表示金额，单位是“分”。
	// 例如 18 元存成 1800，这样可以避免小数计算带来的精度问题。
	Amount int `json:"amount"`

	// Category 表示分类，例如餐饮、交通、购物、工资。
	Category string `json:"category"`

	// Note 表示备注，例如午饭、地铁、买书。
	Note string `json:"note"`

	// HappenedAt 表示这笔账实际发生的时间。
	// 例如今天补录昨天的账，这里应该填昨天的时间。
	HappenedAt time.Time `json:"happened_at"`

	// CreatedAt 表示这条记录被创建到数据库的时间。
	// 我们建表时给它设置了默认值 CURRENT_TIMESTAMP，所以插入时可以由 MySQL 自动填写。
	CreatedAt time.Time `json:"created_at"`
}
