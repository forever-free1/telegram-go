package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/telegram-go/backend/internal/dto"
	"github.com/telegram-go/backend/internal/service"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	fileService *service.FileService
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(fileService *service.FileService) *UploadHandler {
	return &UploadHandler{fileService: fileService}
}

// UploadResponse 上传响应
type UploadResponse struct {
	Path      string `json:"path"`       // 相对路径，存入数据库
	URL       string `json:"url"`        // 完整访问URL
	Filename  string `json:"filename"`    // 原始文件名
	Size      int64  `json:"size"`       // 文件大小
	MediaType string `json:"media_type"` // 媒体类型
}

// Upload 上传文件
// POST /api/upload
func (h *UploadHandler) Upload(c *gin.Context) {
	// 从表单获取文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, "请选择要上传的文件"))
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedImageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}
	allowedAudioExts := []string{".mp3", ".wav", ".ogg", ".m4a", ".aac", ".flac"}
	allowedVideoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm"}

	isImage := contains(ext, allowedImageExts)
	isAudio := contains(ext, allowedAudioExts)
	isVideo := contains(ext, allowedVideoExts)

	if !isImage && !isAudio && !isVideo {
		c.JSON(http.StatusBadRequest, dto.Error(400, "不支持的文件类型，仅支持图片、音频、视频"))
		return
	}

	// 保存文件
	relativePath, accessURL, err := h.fileService.SaveFile(c.Request.Context(), file)
	if err != nil {
		switch err {
		case service.ErrInvalidFileType:
			c.JSON(http.StatusBadRequest, dto.Error(400, "无效的文件类型"))
		case service.ErrFileTooLarge:
			c.JSON(http.StatusBadRequest, dto.Error(400, "文件太大"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Error(500, "文件保存失败"))
		}
		return
	}

	// 确定媒体类型
	mediaType := "document"
	if isImage {
		mediaType = "image"
	} else if isAudio {
		mediaType = "audio"
	} else if isVideo {
		mediaType = "video"
	}

	response := UploadResponse{
		Path:      relativePath,
		URL:       accessURL,
		Filename:  file.Filename,
		Size:      file.Size,
		MediaType: mediaType,
	}

	c.JSON(http.StatusOK, dto.Success(response))
}

// contains 检查切片是否包含元素
func contains(s string, list []string) bool {
	for _, item := range list {
		if s == item {
			return true
		}
	}
	return false
}
