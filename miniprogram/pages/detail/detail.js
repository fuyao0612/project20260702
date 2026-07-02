const { request } = require('../../utils/api')

// centToYuan 把后端返回的“分”转换成表单里显示的“元”。
// 例如 1800 -> 18.00。
function centToYuan(amount) {
  return (amount / 100).toFixed(2)
}

// yuanToCent 把页面输入的元转换成后端需要的分。
function yuanToCent(value) {
  return Math.round(Number(value) * 100)
}

// formatDateForPicker 把后端 RFC3339 时间转换成日期选择器需要的 YYYY-MM-DD。
function formatDateForPicker(value) {
  const date = new Date(value)
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

Page({
  // id 是当前正在编辑的账单 id。
  // loading/submitting/deleting 分别控制加载、保存、删除状态。
  data: {
    id: null,
    loading: true,
    submitting: false,
    deleting: false,
    form: {
      type: 'expense',
      amountYuan: '',
      category: '',
      note: '',
      date: ''
    }
  },

  // onLoad 可以拿到页面跳转时带来的参数。
  // 首页跳转时会传 ?id=账单ID。
  onLoad(options) {
    const id = Number(options.id)

    if (!id) {
      wx.showToast({
        title: '账单不存在',
        icon: 'none'
      })
      return
    }

    this.setData({ id })
    this.loadTransaction()
  },

  // loadTransaction 从后端加载单条账单，并填充到表单。
  async loadTransaction() {
    this.setData({ loading: true })

    try {
      const transaction = await request({
        url: `/api/transactions/${this.data.id}`
      })

      this.setData({
        form: {
          type: transaction.type,
          amountYuan: centToYuan(transaction.amount),
          category: transaction.category,
          note: transaction.note || '',
          date: formatDateForPicker(transaction.happened_at)
        }
      })
    } catch (err) {
      wx.showToast({
        title: err.message || '加载失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  chooseType(event) {
    this.setData({
      'form.type': event.currentTarget.dataset.type
    })
  },

  onAmountInput(event) {
    this.setData({
      'form.amountYuan': event.detail.value
    })
  },

  onCategoryInput(event) {
    this.setData({
      'form.category': event.detail.value
    })
  },

  onNoteInput(event) {
    this.setData({
      'form.note': event.detail.value
    })
  },

  onDateChange(event) {
    this.setData({
      'form.date': event.detail.value
    })
  },

  // submit 调用 PUT /api/transactions/:id 保存修改。
  async submit() {
    const { id, form } = this.data
    const amount = yuanToCent(form.amountYuan)

    if (!amount || amount <= 0) {
      wx.showToast({
        title: '请输入正确金额',
        icon: 'none'
      })
      return
    }

    if (!form.category.trim()) {
      wx.showToast({
        title: '请输入分类',
        icon: 'none'
      })
      return
    }

    this.setData({ submitting: true })

    try {
      await request({
        url: `/api/transactions/${id}`,
        method: 'PUT',
        data: {
          type: form.type,
          amount,
          category: form.category.trim(),
          note: form.note.trim(),
          happened_at: `${form.date}T12:00:00+08:00`
        }
      })

      wx.showToast({
        title: '已保存',
        icon: 'success'
      })

      setTimeout(() => {
        wx.navigateBack()
      }, 500)
    } catch (err) {
      wx.showToast({
        title: err.message || '保存失败',
        icon: 'none'
      })
    } finally {
      this.setData({ submitting: false })
    }
  },

  // confirmDelete 先弹确认框，避免误删。
  confirmDelete() {
    wx.showModal({
      title: '删除账单',
      content: '删除后无法恢复，确定删除吗？',
      confirmColor: '#d14343',
      success: (res) => {
        if (res.confirm) {
          this.deleteTransaction()
        }
      }
    })
  },

  // deleteTransaction 调用 DELETE /api/transactions/:id 删除账单。
  async deleteTransaction() {
    this.setData({ deleting: true })

    try {
      await request({
        url: `/api/transactions/${this.data.id}`,
        method: 'DELETE'
      })

      wx.showToast({
        title: '已删除',
        icon: 'success'
      })

      setTimeout(() => {
        wx.navigateBack()
      }, 500)
    } catch (err) {
      wx.showToast({
        title: err.message || '删除失败',
        icon: 'none'
      })
    } finally {
      this.setData({ deleting: false })
    }
  }
})
