package main

import (
	"image2webp/internal/xwebp"
	"io"
	"log"
	"net/http"
	"strconv"
)

var (
	defaultConverter = xwebp.NewWebPConverter(80, false)
)

func main() {
	// 注册路由
	http.HandleFunc("/v1/upload", uploadHandler)
	http.HandleFunc("/v1/health", healthHandler)

	// 可选的静态文件服务（用于测试页面）
	http.Handle("/", http.FileServer(http.Dir("./front")))

	log.Printf("Starting WebP conversion server on :10080")
	log.Fatal(http.ListenAndServe(":10080", nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析查询参数（可选的质量设置）
	quality := getQualityParam(r)
	lossless := getLosslessParam(r)

	// 创建转换器实例（使用参数或默认值）
	converter := xwebp.NewWebPConverter(quality, lossless)

	// 解析multipart表单（最大32MB）
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 获取上传的文件
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 验证文件类型
	if !converter.IsImageSupported(header.Filename) {
		http.Error(w, "Unsupported image format. Supported: jpg, jpeg, png, bmp, tiff", http.StatusBadRequest)
		return
	}

	// 转换图片
	webpReader, newFilename, err := converter.ConvertToWebP(file, header.Filename)
	if err != nil {
		log.Printf("Conversion error: %v", err)
		http.Error(w, "Conversion failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+newFilename+"\"")
	w.Header().Set("X-Converted-Filename", newFilename)

	// 流式传输结果
	if _, err := io.Copy(w, webpReader); err != nil {
		log.Printf("Stream error: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "healthy", "service": "webp-converter"}`))
}

// 从查询参数获取质量设置
func getQualityParam(r *http.Request) float32 {
	qualityStr := r.URL.Query().Get("quality")
	if qualityStr == "" {
		return 80 // 默认值
	}

	quality, err := strconv.ParseFloat(qualityStr, 32)
	if err != nil || quality < 0 || quality > 100 {
		return 80 // 无效值时使用默认值
	}

	return float32(quality)
}

// 从查询参数获取无损压缩设置
func getLosslessParam(r *http.Request) bool {
	losslessStr := r.URL.Query().Get("lossless")
	return losslessStr == "true" || losslessStr == "1"
}
