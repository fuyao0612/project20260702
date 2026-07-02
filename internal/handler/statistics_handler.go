package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"project20260702/internal/model"
)

// StatisticsHandler 保存统计接口需要用到的依赖。
type StatisticsHandler struct {
	db *gorm.DB
}

// NewStatisticsHandler 创建统计接口处理器。
func NewStatisticsHandler(db *gorm.DB) *StatisticsHandler {
	return &StatisticsHandler{
		db: db,
	}
}

// monthlyCategorySummary 表示某个月内，一个分类的支出汇总。
//
// 这个结构体不是数据库表，而是 SQL 聚合查询的结果。
type monthlyCategorySummary struct {
	Category string `json:"category"`
	Amount   int64  `json:"amount"`
}

// monthlyTypeSummary 表示某个月内，收入或支出的总额。
//
// Type 是 expense 或 income，Amount 是该类型的总金额，单位是分。
type monthlyTypeSummary struct {
	Type   string `json:"type"`
	Amount int64  `json:"amount"`
}

// parseMonthQuery 解析 month 查询参数。
//
// month 必须是 YYYY-MM 格式，例如 2026-07。
// 返回值 monthStart 是本月第一天 00:00:00，nextMonthStart 是下个月第一天 00:00:00。
func parseMonthQuery(c *gin.Context) (month string, monthStart time.Time, nextMonthStart time.Time, ok bool) {
	month = strings.TrimSpace(c.Query("month"))
	if month == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "month 是必填参数，格式类似 2026-07",
		})
		return "", time.Time{}, time.Time{}, false
	}

	parsedMonth, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "month 格式不正确，正确格式类似 2026-07",
		})
		return "", time.Time{}, time.Time{}, false
	}

	return month, parsedMonth, parsedMonth.AddDate(0, 1, 0), true
}

// Monthly 查询月度统计。
//
// 对应接口：
// GET /api/statistics/monthly?month=2026-07
func (h *StatisticsHandler) Monthly(c *gin.Context) {
	month, monthStart, nextMonthStart, ok := parseMonthQuery(c)
	if !ok {
		return
	}

	var typeSummaries []monthlyTypeSummary

	// 按 type 分组统计总金额。
	// SELECT type, SUM(amount) FROM transactions WHERE ... GROUP BY type
	if err := h.db.Model(&model.Transaction{}).
		Select("type, COALESCE(SUM(amount), 0) AS amount").
		Where("happened_at >= ? AND happened_at < ?", monthStart, nextMonthStart).
		Group("type").
		Scan(&typeSummaries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var incomeTotal int64
	var expenseTotal int64

	for _, summary := range typeSummaries {
		switch summary.Type {
		case "income":
			incomeTotal = summary.Amount
		case "expense":
			expenseTotal = summary.Amount
		}
	}

	var categorySummaries []monthlyCategorySummary

	// 只对支出做分类汇总。
	// 记账软件里，分类饼图通常看的是“钱花在哪里”。
	if err := h.db.Model(&model.Transaction{}).
		Select("category, COALESCE(SUM(amount), 0) AS amount").
		Where("type = ? AND happened_at >= ? AND happened_at < ?", "expense", monthStart, nextMonthStart).
		Group("category").
		Order("amount desc").
		Scan(&categorySummaries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"month":               month,
			"income_total":        incomeTotal,
			"expense_total":       expenseTotal,
			"balance":             incomeTotal - expenseTotal,
			"expense_by_category": categorySummaries,
		},
	})
}
