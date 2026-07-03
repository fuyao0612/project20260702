# project20260702 记账小程序

这是一个用于学习和实践的个人记账项目。当前目标是做出一个可以在微信开发者工具中运行的小程序，并配套一个使用 Go 编写的后端服务。

项目目前已经实现：

- 微信小程序端：账单列表、新增账单、编辑账单、删除账单、分类选择、月度统计页
- Go 后端：微信登录、JWT 鉴权、账单 CRUD、分类接口、月度统计接口
- MySQL：用户、账单、分类数据持久化
- Docker：本地 MySQL 开发环境

后续计划会继续加入 AI Agent 能力，例如识图记账、文本记账草稿、账本总结等。

## 技术栈

后端：

- Go
- Gin
- GORM
- MySQL
- JWT
- Docker Compose

小程序：

- 原生微信小程序
- WXML / WXSS / JavaScript
- `wx.request`
- `wx.login`

数据库：

- MySQL 8.4

## 项目结构

```text
project20260702
├─ cmd
│  └─ server                 # Go 后端启动入口
├─ infra
│  └─ docker-compose.mysql.yml # 本地 MySQL Docker 配置
├─ internal
│  ├─ auth                   # JWT token 生成与解析
│  ├─ config                 # 环境变量配置读取
│  ├─ database               # MySQL 连接与默认数据初始化
│  ├─ handler                # Gin 接口处理函数
│  ├─ middleware             # 登录鉴权中间件
│  ├─ model                  # GORM 数据模型
│  ├─ response               # 统一响应格式
│  ├─ router                 # 路由注册
│  └─ wechat                 # 微信 code2Session 登录封装
├─ miniprogram               # 微信小程序前端
├─ .env.example              # 环境变量模板
├─ .gitignore
├─ go.mod
├─ go.sum
└─ README.md
```

## 本地开发环境

需要提前准备：

- Go
- Docker Desktop
- 微信开发者工具
- DBeaver，可选，用于查看 MySQL

## 启动 MySQL

在项目根目录执行：

```powershell
docker compose -f .\infra\docker-compose.mysql.yml up -d
```

查看容器：

```powershell
docker ps
```

当前 MySQL 默认连接信息：

```text
Host: 127.0.0.1
Port: 3306
Database: project20260702
Username: root
Password: 见本地 .env
```

停止 MySQL：

```powershell
docker compose -f .\infra\docker-compose.mysql.yml down
```

## 配置后端

复制 `.env.example` 为 `.env`，并填写本地真实配置。

```env
APP_ENV=local
HTTP_ADDR=:8080

MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_DATABASE=project20260702
MYSQL_USERNAME=root
MYSQL_PASSWORD=your_mysql_password

JWT_SECRET=change_me_to_a_long_random_string

WECHAT_APP_ID=your_wechat_miniprogram_appid
WECHAT_APP_SECRET=your_wechat_miniprogram_appsecret
WECHAT_DEV_OPENID=dev_openid_for_local_testing
```

说明：

- `.env` 是本地真实配置，不提交到 Git。
- `.env.example` 是配置模板，可以提交。
- 本地没有真实微信 `AppSecret` 时，后端会使用 `WECHAT_DEV_OPENID` 模拟登录，方便开发调试。

## 启动后端

在项目根目录执行：

```powershell
go run ./cmd/server
```

健康检查：

```text
http://127.0.0.1:8080/ping
```

后端启动时会自动：

- 连接 MySQL
- AutoMigrate 同步表结构
- 初始化默认分类

## 运行微信小程序

使用微信开发者工具导入：

```text
E:\project20260702\miniprogram
```

本地开发时，小程序请求地址在：

```text
miniprogram/utils/api.js
```

当前默认：

```js
const API_BASE_URL = 'http://127.0.0.1:8080'
```

如果是本地调试，需要在微信开发者工具中勾选“不校验合法域名、web-view、TLS 版本以及 HTTPS 证书”。

部署到服务器后，需要把 `API_BASE_URL` 改成 HTTPS 域名，例如：

```js
const API_BASE_URL = 'https://api.example.com'
```

## 已有接口

统一响应格式：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

登录：

```text
POST /api/auth/wechat-login
```

分类：

```text
GET /api/categories?type=expense
GET /api/categories?type=income
```

账单：

```text
GET    /api/transactions
GET    /api/transactions?month=2026-07
GET    /api/transactions/:id
POST   /api/transactions
PUT    /api/transactions/:id
DELETE /api/transactions/:id
```

统计：

```text
GET /api/statistics/monthly?month=2026-07
```

除登录接口外，业务接口都需要请求头：

```text
Authorization: Bearer <token>
```

## 当前小程序页面

```text
pages/index       首页：月度概览、账单列表
pages/create      新增账单
pages/detail      编辑/删除账单
pages/statistics  月度统计、分类支出排行
```

## 默认分类

支出分类：

```text
餐饮、交通、购物、学习、娱乐、医疗、住房、其他
```

收入分类：

```text
工资、兼职、红包、理财、其他
```

## 后续计划

- 完善真实微信小程序 AppID / AppSecret 配置
- 部署 Go 后端到云服务器
- 配置 Nginx 和 HTTPS
- 接入微信小程序体验版
- 增加 AI 文本记账草稿
- 增加 AI 识图记账
- 增加 AI 月度账本总结
- 增加更完整的统计图表

## 注意事项

- 不要提交 `.env`。
- 不要提交微信 `AppSecret`、AI API Key、数据库真实密码。
- 金额在后端和数据库中统一使用“分”为单位，例如 `18.50 元 = 1850`。
- 当前项目处于学习开发阶段，数据库迁移暂时使用 GORM `AutoMigrate`。
