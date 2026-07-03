const { request } = require('../../utils/api')

// formatAmount 把“分”转换成“元”。
function formatAmount(amount) {
  return (amount / 100).toFixed(2)
}

// getCurrentMonth 返回当前月份，格式是 YYYY-MM。
function getCurrentMonth() {
  const date = new Date()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  return `${date.getFullYear()}-${month}`
}

Page({
  data: {
    month: getCurrentMonth(),
    loading: false,
    incomeTotalText: '0.00',
    expenseTotalText: '0.00',
    balanceText: '0.00',
    balanceClass: 'positive',
    categoryItems: []
  },

  onLoad() {
    this.loadStatistics()
  },

  // 月份选择器变化后，重新加载该月统计。
  onMonthChange(event) {
    this.setData({
      month: event.detail.value
    })

    this.loadStatistics()
  },

  // loadStatistics 调用后端月度统计接口。
  async loadStatistics() {
    this.setData({ loading: true })

    try {
      const statistics = await request({
        url: `/api/statistics/monthly?month=${this.data.month}`
      })

      const expenseTotal = statistics.expense_total || 0
      const balance = statistics.balance || 0
      const categories = statistics.expense_by_category || []

      const categoryItems = categories.map((item) => {
        const percent = expenseTotal > 0 ? Math.round((item.amount / expenseTotal) * 100) : 0

        return {
          ...item,
          amountText: formatAmount(item.amount),
          percent,
          percentText: `${percent}%`
        }
      })

      this.setData({
        incomeTotalText: formatAmount(statistics.income_total || 0),
        expenseTotalText: formatAmount(expenseTotal),
        balanceText: formatAmount(Math.abs(balance)),
        balanceClass: balance >= 0 ? 'positive' : 'negative',
        categoryItems
      })
    } catch (err) {
      wx.showToast({
        title: err.message || '加载失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  }
})
