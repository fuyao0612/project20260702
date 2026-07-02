package handler

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"project20260702/internal/model"
	"project20260702/internal/response"
)

// TransactionHandler 保存账单接口需要用到的依赖。
//
// 现在它只有数据库连接 db。
// 后面业务变复杂时，还可以继续把 service 放进来。
type TransactionHandler struct {
	db *gorm.DB
}

// getTransactionID 从 URL 路径里读取账单 id。
//
// 例如请求 PUT /api/transactions/2 时：
// c.Param("id") 读到的就是字符串 "2"。
// 这里使用 ShouldBindUri，是因为 Gin 可以帮我们做“必须是数字”的基础校验。
func getTransactionID(c *gin.Context) (uint64, bool) {
	var uri struct {
		ID uint64 `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&uri); err != nil {
		response.BadRequest(c, "账单 id 不正确")
		return 0, false
	}

	return uri.ID, true
}

// NewTransactionHandler 创建账单接口处理器。
//
// 这里用构造函数传入 db，是为了避免在每个函数里重复连接数据库。
func NewTransactionHandler(db *gorm.DB) *TransactionHandler {
	return &TransactionHandler{
		db: db,
	}
}

// createTransactionRequest 表示“新增账单接口”允许前端传入的数据。
//
// 这里单独定义请求结构体，而不是直接复用 model.Transaction，是为了更安全：
// 前端不能自己传 id、created_at 这类应该由数据库控制的字段。
type createTransactionRequest struct {
	// Type 表示账单类型，只允许 expense 或 income。
	Type string `json:"type" binding:"required"`

	// Amount 表示金额，单位是分。必须大于 0。
	Amount int `json:"amount" binding:"required,gt=0"`

	// Category 表示账单分类，例如餐饮、交通、购物。
	Category string `json:"category" binding:"required"`

	// Note 表示备注，可以为空。
	Note string `json:"note"`

	// HappenedAt 表示账单发生时间。
	// 前端传 JSON 时使用 RFC3339 格式，例如：2026-07-02T12:30:00+08:00。
	HappenedAt time.Time `json:"happened_at" binding:"required"`
}

// List 查询账单列表。
//
// 对应接口：
// GET /api/transactions
func (h *TransactionHandler) List(c *gin.Context) {
	// transactions 是一个切片，可以理解成“多条账单记录的列表”。
	var transactions []model.Transaction

	// query 先从 transactions 表开始构造查询。
	// 后面如果有筛选条件，就继续往 query 上追加 Where。
	query := h.db.Model(&model.Transaction{})

	// month 是可选查询参数，格式是 YYYY-MM，例如 2026-07。
	// 当前端请求 /api/transactions?month=2026-07 时，只返回 2026 年 7 月的账单。
	month := strings.TrimSpace(c.Query("month"))
	if month != "" {
		monthStart, err := time.ParseInLocation("2006-01", month, time.Local)
		if err != nil {
			response.BadRequest(c, "month 格式不正确，正确格式类似 2026-07")
			return
		}

		// 月份筛选使用左闭右开区间：
		// happened_at >= 本月第一天 00:00:00
		// happened_at < 下个月第一天 00:00:00
		// 这样比用 <= 本月最后一天 23:59:59 更可靠。
		nextMonthStart := monthStart.AddDate(0, 1, 0)
		query = query.Where("happened_at >= ? AND happened_at < ?", monthStart, nextMonthStart)
	}

	// Order("happened_at desc") 表示按发生时间倒序排列。
	// Find(&transactions) 会执行查询，并把结果填充到 transactions 变量里。
	if err := query.Order("happened_at desc").Find(&transactions).Error; err != nil {
		// 如果数据库查询失败，返回 HTTP 500。
		// 500 表示服务器内部错误。
		response.InternalError(c, err.Error())
		return
	}

	// 查询成功后，把账单列表包装在 data 字段里返回。
	// 这种统一格式后面会方便小程序端处理。
	response.Success(c, transactions)
}

// Get 查询单条账单详情。
//
// 对应接口：
// GET /api/transactions/:id
func (h *TransactionHandler) Get(c *gin.Context) {
	id, ok := getTransactionID(c)
	if !ok {
		return
	}

	var transaction model.Transaction

	// First 会按主键查询单条记录。
	// 如果 id 不存在，GORM 会返回 gorm.ErrRecordNotFound。
	if err := h.db.First(&transaction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "账单不存在")
			return
		}

		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, transaction)
}

// Create 新增一条账单。
//
// 对应接口：
// POST /api/transactions
func (h *TransactionHandler) Create(c *gin.Context) {
	var req createTransactionRequest

	// ShouldBindJSON 会把请求体里的 JSON 解析到 req 结构体中。
	// binding 标签会做基础校验，例如必填、金额必须大于 0。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数不正确")
		return
	}

	// 账单类型先只允许支出和收入。
	// 这样可以避免数据库里出现拼错或不认识的类型。
	if req.Type != "expense" && req.Type != "income" {
		response.BadRequest(c, "type 只能是 expense 或 income")
		return
	}

	transaction := model.Transaction{
		Type:       req.Type,
		Amount:     req.Amount,
		Category:   req.Category,
		Note:       req.Note,
		HappenedAt: req.HappenedAt,
	}

	// Create 会执行 INSERT，把这条账单写入 MySQL。
	// 写入成功后，GORM 会把数据库生成的 id 填回 transaction.ID。
	if err := h.db.Create(&transaction).Error; err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 新增成功通常返回 HTTP 201 Created。
	// data 里返回刚刚创建出来的账单，方便前端立刻更新页面。
	response.Created(c, transaction)
}

// Update 修改一条账单。
//
// 对应接口：
// PUT /api/transactions/:id
func (h *TransactionHandler) Update(c *gin.Context) {
	id, ok := getTransactionID(c)
	if !ok {
		return
	}

	var req createTransactionRequest

	// 修改账单时，前端也需要传完整账单内容。
	// 这样第一版逻辑更直观：用新的 type、amount、category、note、happened_at 覆盖旧值。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数不正确")
		return
	}

	if req.Type != "expense" && req.Type != "income" {
		response.BadRequest(c, "type 只能是 expense 或 income")
		return
	}

	var transaction model.Transaction

	// First 会按主键查询单条记录。
	// 如果 id 不存在，GORM 会返回 gorm.ErrRecordNotFound。
	if err := h.db.First(&transaction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "账单不存在")
			return
		}

		response.InternalError(c, err.Error())
		return
	}

	transaction.Type = req.Type
	transaction.Amount = req.Amount
	transaction.Category = req.Category
	transaction.Note = req.Note
	transaction.HappenedAt = req.HappenedAt

	// Save 会把修改后的结构体保存回数据库。
	// 因为 transaction 已经有 ID，GORM 会执行 UPDATE，而不是 INSERT。
	if err := h.db.Save(&transaction).Error; err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, transaction)
}

// Delete 删除一条账单。
//
// 对应接口：
// DELETE /api/transactions/:id
func (h *TransactionHandler) Delete(c *gin.Context) {
	id, ok := getTransactionID(c)
	if !ok {
		return
	}

	var transaction model.Transaction

	if err := h.db.First(&transaction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "账单不存在")
			return
		}

		response.InternalError(c, err.Error())
		return
	}

	// Delete 会删除这条记录。
	// 现在我们的表没有软删除字段，所以这里执行的是物理删除。
	if err := h.db.Delete(&transaction).Error; err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 删除成功后不需要返回完整数据，告诉前端删除成功即可。
	response.Success(c, gin.H{
		"message": "deleted",
	})
}
