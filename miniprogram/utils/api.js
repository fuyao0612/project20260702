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
        const body = res.data || {}

        // 后端现在统一返回 { code, message, data }。
        // code === 0 表示业务成功，此时把真正业务数据 body.data 交给页面。
        if (res.statusCode >= 200 && res.statusCode < 300 && body.code === 0) {
          resolve(body.data)
          return
        }

        reject({
          message: body.message || '请求失败',
          code: body.code
        })
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
