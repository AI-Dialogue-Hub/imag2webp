package xwebp

import "net/http"

type WebpController struct {
	converter *WebPConverter
}

func NewWebpController() *WebpController {
	return &WebpController{
		converter: NewWebPConverter(80, false),
	}
}

func (ctl *WebpController) ConvertFromWeb(req *http.Request) {

}
