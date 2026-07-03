const { request } = require('../../utils/api')

Page({
  data: {
    loading: false,
    testing: false,
    saving: false,
    apiKeyPlaceholder: '请输入 API Key',
    protocols: ['chat_completions', 'responses'],
    protocolNames: ['Chat Completions', 'Responses'],
    protocolIndex: 0,
    form: {
      base_url: '',
      protocol: 'chat_completions',
      endpoint: '/chat/completions',
      model: '',
      api_key: ''
    }
  },

  onLoad() {
    this.loadSetting()
  },

  async loadSetting() {
    this.setData({ loading: true })

    try {
      const setting = await request({
        url: '/api/ai/settings'
      })

      const protocol = setting.protocol || 'chat_completions'
      const protocolIndex = this.data.protocols.indexOf(protocol)

      this.setData({
        apiKeyPlaceholder: setting.has_api_key ? `已配置：${setting.api_key_mask}` : '请输入 API Key',
        protocolIndex: protocolIndex >= 0 ? protocolIndex : 0,
        form: {
          base_url: setting.base_url || '',
          protocol,
          endpoint: setting.endpoint || '/chat/completions',
          model: setting.model || '',
          api_key: ''
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

  onBaseURLInput(event) {
    this.setData({ 'form.base_url': event.detail.value })
  },

  onProtocolChange(event) {
    const protocolIndex = Number(event.detail.value)
    const protocol = this.data.protocols[protocolIndex]
    const endpoint = protocol === 'responses' ? '/responses' : '/chat/completions'

    this.setData({
      protocolIndex,
      'form.protocol': protocol,
      'form.endpoint': endpoint
    })
  },

  onEndpointInput(event) {
    this.setData({ 'form.endpoint': event.detail.value })
  },

  onModelInput(event) {
    this.setData({ 'form.model': event.detail.value })
  },

  onAPIKeyInput(event) {
    this.setData({ 'form.api_key': event.detail.value })
  },

  validateForm() {
    const { form } = this.data

    if (!form.base_url.trim()) {
      return '请输入 Base URL'
    }
    if (!form.model.trim()) {
      return '请输入模型名称'
    }
    if (!form.endpoint.trim()) {
      return '请输入端点'
    }

    return ''
  },

  async testSetting() {
    const message = this.validateForm()
    if (message) {
      wx.showToast({ title: message, icon: 'none' })
      return
    }

    this.setData({ testing: true })

    try {
      await request({
        url: '/api/ai/settings/test',
        method: 'POST',
        data: this.data.form
      })

      wx.showToast({
        title: '配置可用',
        icon: 'success'
      })
    } catch (err) {
      wx.showToast({
        title: err.message || '测试失败',
        icon: 'none'
      })
    } finally {
      this.setData({ testing: false })
    }
  },

  async saveSetting() {
    const message = this.validateForm()
    if (message) {
      wx.showToast({ title: message, icon: 'none' })
      return
    }

    this.setData({ saving: true })

    try {
      await request({
        url: '/api/ai/settings',
        method: 'PUT',
        data: this.data.form
      })

      wx.showToast({
        title: '已保存',
        icon: 'success'
      })

      this.setData({
        'form.api_key': ''
      })

      this.loadSetting()
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
