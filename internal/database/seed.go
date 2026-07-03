package database

import (
	"gorm.io/gorm"

	"project20260702/internal/model"
)

// SeedDefaultCategories 初始化系统默认分类。
//
// FirstOrCreate 会先查询分类是否存在，不存在才创建。
// 所以这个函数可以在每次服务启动时执行，不会重复插入同名分类。
func SeedDefaultCategories(db *gorm.DB) error {
	categories := []model.Category{
		{Type: "expense", Name: "餐饮", Sort: 10},
		{Type: "expense", Name: "交通", Sort: 20},
		{Type: "expense", Name: "购物", Sort: 30},
		{Type: "expense", Name: "学习", Sort: 40},
		{Type: "expense", Name: "娱乐", Sort: 50},
		{Type: "expense", Name: "医疗", Sort: 60},
		{Type: "expense", Name: "住房", Sort: 70},
		{Type: "expense", Name: "其他", Sort: 999},
		{Type: "income", Name: "工资", Sort: 10},
		{Type: "income", Name: "兼职", Sort: 20},
		{Type: "income", Name: "红包", Sort: 30},
		{Type: "income", Name: "理财", Sort: 40},
		{Type: "income", Name: "其他", Sort: 999},
	}

	for _, category := range categories {
		if err := db.Where("type = ? AND name = ?", category.Type, category.Name).
			FirstOrCreate(&category).Error; err != nil {
			return err
		}
	}

	return nil
}
