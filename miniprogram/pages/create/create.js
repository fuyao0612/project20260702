const { request } = require('../../utils/api')

// getToday 返回今天的日期，格式是 YYYY-MM-DD。
// 新增账单页面默认把账单日期设为今天。
function getToday() {
  const date = new Date()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

// yuanToCent 把页面输入的元转换成后端需要的分。
// 使用 Math.round 是为了支持 18.5 这种输入，得到 1850 分。
function yuanToCent(value) {
  return Math.round(Number(value) * 100)
}

Page({
  // data 是页面状态。
  // submitting 控制保存按钮的 loading 状态。
  // form 保存用户正在填写的表单内容。
  data: {
    submitting: false,
    form: {
      type: 'expense',
      amountYuan: '',
      category: '',
      note: '',
      date: getToday()
    }
  },

  // chooseType 根据用户点击的分段按钮，切换支出/收入。
  chooseType(event) {
    this.setData({
      'form.type': event.currentTarget.dataset.type
    })
  },

  // 金额输入框变化时，把最新输入值同步到 form.amountYuan。
  onAmountInput(event) {
    this.setData({
      'form.amountYuan': event.detail.value
    })
  },

  // 分类输入框变化时，把最新输入值同步到 form.category。
  onCategoryInput(event) {
    this.setData({
      'form.category': event.detail.value
    })
  },

  // 备注输入框变化时，把最新输入值同步到 form.note。
  onNoteInput(event) {
    this.setData({
      'form.note': event.detail.value
    })
  },

  // 日期选择器变化时，把选择结果同步到 form.date。
  onDateChange(event) {
    this.setData({
      'form.date': event.detail.value
    })
  },

  // submit 负责校验表单并调用后端新增账单接口。
  async submit() {
    const { form } = this.data
    const amount = yuanToCent(form.amountYuan)

    // 金额为空、不是数字、或者小于等于 0，都不允许提交。
    if (!amount || amount <= 0) {
      wx.showToast({
        title: '请输入正确金额',
        icon: 'none'
      })
      return
    }

    // 分类是必填项。
    if (!form.category.trim()) {
      wx.showToast({
        title: '请输入分类',
        icon: 'none'
      })
      return
    }

    this.setData({ submitting: true })

    try {
      // 这里调用后端 POST /api/transactions。
      // 后端要求 amount 使用“分”，happened_at 使用 RFC3339 时间格式。
      await request({
        url: '/api/transactions',
        method: 'POST',
        data: {
          type: form.type,
          amount,
          category: form.category.trim(),
          note: form.note.trim(),
          // 当前页面只选择日期，没有选择具体时间。
          // 所以先固定为当天 12:00:00，后面可以升级成日期+时间选择。
          happened_at: `${form.date}T12:00:00+08:00`
        }
      })

      wx.showToast({
        title: '已保存',
        icon: 'success'
      })

      // 等 Toast 显示一小会儿，再返回首页。
      // 首页 onShow 会重新加载数据，所以新增后能看到新账单。
      setTimeout(() => {
        wx.navigateBack()
      }, 500)
    } catch (err) {
      wx.showToast({
        title: err.error || '保存失败',
        icon: 'none'
      })
    } finally {
      // 无论保存成功还是失败，都关闭按钮 loading。
      this.setData({ submitting: false })
    }
  }
})
