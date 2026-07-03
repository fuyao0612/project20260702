package handler

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"project20260702/internal/ai"
	"project20260702/internal/model"
	"project20260702/internal/response"
)

// AIHandler 处理 AI 相关接口。
type AIHandler struct {
	db       *gorm.DB
	aiClient *ai.Client
}

// NewAIHandler 创建 AI 接口处理器。
func NewAIHandler(db *gorm.DB, aiClient *ai.Client) *AIHandler {
	return &AIHandler{
		db:       db,
		aiClient: aiClient,
	}
}

// transactionDraftRequest 是文本记账草稿接口的请求参数。
type transactionDraftRequest struct {
	Text string `json:"text" binding:"required"`
}

// TransactionDraft 根据自然语言生成账单草稿。
//
// 对应接口：
// POST /api/ai/transaction-draft
func (h *AIHandler) TransactionDraft(c *gin.Context) {
	var req transactionDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "text 是必填参数")
		return
	}

	text := strings.TrimSpace(req.Text)
	if text == "" {
		response.BadRequest(c, "text 不能为空")
		return
	}

	expenseCategories, err := h.categoryNames("expense")
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	incomeCategories, err := h.categoryNames("income")
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	draft, err := h.aiClient.GenerateTransactionDraft(text, time.Now(), expenseCategories, incomeCategories)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if err := validateDraft(draft, expenseCategories, incomeCategories); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, draft)
}

func (h *AIHandler) categoryNames(categoryType string) ([]string, error) {
	var categories []model.Category

	if err := h.db.Where("type = ?", categoryType).
		Order("sort asc, id asc").
		Find(&categories).Error; err != nil {
		return nil, err
	}

	names := make([]string, 0, len(categories))
	for _, category := range categories {
		names = append(names, category.Name)
	}

	return names, nil
}

func validateDraft(draft ai.TransactionDraft, expenseCategories []string, incomeCategories []string) error {
	if draft.Type != "expense" && draft.Type != "income" {
		return errString("AI 返回的 type 不正确")
	}

	if draft.Amount <= 0 {
		return errString("AI 返回的金额不正确")
	}

	if _, err := time.Parse(time.RFC3339, draft.HappenedAt); err != nil {
		return errString("AI 返回的时间格式不正确")
	}

	allowedCategories := expenseCategories
	if draft.Type == "income" {
		allowedCategories = incomeCategories
	}

	for _, category := range allowedCategories {
		if draft.Category == category {
			return nil
		}
	}

	return errString("AI 返回的分类不在可选分类中")
}

type errString string

func (e errString) Error() string {
	return string(e)
}
