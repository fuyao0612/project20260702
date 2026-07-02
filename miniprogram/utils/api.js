// API_BASE_URL 是小程序请求后端的基础地址。
//
// 本地开发时，微信开发者工具可以请求 http://127.0.0.1:8080。
// 真机体验版不能访问你电脑的 localhost，后面部署到云服务器后要改成 HTTPS 域名。
const API_BASE_URL = 'http://127.0.0.1:8080'

// request 是对 wx.request 的简单封装。
// 这样页面里不用反复写 url、method、success、fail 这些模板代码。
function request(options) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE_URL}${options.url}`,
      method: options.method || 'GET',
      data: options.data || {},
      header: {
        'content-type': 'application/json'
      },
      success(res) {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data)
          return
        }

        reject(res.data || { error: '请求失败' })
      },
      fail(err) {
        reject(err)
      }
    })
  })
}

module.exports = {
  API_BASE_URL,
  request
}
