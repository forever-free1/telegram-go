package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
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
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

// SaveFile 保存文件
// 返回: 文件相对路径、完整访问URL、错误
func (s *FileService) SaveFile(ctx context.Context, file *multipart.FileHeader) (string, string, error) {
	// 验证文件类型
	fileType, err := s.getFileTypeByExtension(file.Filename)
	if err != nil {
		s.logger.Warn("invalid file type", zap.String("filename", file.Filename))
		return "", "", ErrInvalidFileType
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
