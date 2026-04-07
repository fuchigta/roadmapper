// Package server は roadmapper dev コマンド用の開発サーバを提供する。
// - dist/ を静的ファイルとして配信
// - HTML レスポンスにライブリロードスクリプトを注入
// - /sse エンドポイントで Server-Sent Events を配信
// - ソースファイル変更時にリビルドしてブラウザへ通知
package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const liveReloadScript = `
<script>
(function(){
  var src = new EventSource("/__sse");
  src.onmessage = function(e){
    if (e.data === "reload") { location.reload(); }
  };
  src.onerror = function(){
    setTimeout(function(){
      var s = new EventSource("/__sse");
      s.onmessage = src.onmessage;
    }, 2000);
  };
})();
</script>
`

// Server はファイル配信とライブリロードを担当する。
type Server struct {
	distDir string
	mu      sync.Mutex
	clients []chan struct{}
}

// New は新しい Server を返す。
func New(distDir string) *Server {
	return &Server{distDir: distDir}
}

// Notify は接続中のブラウザにリロードシグナルを送る。
func (s *Server) Notify() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Handler は HTTP ハンドラを返す。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/__sse", s.sseHandler)
	mux.HandleFunc("/", s.staticHandler)
	return mux
}

func (s *Server) sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.clients = append(s.clients, ch)
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		for i, c := range s.clients {
			if c == ch {
				s.clients = append(s.clients[:i], s.clients[i+1:]...)
				break
			}
		}
		s.mu.Unlock()
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 接続確認用の初回イベント
	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	// keepalive ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case <-ch:
			fmt.Fprintf(w, "data: reload\n\n")
			flusher.Flush()
		}
	}
}

func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	if urlPath == "/" || strings.HasSuffix(urlPath, "/") {
		urlPath += "index.html"
	}
	filePath := filepath.Join(s.distDir, filepath.FromSlash(urlPath))

	data, err := os.ReadFile(filePath)
	if err != nil {
		// ディレクトリ配下の index.html にフォールバック
		idxPath := filepath.Join(filePath, "index.html")
		if data2, err2 := os.ReadFile(idxPath); err2 == nil {
			data = data2
			filePath = idxPath
		} else {
			http.NotFound(w, r)
			return
		}
	}

	// HTML にライブリロードスクリプトを注入
	if strings.HasSuffix(filePath, ".html") {
		data = injectLiveReload(data)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", contentType(filePath))
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Write(data)
}

func injectLiveReload(html []byte) []byte {
	tag := []byte("</body>")
	if idx := bytes.Index(html, tag); idx >= 0 {
		return append(html[:idx], append([]byte(liveReloadScript), html[idx:]...)...)
	}
	return html
}

func contentType(path string) string {
	switch {
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".xml"):
		return "application/xml"
	case strings.HasSuffix(path, ".rss"):
		return "application/rss+xml"
	default:
		return "application/octet-stream"
	}
}

// Start は HTTP サーバを port で起動する (ブロッキング)。
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Dev server: http://localhost%s", addr)
	return http.ListenAndServe(addr, s.Handler())
}
