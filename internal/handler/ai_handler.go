package handler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"project20260702/internal/ai"
	"project20260702/internal/config"
	"project20260702/internal/middleware"
	"project20260702/internal/model"
	"project20260702/internal/response"
)

// AIHandler 处理 AI 相关接口。
type AIHandler struct {
	db       *gorm.DB
	fallback config.AIConfig
}

// NewAIHandler 创建 AI 接口处理器。
func NewAIHandler(db *gorm.DB, fallback config.AIConfig) *AIHandler {
	return &AIHandler{
		db:       db,
		fallback: fallback,
	}
}

// transactionDraftRequest 是文本记账草稿接口的请求参数。
type transactionDraftRequest struct {
	Text string `json:"text" binding:"required"`
}

// imageTransactionDraftRequest 是图片识别记账草稿接口的请求参数。
type imageTransactionDraftRequest struct {
	ImagePath string `json:"image_path" binding:"required"`
	Text      string `json:"text"`
}

// aiSettingRequest 是保存/测试 AI 配置时的请求参数。
type aiSettingRequest struct {
	BaseURL  string `json:"base_url"`
	Protocol string `json:"protocol"`
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
}

// GetSetting 查询当前用户的 AI 配置。
//
// 对应接口：
// GET /api/ai/settings
func (h *AIHandler) GetSetting(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, 40101, "请先登录")
		return
	}

	setting, err := h.findSetting(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if setting == nil {
		response.Success(c, gin.H{
			"base_url":     h.fallback.BaseURL,
			"protocol":     normalizeProtocol(h.fallback.Protocol, h.fallback.Endpoint),
			"endpoint":     h.fallback.Endpoint,
			"model":        h.fallback.Model,
			"has_api_key":  h.fallback.APIKey != "",
			"api_key_mask": maskAPIKey(h.fallback.APIKey),
			"source":       "env",
		})
		return
	}

	response.Success(c, gin.H{
		"base_url":     setting.BaseURL,
		"protocol":     normalizeProtocol(setting.Protocol, setting.Endpoint),
		"endpoint":     setting.Endpoint,
		"model":        setting.Model,
		"has_api_key":  setting.APIKey != "",
		"api_key_mask": maskAPIKey(setting.APIKey),
		"source":       "user",
	})
}

// SaveSetting 保存当前用户的 AI 配置。
//
// 对应接口：
// PUT /api/ai/settings
func (h *AIHandler) SaveSetting(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, 40101, "请先登录")
		return
	}

	var req aiSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数不正确")
		return
	}

	normalized, err := normalizeAISettingRequest(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	setting, err := h.findSetting(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if setting == nil {
		if normalized.APIKey == "" {
			response.BadRequest(c, "首次保存时 API Key 不能为空")
			return
		}

		setting = &model.AISetting{
			UserID: userID,
		}
	}

	setting.BaseURL = normalized.BaseURL
	setting.Protocol = normalized.Protocol
	setting.Endpoint = normalized.Endpoint
	setting.Model = normalized.Model

	// API Key 为空时表示不修改旧 key，方便用户只改模型或 base url。
	if normalized.APIKey != "" {
		setting.APIKey = normalized.APIKey
	}

	if err := h.db.Save(setting).Error; err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "saved",
	})
}

// TestSetting 测试当前用户传入或已保存的 AI 配置是否可用。
//
// 对应接口：
// POST /api/ai/settings/test
func (h *AIHandler) TestSetting(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, 40101, "请先登录")
		return
	}

	var req aiSettingRequest
	_ = c.ShouldBindJSON(&req)

	runtimeConfig, err := h.runtimeConfigForUser(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 如果请求里带了字段，就用请求字段临时覆盖已保存配置。
	if strings.TrimSpace(req.BaseURL) != "" {
		runtimeConfig.BaseURL = strings.TrimSpace(req.BaseURL)
	}
	if strings.TrimSpace(req.Protocol) != "" {
		runtimeConfig.Protocol = normalizeProtocol(req.Protocol, req.Endpoint)
	}
	if strings.TrimSpace(req.Endpoint) != "" {
		runtimeConfig.Endpoint = strings.TrimSpace(req.Endpoint)
	}
	if strings.TrimSpace(req.Model) != "" {
		runtimeConfig.Model = strings.TrimSpace(req.Model)
	}
	if strings.TrimSpace(req.APIKey) != "" {
		runtimeConfig.APIKey = strings.TrimSpace(req.APIKey)
	}

	if err := ai.NewClientWithConfig(runtimeConfig).TestConnection(); err != nil {
		response.BadRequest(c, "AI 配置测试失败："+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "AI 配置可用",
	})
}

