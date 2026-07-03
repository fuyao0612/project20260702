const { request } = require('../../utils/api')

function centToYuan(amount) {
  return (amount / 100).toFixed(2)
}

function yuanToCent(value) {
  return Math.round(Number(value) * 100)
}

function formatDateForPicker(value) {
  const date = new Date(value)
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

function formatTimeForPicker(value) {
  const date = new Date(value)
  const hour = `${date.getHours()}`.padStart(2, '0')
  const minute = `${date.getMinutes()}`.padStart(2, '0')
  return `${hour}:${minute}`
}

function buildHappenedAt(date, time) {
  return `${date}T${time}:00+08:00`
}

Page({
  data: {
    text: '',
    generating: false,
    saving: false,
    draftReady: false,
    categories: [],
    categoryNames: [],
    categoryIndex: 0,
    form: {
      type: 'expense',
      amountYuan: '',
      category: '',
      note: '',
      date: '',
      time: ''
    }
  },

  onTextInput(event) {
    this.setData({
      text: event.detail.value
    })
  },

  goAISettings() {
    wx.navigateTo({
      url: '/pages/ai-settings/ai-settings'
    })
  },

  async loadCategories(type, selectedCategoryName = '') {
    const categories = await request({
      url: `/api/categories?type=${type}`
    })

    const categoryNames = categories.map((item) => item.name)
    let categoryIndex = categories.findIndex((item) => item.name === selectedCategoryName)

    if (categoryIndex < 0) {
      categoryIndex = 0
    }

    const selectedCategory = categories[categoryIndex]

    this.setData({
      categories,
      categoryNames,
      categoryIndex,
      'form.category': selectedCategory ? selectedCategory.name : ''
    })
  },

  async generateDraft() {
    const text = this.data.text.trim()

    if (!text) {
      wx.showToast({
        title: '请输入记账内容',
        icon: 'none'
      })
      return
    }

    this.setData({ generating: true })

    try {
      const draft = await request({
        url: '/api/ai/transaction-draft',
        method: 'POST',
        data: { text }
      })

      await this.loadCategories(draft.type, draft.category)

      this.setData({
        draftReady: true,
        form: {
          type: draft.type,
          amountYuan: centToYuan(draft.amount),
          category: this.data.form.category,
          note: draft.note || '',
          date: formatDateForPicker(draft.happened_at),
          time: formatTimeForPicker(draft.happened_at)
        }
      })
    } catch (err) {
      wx.showToast({
        title: err.message || '生成失败',
        icon: 'none'
      })
    } finally {
      this.setData({ generating: false })
    }
  },

  chooseType(event) {
    const type = event.currentTarget.dataset.type

    this.setData({
      'form.type': type
    })

    this.loadCategories(type)
  },

  onAmountInput(event) {
    this.setData({
      'form.amountYuan': event.detail.value
    })
  },

  onCategoryChange(event) {
    const categoryIndex = Number(event.detail.value)
    const category = this.data.categories[categoryIndex]

    this.setData({
      categoryIndex,
      'form.category': category ? category.name : ''
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

  onTimeChange(event) {
    this.setData({
      'form.time': event.detail.value
    })
  },

  async saveDraft() {
    const { form } = this.data
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
        title: '请选择分类',
        icon: 'none'
      })
      return
    }

    this.setData({ saving: true })

    try {
      await request({
        url: '/api/transactions',
        method: 'POST',
        data: {
          type: form.type,
          amount,
          category: form.category.trim(),
          note: form.note.trim(),
          happened_at: buildHappenedAt(form.date, form.time)
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
      this.setData({ saving: false })
    }
  }
})
