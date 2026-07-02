const { request } = require('../../utils/api')

// formatAmount 把“分”转换成页面展示用的“元”。
// 例如 1800 -> 18.00。
function formatAmount(amount) {
  return (amount / 100).toFixed(2)
}

// formatDate 把后端返回的时间转换成简短日期。
function formatDate(value) {
  const date = new Date(value)
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${month}-${day}`
}

// getCurrentMonth 返回当前月份，格式是 YYYY-MM。
// 月度统计接口需要这个参数。
function getCurrentMonth() {
  const date = new Date()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  return `${date.getFullYear()}-${month}`
}

Page({
  // data 是页面状态。
  // WXML 里使用的 transactions、incomeTotalText、expenseTotalText 都来自这里。
  data: {
    loading: false,
    transactions: [],
    incomeTotalText: '0.00',
    expenseTotalText: '0.00'
  },

  // onLoad 是页面首次加载时触发的生命周期函数。
  // 用户第一次进入首页时，在这里加载账单和统计数据。
  onLoad() {
    this.loadPageData()
  },

  onShow() {
    // 从新增页面返回首页时，重新加载数据。
    this.loadPageData()
  },

  // onPullDownRefresh 对应用户下拉刷新。
  // 数据加载完成后要调用 wx.stopPullDownRefresh() 停止顶部刷新动画。
  onPullDownRefresh() {
    this.loadPageData().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 跳转到新增账单页面。
  goCreate() {
    wx.navigateTo({
      url: '/pages/create/create'
    })
  },

  // 跳转到账单详情页。
  // 首页列表项上通过 data-id 保存账单 id，这里从 event.currentTarget.dataset 取出来。
  goDetail(event) {
    const id = event.currentTarget.dataset.id

    wx.navigateTo({
      url: `/pages/detail/detail?id=${id}`
    })
  },

  // loadPageData 负责加载首页所需的全部数据。
  // 当前包括：本月账单列表、本月统计。
  async loadPageData() {
    this.setData({ loading: true })

    try {
      const month = getCurrentMonth()

      // Promise.all 可以同时请求账单列表和统计数据，减少等待时间。
      const [listResult, statisticsResult] = await Promise.all([
        request({ url: `/api/transactions?month=${month}` }),
        request({ url: `/api/statistics/monthly?month=${month}` })
      ])

      const transactions = (listResult || []).map((item) => ({
        ...item,
        amountText: formatAmount(item.amount),
        dateText: formatDate(item.happened_at)
      }))

      const statistics = statisticsResult || {}

      // setData 会更新页面状态，并触发 WXML 重新渲染。
      this.setData({
        transactions,
        incomeTotalText: formatAmount(statistics.income_total || 0),
        expenseTotalText: formatAmount(statistics.expense_total || 0)
      })
    } catch (err) {
      // wx.showToast 是微信小程序常用的轻提示。
      // icon: 'none' 表示只显示文字，不显示成功/失败图标。
      wx.showToast({
        title: err.message || '加载失败',
        icon: 'none'
      })
    } finally {
      // finally 无论请求成功还是失败都会执行。
      // 这里用来关闭 loading 状态。
      this.setData({ loading: false })
    }
  }
})
