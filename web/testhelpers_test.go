package web_test

import (
	"io/fs"
	"strings"
	"sync"
	"testing"

	"github.com/fuchigta/roadmapper/web"
)

var (
	appJSOnce  sync.Once
	appJSCache string
	appJSErr   error
)

func readAppJS() (string, error) {
	appJSOnce.Do(func() {
		data, err := fs.ReadFile(web.FS, "static/app.js")
		if err != nil {
			appJSErr = err
			return
		}
		appJSCache = string(data)
	})
	return appJSCache, appJSErr
}

func extractJSSection(t *testing.T, beginMarker, endMarker string) string {
	t.Helper()
	s, err := readAppJS()
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	b := strings.Index(s, beginMarker)
	e := strings.Index(s, endMarker)
	if b < 0 || e < 0 || e <= b {
		t.Fatalf("markers %q / %q not found in app.js", beginMarker, endMarker)
	}
	return s[b : e+len(endMarker)]
}
