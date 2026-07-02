package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"project20260702/internal/auth"
	"project20260702/internal/config"
	"project20260702/internal/model"
	"project20260702/internal/response"
	"project20260702/internal/wechat"
)

// AuthHandler 处理登录相关接口。
type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
	wechat    config.WeChatConfig
}

// NewAuthHandler 创建登录接口处理器。
func NewAuthHandler(db *gorm.DB, jwtSecret string, wechatConfig config.WeChatConfig) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
		wechat:    wechatConfig,
	}
}

// wechatLoginRequest 是小程序登录接口的请求参数。
type wechatLoginRequest struct {
	// Code 是小程序 wx.login 返回的临时登录凭证。
	Code string `json:"code" binding:"required"`
}

// WeChatLogin 处理微信小程序登录。
//
// 对应接口：
// POST /api/auth/wechat-login
func (h *AuthHandler) WeChatLogin(c *gin.Context) {
	var req wechatLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "code 是必填参数")
		return
	}

	session, err := wechat.Code2Session(h.wechat, req.Code)
	if err != nil {
		response.BadRequest(c, "微信登录失败："+err.Error())
		return
	}

	var user model.User

	// FirstOrCreate 会先按 openid 查用户。
	// 如果用户不存在，就创建一条新用户记录。
	if err := h.db.Where("open_id = ?", session.OpenID).FirstOrCreate(&user, model.User{
		OpenID: session.OpenID,
	}).Error; err != nil {
		response.InternalError(c, err.Error())
		return
	}

	token, err := auth.GenerateToken(user.ID, h.jwtSecret)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id": user.ID,
		},
	})
}
