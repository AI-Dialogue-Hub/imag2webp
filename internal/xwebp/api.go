package xwebp

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"  // 注册GIF解码器
	_ "image/jpeg" // 注册JPEG解码器
	_ "image/png"  // 注册PNG解码器

	"github.com/chai2010/webp"
	"golang.org/x/image/bmp"  // BMP支持
	"golang.org/x/image/tiff" // TIFF支持
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
func (wc *WebPConverter) ConvertToWebP(src multipart.File, filename string) (io.Reader, string, error) {
	// 重置文件读取位置
	if seeker, ok := src.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// 先读取文件内容到字节切片，便于多次尝试解码
	fileData, err := io.ReadAll(src)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	// 检测内容类型
	contentType := http.DetectContentType(fileData)
	log.Printf("Detected content type: %s, file size: %d bytes", contentType, len(fileData))

	var img image.Image
	var decodeErr error
	var format string

	// 根据内容类型尝试不同的解码策略
	switch {
	case strings.Contains(contentType, "png"):
		img, format, decodeErr = image.Decode(bytes.NewReader(fileData))
	case strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg"):
		img, format, decodeErr = image.Decode(bytes.NewReader(fileData))
	case strings.Contains(contentType, "gif"):
		img, format, decodeErr = image.Decode(bytes.NewReader(fileData))
	case strings.Contains(contentType, "bmp"):
		img, decodeErr = bmp.Decode(bytes.NewReader(fileData))
		format = "bmp"
	case strings.Contains(contentType, "tiff"):
		img, decodeErr = tiff.Decode(bytes.NewReader(fileData))
		format = "tiff"
	default:
		decodeErr = fmt.Errorf("unsupported content type: %s", contentType)
	}

	if decodeErr != nil {
		// 如果特定解码失败，尝试通用解码
		img, format, decodeErr = image.Decode(bytes.NewReader(fileData))
		if decodeErr != nil {
			return nil, "", fmt.Errorf("failed to decode image (type: %s): %w", contentType, decodeErr)
		}
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

// decodePNG 专门处理PNG解码，可以添加特殊处理逻辑
func (wc *WebPConverter) decodePNG(data []byte) (image.Image, error) {
	// 首先尝试标准解码
	img, format, decodeErr := image.Decode(bytes.NewReader(data))
	log.Printf("format:%v", format)
	if decodeErr == nil {
		return img, nil
	}

	// 如果标准解码失败，可以在这里添加特殊处理逻辑
	// 例如处理非标准PNG、带透明通道的PNG等

	log.Printf("Standard PNG decode failed: %v, attempting alternative approaches", decodeErr)

	// 作为备选，可以尝试其他PNG处理库
	// 这里暂时返回错误，你可以根据需要扩展

	return nil, decodeErr
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
