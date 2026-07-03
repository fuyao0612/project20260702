package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"project20260702/internal/middleware"
	"project20260702/internal/response"
)

const (
	// maxUploadImageSize 限制单张图片最大 5MB，避免用户误传大文件把服务器磁盘打满。
	maxUploadImageSize = 5 << 20

	// uploadImageRoot 是本地保存图片的根目录。
	// 这个目录只用于开发和学习阶段，正式项目可以换成对象存储。
	uploadImageRoot = "uploads/images"
)

// UploadHandler 处理文件上传相关接口。
type UploadHandler struct{}

// NewUploadHandler 创建上传处理器。
func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// Image 上传一张图片到后端本地目录。
//
// 对应接口：
// POST /api/uploads/images
func (h *UploadHandler) Image(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, 40101, "请先登录")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传图片文件")
		return
	}

	if fileHeader.Size <= 0 || fileHeader.Size > maxUploadImageSize {
		response.BadRequest(c, "图片大小不能超过 5MB")
		return
	}

	ext, mimeType, err := detectImageType(fileHeader)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userDir := filepath.Join(uploadImageRoot, strconv.FormatUint(userID, 10))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	fileName := fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), randomHex(6), ext)
	savePath := filepath.Join(userDir, fileName)

	if err := c.SaveUploadedFile(fileHeader, savePath); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"path":      filepath.ToSlash(savePath),
		"mime_type": mimeType,
	})
}

func detectImageType(fileHeader *multipart.FileHeader) (string, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return "", "", err
	}

	mimeType := http.DetectContentType(buffer[:n])
	switch mimeType {
	case "image/jpeg":
		return ".jpg", mimeType, nil
	case "image/png":
		return ".png", mimeType, nil
	case "image/webp":
		return ".webp", mimeType, nil
	default:
		return "", "", fmt.Errorf("只支持 jpg、png、webp 图片")
	}
}

func randomHex(byteCount int) string {
	bytes := make([]byte, byteCount)
	if _, err := rand.Read(bytes); err != nil {
		// 随机数失败非常少见，这里用时间兜底，避免上传流程直接中断。
		return strings.ReplaceAll(strconv.FormatInt(time.Now().UnixNano(), 16), "-", "")
	}

	return hex.EncodeToString(bytes)
}
