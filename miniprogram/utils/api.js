// API_BASE_URL 是小程序请求后端的基础地址。
//
// 本地开发时，微信开发者工具可以请求 http://127.0.0.1:8080。
// 真机体验版不能访问你电脑的 localhost，后面部署到云服务器后要改成 HTTPS 域名。
const API_BASE_URL = 'http://127.0.0.1:8080'

// TOKEN_STORAGE_KEY 是 token 在微信本地缓存里的 key。
const TOKEN_STORAGE_KEY = 'BOOKKEEPING_TOKEN'

// rawRequest 是对 wx.request 的最底层封装。
//
// 它只负责发请求，不负责自动登录。
// 这样 login() 自己调用登录接口时，不会递归触发自己。
function rawRequest(options) {
  return new Promise((resolve, reject) => {
    const token = wx.getStorageSync(TOKEN_STORAGE_KEY)

    wx.request({
      url: `${API_BASE_URL}${options.url}`,
      method: options.method || 'GET',
      data: options.data || {},
      header: {
        'content-type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {})
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

// login 调用微信登录，并把后端返回的 token 存到本地。
//
// wx.login() 会拿到一个临时 code。
// 后端用这个 code 换 openid，再生成我们自己系统的 token。
function login() {
  return new Promise((resolve, reject) => {
    wx.login({
      success: async (res) => {
        if (!res.code) {
          reject({ message: '微信登录失败' })
          return
        }

        try {
          const data = await rawRequest({
            url: '/api/auth/wechat-login',
            method: 'POST',
            data: {
              code: res.code
            }
          })

          wx.setStorageSync(TOKEN_STORAGE_KEY, data.token)
          resolve(data)
        } catch (err) {
          reject(err)
        }
      },
      fail(err) {
        reject(err)
      }
    })
  })
}

// ensureLogin 确保本地已有 token。
//
// 第一版只判断有没有 token。
// 如果 token 过期，request 收到 401 后会清掉 token 并重新登录一次。
async function ensureLogin() {
  const token = wx.getStorageSync(TOKEN_STORAGE_KEY)
  if (token) {
    return
  }

  await login()
}

// request 是页面实际使用的请求函数。
//
// 它会先确保已登录，再请求业务接口。
// 如果后端返回 401，说明 token 失效，会重新登录并重试一次。
async function request(options) {
  await ensureLogin()

  try {
    return await rawRequest(options)
  } catch (err) {
    if (err.code === 40101) {
      wx.removeStorageSync(TOKEN_STORAGE_KEY)
      await login()
      return rawRequest(options)
    }

    throw err
  }
}

// getToken 读取当前保存在微信本地缓存里的登录 token。
//
// 上传文件时要手动把 token 放到 header 里，所以单独暴露这个小函数。
function getToken() {
  return wx.getStorageSync(TOKEN_STORAGE_KEY)
}

// uploadFile 是对 wx.uploadFile 的封装。
//
// wx.uploadFile 不会走 wx.request，所以这里也要自己补 Authorization。
// 后端仍然返回统一的 { code, message, data }，这里会帮页面拆出 data。
async function uploadFile(options) {
  await ensureLogin()

  const token = getToken()

  return new Promise((resolve, reject) => {
    wx.uploadFile({
      url: `${API_BASE_URL}${options.url}`,
      filePath: options.filePath,
      name: options.name || 'file',
      formData: options.formData || {},
      header: {
        ...(token ? { Authorization: `Bearer ${token}` } : {})
      },
      success(res) {
        let body = {}

        try {
          body = JSON.parse(res.data || '{}')
        } catch (err) {
          reject({ message: '上传接口返回内容不是合法 JSON' })
          return
        }

        if (res.statusCode >= 200 && res.statusCode < 300 && body.code === 0) {
          resolve(body.data)
          return
        }

        reject({
          message: body.message || '上传失败',
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
  getToken,
  login,
  request,
  uploadFile
}
