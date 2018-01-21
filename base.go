package hst

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"
)

type base struct {
	s      *http.Server
	handle *http.ServeMux
	Addr   string
}

// HandleFunc ...
// Example:
//		HandleFunc("/", func(c *hst.Context){}, func(c *hst.Context){})
func (o *base) HandleFunc(pattern string, handler ...HandlerFunc) {
	o.handle.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		c := &Context{
			W:     w,
			R:     r,
			close: false,
		}
		for _, v := range handler {
			v(c)
			if c.close {
				break
			}
		}
	})
}

// Shutdown 优雅得关闭服务
func (o *base) Shutdown(waitTime time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), waitTime)
	defer cancel()
	o.s.Shutdown(ctx)
}

// Favicon 显示favicon.ico
func (o *base) Favicon() {
	o.handle.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		bs := []byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10, 0x02, 0x00, 0x01, 0x00, 0x01, 0x00, 0xb0, 0x00,
			0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x20, 0x00,
			0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00, 0x12, 0x0b,
			0x00, 0x00, 0x12, 0x0b, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x5d, 0x5d,
			0x5d, 0x00, 0xff, 0xff, 0xff, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xfb,
			0x00, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xe0, 0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0xff, 0xbf,
			0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0xfb, 0xff, 0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0x6f, 0xff,
			0x00, 0x00, 0x6f, 0xff, 0x00, 0x00, 0x6f, 0xff, 0x00, 0x00, 0x0f, 0xff, 0x00, 0x00, 0x6f, 0xff,
			0x00, 0x00, 0x6f, 0xff, 0x00, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xfb,
			0x00, 0x00, 0xff, 0xfb, 0x00, 0x00, 0xff, 0xe0, 0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0xff, 0xbf,
			0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0xfb, 0xff, 0x00, 0x00, 0xf8, 0x3f, 0x00, 0x00, 0x6f, 0xff,
			0x00, 0x00, 0x6f, 0xff, 0x00, 0x00, 0x6f, 0xff, 0x00, 0x00, 0x0f, 0xff, 0x00, 0x00, 0x6f, 0xff,
			0x00, 0x00, 0x6f, 0xff, 0x00, 0x00}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(bs)
	})
}

// Static 静态文件
func (o *base) Static(partten, path string) {
	o.handle.Handle(partten, http.StripPrefix(partten, http.FileServer(http.Dir(path))))
}

// HandlePfx 输出pfx证书给浏览器安装
// Example:
//		HandlePfx("/ssl.pfx", "/a/b/c.ssl.pfx"))
func (o *base) HandlePfx(partten, pfxPath string) {
	o.handle.HandleFunc(partten, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-x509-ca-cert")
		caCrt, err := ioutil.ReadFile(pfxPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(caCrt)
	})
}
