package xwebp

import (
	"fmt"
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
)

// WebPConverter WebP转换器结构体
type WebPConverter struct {
	Quality  float32 // WebP质量 (0-100)
	Lossless bool    // 是否使用无损压缩
}

// NewWebPConverter 创建WebP转换器实例
func NewWebPConverter(quality float32, lossless bool) *WebPConverter {
	if quality < 0 {
		quality = 80 // 默认质量
	}
	if quality > 100 {
		quality = 100
	}

	return &WebPConverter{
		Quality:  quality,
		Lossless: lossless,
	}
}

// ConvertToWebP 将图片转换为WebP格式
func (wc *WebPConverter) ConvertToWebP(src multipart.File, filename string) (io.Reader, string, error) {
	// 重置文件读取位置
	if seeker, ok := src.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// 先读取文件开头字节来验证格式
	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	// 重置读取位置
	if seeker, ok := src.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// 检查文件魔数（magic number）
	contentType := http.DetectContentType(buf[:n])
	log.Printf("Detected content type: %s", contentType)

	// 解码原始图片
	img, format, err := image.Decode(src)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image (detected: %s): %w", contentType, err)
	}

	log.Printf("Successfully decoded image format: %s", format)

	// 创建内存缓冲区存储WebP数据
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		// 编码为WebP
		err := webp.Encode(pw, img, &webp.Options{
			Lossless: wc.Lossless,
			Quality:  wc.Quality,
		})
		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to encode webp: %w", err))
			return
		}
	}()

	// 生成新的文件名
	newFilename := wc.generateWebpFilename(filename)

	return pr, newFilename, nil
}

// generateWebpFilename 生成WebP格式的文件名
func (wc *WebPConverter) generateWebpFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	baseName := strings.TrimSuffix(originalName, ext)
	return baseName + ".webp"
}

// IsImageSupported 检查图片格式是否支持转换
func (wc *WebPConverter) IsImageSupported(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supported := []string{".jpg", ".jpeg", ".png", ".bmp", ".tiff", ".tif"}

	for _, supportedExt := range supported {
		if ext == supportedExt {
			return true
		}
	}
	return false
}

// ConvertImageToWebP 通用工具方法：将图片转换为WebP格式
func ConvertImageToWebP(src multipart.File, filename string, quality float32, lossless bool) (io.Reader, string, error) {
	converter := NewWebPConverter(quality, lossless)
	return converter.ConvertToWebP(src, filename)
}

// ConvertLocalFileToWebP 如果你需要单独处理文件而不是multipart.File，可以使用这个方法
func ConvertLocalFileToWebP(inputPath, outputPath string, quality float32, lossless bool) error {
	// 打开输入文件
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// 解码图片
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// 创建输出文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// 编码为WebP
	err = webp.Encode(outFile, img, &webp.Options{
		Lossless: lossless,
		Quality:  quality,
	})
	if err != nil {
		return fmt.Errorf("failed to encode webp: %w", err)
	}

	return nil
}
