const { request, uploadFile } = require('../../utils/api')

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
    imageGenerating: false,
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

      await this.applyDraft(draft)
    } catch (err) {
      wx.showToast({
        title: err.message || '生成失败',
        icon: 'none'
      })
    } finally {
      this.setData({ generating: false })
    }
  },

  async chooseImageAndGenerateDraft() {
    if (this.data.imageGenerating) {
      return
    }

    try {
      const media = await new Promise((resolve, reject) => {
        wx.chooseMedia({
          count: 1,
          mediaType: ['image'],
          sourceType: ['album', 'camera'],
          success: resolve,
          fail: reject
        })
      })

      const filePath = media.tempFiles && media.tempFiles[0] ? media.tempFiles[0].tempFilePath : ''
      if (!filePath) {
        wx.showToast({
          title: '没有选择图片',
          icon: 'none'
        })
        return
      }

      this.setData({ imageGenerating: true })

      // 第一步：把图片上传到 Go 后端，后端会保存到 uploads/images 目录。
      const uploadResult = await uploadFile({
        url: '/api/uploads/images',
        filePath,
        name: 'file'
      })

      // 第二步：把后端返回的图片路径交给 AI 识别接口，生成账单草稿。
      const draft = await request({
        url: '/api/ai/image-transaction-draft',
        method: 'POST',
        data: {
          image_path: uploadResult.path,
          text: this.data.text.trim()
        }
      })

      await this.applyDraft(draft)
    } catch (err) {
      wx.showToast({
        title: err.message || '图片识别失败',
        icon: 'none'
      })
    } finally {
      this.setData({ imageGenerating: false })
    }
  },

  async applyDraft(draft) {
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
