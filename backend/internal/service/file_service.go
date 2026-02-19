package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

var (
	ErrInvalidFileType = errors.New("invalid file type")
	ErrFileTooLarge   = errors.New("file too large")
	ErrSaveFailed     = errors.New("failed to save file")
)

// FileType 文件类型
type FileType int

const (
	FileTypeImage FileType = iota + 1
	FileTypeAudio
	FileTypeVideo
	FileTypeDocument
)

// FileService 文件服务
type FileService struct {
	uploadPath string
	baseURL    string
	maxSize    int64 // bytes
	logger     *zap.Logger
}

// FileServiceOption 文件服务配置选项
type FileServiceOption func(*FileService)

// WithMaxSize 设置最大文件大小
func WithMaxSize(maxSize int64) FileServiceOption {
	return func(s *FileService) {
		s.maxSize = maxSize
	}
}

// NewFileService 创建文件服务
func NewFileService(uploadPath, baseURL string, logger *zap.Logger, opts ...FileServiceOption) *FileService {
	s := &FileService{
		uploadPath: uploadPath,
		baseURL:    baseURL,
		maxSize:    100 * 1024 * 1024, // 默认 100MB
		logger:     logger,
	}

	for _, opt := range opts {
		opt(s)
	}

	// 确保上传目录存在
	if err := s.ensureUploadDir(); err != nil {
		logger.Error("failed to create upload directory", zap.Error(err))
	}

	return s
}

// ensureUploadDir 确保上传目录存在
func (s *FileService) ensureUploadDir() error {
	dirs := []string{
		s.uploadPath,
		filepath.Join(s.uploadPath, "images"),
		filepath.Join(s.uploadPath, "audio"),
		filepath.Join(s.uploadPath, "video"),
		filepath.Join(s.uploadPath, "documents"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// getFileTypeByExtension 根据文件扩展名获取文件类型
func (s *FileService) getFileTypeByExtension(filename string) (FileType, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}
	audioExts := []string{".mp3", ".wav", ".ogg", ".m4a", ".aac", ".flac"}
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm"}
	documentExts := []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt"}

	for _, e := range imageExts {
		if ext == e {
			return FileTypeImage, nil
		}
	}

	for _, e := range audioExts {
		if ext == e {
			return FileTypeAudio, nil
		}
	}

	for _, e := range videoExts {
		if ext == e {
			return FileTypeVideo, nil
		}
	}

	for _, e := range documentExts {
		if ext == e {
			return FileTypeDocument, nil
		}
	}

	return 0, ErrInvalidFileType
}

// detectMIMEType 检测文件的MIME类型
// 读取文件前512字节来判断真实的MIME类型
func (s *FileService) detectMIMEType(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 读取前512字节
	header := make([]byte, 512)
	n, err := src.Read(header)
	if err != nil && err != io.EOF {
		return "", err
	}
	header = header[:n]

	// 使用 http.DetectContentType 检测MIME类型
	contentType := http.DetectContentType(header)
	return contentType, nil
}

// validateMIMEType 验证文件MIME类型是否与扩展名匹配
func (s *FileService) validateMIMEType(filename string, contentType string) error {
	ext := strings.ToLower(filepath.Ext(filename))

	// 允许的MIME类型映射
	allowedTypes := map[string][]string{
		"image/jpeg":               {".jpg", ".jpeg"},
		"image/png":                {".png"},
		"image/gif":                {".gif"},
		"image/webp":               {".webp"},
		"image/bmp":                {".bmp"},
		"audio/mpeg":               {".mp3"},
		"audio/wav":                {".wav"},
		"audio/ogg":                {".ogg"},
		"audio/mp4":                {".m4a", ".aac"},
		"audio/x-flac":             {".flac"},
		"video/mp4":                {".mp4"},
		"video/x-msvideo":          {".avi"},
		"video/quicktime":          {".mov"},
		"video/x-ms-wmv":           {".wmv"},
		"video/x-flv":              {".flv"},
		"video/x-matroska":         {".mkv"},
		"video/webm":               {".webm"},
		"application/pdf":           {".pdf"},
		"application/msword":       {".doc"},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {".docx"},
		"application/vnd.ms-excel": {".xls"},
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":      {".xlsx"},
		"application/vnd.ms-powerpoint": {".ppt"},
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": {".pptx"},
		"text/plain":               {".txt"},
	}

	allowedExts, ok := allowedTypes[contentType]
	if !ok {
		return fmt.Errorf("unsupported MIME type: %s", contentType)
	}

	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return nil
		}
	}

	return fmt.Errorf("MIME type %s does not match extension %s", contentType, ext)
}

