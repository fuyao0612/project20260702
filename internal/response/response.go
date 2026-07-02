package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// CodeOK 表示请求成功。
	CodeOK = 0

	// CodeBadRequest 表示请求参数不正确。
	CodeBadRequest = 40001

	// CodeNotFound 表示请求的数据不存在。
	CodeNotFound = 40401

	// CodeInternalError 表示服务器内部错误。
	CodeInternalError = 50001
)

// Body 是所有接口统一返回给前端的 JSON 结构。
//
// 成功时：
// {"code":0,"message":"ok","data":...}
//
// 失败时：
// {"code":40001,"message":"请求参数不正确","data":null}
type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 返回成功响应。
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code:    CodeOK,
		Message: "ok",
		Data:    data,
	})
}

// Created 返回资源创建成功响应。
//
// 新增账单这类接口通常使用 HTTP 201。
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Body{
		Code:    CodeOK,
		Message: "ok",
		Data:    data,
	})
}

// Error 返回失败响应。
//
// httpStatus 是 HTTP 状态码，例如 400、404、500。
// code 是业务错误码，方便前端或后续日志排查。
func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Body{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// BadRequest 返回参数错误。
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, CodeBadRequest, message)
}

// NotFound 返回数据不存在。
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, CodeNotFound, message)
}

// InternalError 返回服务器内部错误。
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, CodeInternalError, message)
}