// TransactionDraft 根据自然语言生成账单草稿。
//
// 对应接口：
// POST /api/ai/transaction-draft
func (h *AIHandler) TransactionDraft(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, 40101, "请先登录")
		return
	}

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

	runtimeConfig, err := h.runtimeConfigForUser(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	draft, err := ai.NewClientWithConfig(runtimeConfig).GenerateTransactionDraft(text, time.Now(), expenseCategories, incomeCategories)
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

// ImageTransactionDraft 根据上传图片生成账单草稿。
//
// 对应接口：
// POST /api/ai/image-transaction-draft
func (h *AIHandler) ImageTransactionDraft(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, 40101, "请先登录")
		return
	}

	var req imageTransactionDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "image_path 是必填参数")
		return
	}

	imagePath, mimeType, err := safeUploadedImagePath(userID, req.ImagePath)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		response.InternalError(c, err.Error())
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

	runtimeConfig, err := h.runtimeConfigForUser(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	imageDataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imageBytes))
	draft, err := ai.NewClientWithConfig(runtimeConfig).GenerateTransactionDraftFromImage(imageDataURL, req.Text, time.Now(), expenseCategories, incomeCategories)
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

func (h *AIHandler) runtimeConfigForUser(userID uint64) (ai.RuntimeConfig, error) {
	setting, err := h.findSetting(userID)
	if err != nil {
		return ai.RuntimeConfig{}, err
	}

	if setting != nil {
		return ai.RuntimeConfig{
			BaseURL:  setting.BaseURL,
			APIKey:   setting.APIKey,
			Model:    setting.Model,
			Protocol: normalizeProtocol(setting.Protocol, setting.Endpoint),
			Endpoint: setting.Endpoint,
		}, nil
	}

	return ai.RuntimeConfig{
		BaseURL:  h.fallback.BaseURL,
		APIKey:   h.fallback.APIKey,
		Model:    h.fallback.Model,
		Protocol: normalizeProtocol(h.fallback.Protocol, h.fallback.Endpoint),
		Endpoint: h.fallback.Endpoint,
	}, nil
}

func (h *AIHandler) findSetting(userID uint64) (*model.AISetting, error) {
	var setting model.AISetting

	if err := h.db.Where("user_id = ?", userID).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &setting, nil
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

func normalizeAISettingRequest(req aiSettingRequest) (aiSettingRequest, error) {
	req.BaseURL = strings.TrimRight(strings.TrimSpace(req.BaseURL), "/")
	req.Protocol = normalizeProtocol(req.Protocol, req.Endpoint)
	req.Endpoint = strings.TrimSpace(req.Endpoint)
	req.APIKey = strings.TrimSpace(req.APIKey)
	req.Model = strings.TrimSpace(req.Model)

	if req.BaseURL == "" {
		return req, errString("Base URL 不能为空")
	}
	if req.Model == "" {
		return req, errString("模型名称不能为空")
	}
	if req.Endpoint == "" {
		req.Endpoint = defaultEndpoint(req.Protocol)
	}
	if !strings.HasPrefix(req.Endpoint, "/") {
		req.Endpoint = "/" + req.Endpoint
	}
	if !strings.HasPrefix(req.BaseURL, "http://") && !strings.HasPrefix(req.BaseURL, "https://") {
		return req, errString("Base URL 必须以 http:// 或 https:// 开头")
	}

	return req, nil
}

func normalizeProtocol(protocol string, endpoint string) string {
	protocol = strings.TrimSpace(protocol)
	if protocol == "responses" || strings.Contains(endpoint, "responses") {
		return "responses"
	}

	return "chat_completions"
}

func defaultEndpoint(protocol string) string {
	if protocol == "responses" {
		return "/responses"
	}

	return "/chat/completions"
}

func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}

	return key[:4] + "****" + key[len(key)-4:]
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

func safeUploadedImagePath(userID uint64, rawPath string) (string, string, error) {
	cleanPath := filepath.Clean(filepath.FromSlash(strings.TrimSpace(rawPath)))
	userDir := filepath.Join(uploadImageRoot, strconv.FormatUint(userID, 10))

	absUserDir, err := filepath.Abs(userDir)
	if err != nil {
		return "", "", err
	}

	absImagePath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", "", err
	}

	// 只允许读取当前登录用户自己目录下的图片，避免用户传入 ../ 读取任意文件。
	if absImagePath != absUserDir && !strings.HasPrefix(absImagePath, absUserDir+string(os.PathSeparator)) {
		return "", "", errString("图片路径不合法")
	}

	switch strings.ToLower(filepath.Ext(absImagePath)) {
	case ".jpg", ".jpeg":
		return absImagePath, "image/jpeg", nil
	case ".png":
		return absImagePath, "image/png", nil
	case ".webp":
		return absImagePath, "image/webp", nil
	default:
		return "", "", errString("只支持 jpg、png、webp 图片")
	}
}

type errString string

func (e errString) Error() string {
	return string(e)
}