// getSubDir 获取子目录
func (s *FileService) getSubDir(fileType FileType) string {
	switch fileType {
	case FileTypeImage:
		return "images"
	case FileTypeAudio:
		return "audio"
	case FileTypeVideo:
		return "video"
	case FileTypeDocument:
		return "documents"
	default:
		return "documents"
	}
}

// generateFilename 生成唯一文件名
func (s *FileService) generateFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	now := time.Now()
	return fmt.Sprintf("%d%d%d%d%d%d_%s%s",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Nanosecond(),
		RandomString(8), ext)
}

// RandomString 生成随机字符串
func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}

// SaveFile 保存文件
// 返回: 文件相对路径、完整访问URL、错误
func (s *FileService) SaveFile(ctx context.Context, file *multipart.FileHeader) (string, string, error) {
	// 验证文件类型（基于扩展名）
	fileType, err := s.getFileTypeByExtension(file.Filename)
	if err != nil {
		s.logger.Warn("invalid file type", zap.String("filename", file.Filename))
		return "", "", ErrInvalidFileType
	}

	// 检测并验证MIME类型
	mimeType, err := s.detectMIMEType(file)
	if err != nil {
		s.logger.Warn("failed to detect MIME type", zap.String("filename", file.Filename), zap.Error(err))
		// 如果检测失败，至少验证扩展名
	} else {
		// 验证MIME类型与扩展名是否匹配
		if err := s.validateMIMEType(file.Filename, mimeType); err != nil {
			s.logger.Warn("MIME type validation failed", zap.String("filename", file.Filename),
				zap.String("mime_type", mimeType), zap.Error(err))
			return "", "", ErrInvalidFileType
		}
		s.logger.Debug("MIME type validated", zap.String("filename", file.Filename),
			zap.String("mime_type", mimeType))
	}

	// 验证文件大小
	if s.maxSize > 0 && file.Size > s.maxSize {
		s.logger.Warn("file too large", zap.Int64("size", file.Size), zap.Int64("max", s.maxSize))
		return "", "", ErrFileTooLarge
	}

	// 生成唯一文件名
	filename := s.generateFilename(file.Filename)
	subDir := s.getSubDir(fileType)
	relativePath := filepath.Join(subDir, filename)
	absolutePath := filepath.Join(s.uploadPath, relativePath)

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		s.logger.Error("failed to open uploaded file", zap.Error(err))
		return "", "", ErrSaveFailed
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(absolutePath)
	if err != nil {
		s.logger.Error("failed to create destination file", zap.Error(err))
		return "", "", ErrSaveFailed
	}
	defer dst.Close()

	// 复制内容
	if _, err := io.Copy(dst, src); err != nil {
		s.logger.Error("failed to copy file content", zap.Error(err))
		return "", "", ErrSaveFailed
	}

	// 返回相对路径和完整URL
	relativePath = strings.ReplaceAll(relativePath, "\\", "/")
	accessURL := fmt.Sprintf("%s/%s", s.baseURL, relativePath)

	s.logger.Info("file saved successfully",
		zap.String("filename", filename),
		zap.String("path", relativePath),
		zap.Int64("size", file.Size))

	return relativePath, accessURL, nil
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(ctx context.Context, relativePath string) error {
	absolutePath := filepath.Join(s.uploadPath, relativePath)
	if err := os.Remove(absolutePath); err != nil {
		s.logger.Error("failed to delete file", zap.Error(err))
		return err
	}
	return nil
}

// GetFilePath 获取文件的完整路径
func (s *FileService) GetFilePath(relativePath string) string {
	return filepath.Join(s.uploadPath, relativePath)
}

// GetFileURL 获取文件的访问URL
func (s *FileService) GetFileURL(relativePath string) string {
	relativePath = strings.ReplaceAll(relativePath, "\\", "/")
	return fmt.Sprintf("%s/%s", s.baseURL, relativePath)
}
