package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"project20260702/internal/model"
	"project20260702/internal/response"
)

// CategoryHandler 处理分类相关接口。
type CategoryHandler struct {
	db *gorm.DB
}

// NewCategoryHandler 创建分类接口处理器。
func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{
		db: db,
	}
}

// List 查询分类列表。
//
// 对应接口：
// GET /api/categories?type=expense
func (h *CategoryHandler) List(c *gin.Context) {
	categoryType := strings.TrimSpace(c.Query("type"))
	if categoryType != "expense" && categoryType != "income" {
		response.BadRequest(c, "type 只能是 expense 或 income")
		return
	}

	var categories []model.Category

	if err := h.db.Where("type = ?", categoryType).
		Order("sort asc, id asc").
		Find(&categories).Error; err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, categories)
}
