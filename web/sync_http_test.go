package web_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dop251/goja"
)

func loadHTTPSyncJS(t *testing.T) string {
	return extractJSSection(t, "// TEST:HTTP_BEGIN", "// TEST:HTTP_END")
}

// ===== httptest server =====

type capturedReq struct {
	method string
	path   string
	body   string
}

// newSyncServer starts an httptest.Server that applies handler per URL path.
// handler receives (w, r) and should write the response.
// All requests are also recorded in the returned slice (pointer).
func newSyncServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *[]capturedReq) {
	t.Helper()
	var reqs []capturedReq
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		reqs = append(reqs, capturedReq{r.Method, r.URL.Path, string(body)})
		r.Body = io.NopCloser(bytes.NewReader(body))
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &reqs
}

// ===== Goja VM setup =====

// setupSyncVM creates a Goja runtime wired to the given server URL.
// `initialProgress` is the JS object to set as `progress` (may be nil for empty).
func setupSyncVM(t *testing.T, serverURL string, initialProgress map[string]interface{}) *goja.Runtime {
	t.Helper()
	vm := goja.New()

	// cfg: provides `sync` configuration
	vm.Set("cfg", map[string]interface{}{
		"progressSync": map[string]interface{}{
			"enabled":  true,
			"endpoint": serverURL,
		},
	})

	// progress: in-memory progress store
	if initialProgress == nil {
		initialProgress = make(map[string]interface{})
	}
	vm.Set("progress", vm.ToValue(initialProgress))

	// localStorage: data stored as JS own properties; methods are non-enumerable
	// so Object.keys(localStorage) returns only stored keys.
	if _, err := vm.RunString(`
		var localStorage = {};
		Object.defineProperty(localStorage, 'getItem', {
			enumerable: false,
			value: function(k) { return localStorage[k] !== undefined ? String(localStorage[k]) : null; }
		});
		Object.defineProperty(localStorage, 'setItem', {
			enumerable: false,
			value: function(k, v) { localStorage[k] = String(v); }
		});
		Object.defineProperty(localStorage, 'removeItem', {
			enumerable: false,
			value: function(k) { delete localStorage[k]; }
		});
	`); err != nil {
		t.Fatalf("setup localStorage: %v", err)
	}

	// crypto.randomUUID: returns a fixed device ID for predictable URL assertions
	cryptoObj := vm.NewObject()
	cryptoObj.Set("randomUUID", func(goja.FunctionCall) goja.Value {
		return vm.ToValue("test-device-uuid")
	})
	vm.Set("crypto", cryptoObj)

	// console.warn: suppress [roadmapper] warnings from flushPut failure path
	consoleObj := vm.NewObject()
	consoleObj.Set("warn", func(goja.FunctionCall) goja.Value { return goja.Undefined() })
	vm.Set("console", consoleObj)

	// setTimeout: call the callback immediately (no 800ms wait in tests)
	vm.Set("setTimeout", func(call goja.FunctionCall) goja.Value {
		if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
			fn(goja.Undefined()) //nolint:errcheck
		}
		return vm.ToValue(0)
	})
	vm.Set("clearTimeout", func(goja.FunctionCall) goja.Value { return goja.Undefined() })

	// fetch: real HTTP client → httptest.Server
	client := &http.Client{}
	vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()
		method := "GET"
		var reqBody string
		if !goja.IsUndefined(call.Argument(1)) {
			if opts, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				if m, ok := opts["method"].(string); ok {
					method = m
				}
				if b, ok := opts["body"].(string); ok {
					reqBody = b
				}
			}
		}

		var bodyRdr io.Reader
		if reqBody != "" {
			bodyRdr = strings.NewReader(reqBody)
		}
		req, err := http.NewRequest(method, url, bodyRdr)
		if err != nil {
			promise, _, reject := vm.NewPromise()
			reject(vm.ToValue(err.Error()))
			return vm.ToValue(promise)
		}

		resp, err := client.Do(req)
		promise, resolve, reject := vm.NewPromise()
		if err != nil {
			reject(vm.ToValue(err.Error()))
			return vm.ToValue(promise)
		}
		defer resp.Body.Close()
		respBodyBytes, _ := io.ReadAll(resp.Body)
		respBodyStr := string(respBodyBytes)

		// Build JS response object
		resObj := vm.NewObject()
		resObj.Set("status", resp.StatusCode)
		resObj.Set("ok", resp.StatusCode >= 200 && resp.StatusCode < 300)
		resObj.Set("json", func(goja.FunctionCall) goja.Value {
			p, res, rej := vm.NewPromise()
			var data interface{}
			if err := json.Unmarshal([]byte(respBodyStr), &data); err != nil {
				rej(vm.ToValue(err.Error()))
			} else {
				res(vm.ToValue(data))
			}
			return vm.ToValue(p)
		})
		resolve(resObj)
		return vm.ToValue(promise)
	})

	// Load the sync JS section extracted from app.js
	if _, err := vm.RunString(loadHTTPSyncJS(t)); err != nil {
		t.Fatalf("load sync JS: %v", err)
	}
	return vm
}

// runAsync runs a JS expression that returns a Promise and waits for it to settle.
// Returns the resolved value (or calls t.Fatal on rejection / timeout).
func runAsync(t *testing.T, vm *goja.Runtime, expr string) goja.Value {
	t.Helper()
	var result goja.Value
	var settled bool
	vm.Set("__testResolve", func(call goja.FunctionCall) goja.Value {
		result = call.Argument(0)
		settled = true
		return goja.Undefined()
	})
	vm.Set("__testReject", func(call goja.FunctionCall) goja.Value {
		t.Errorf("async expression rejected: %v", call.Argument(0))
		settled = true
		return goja.Undefined()
	})
	if _, err := vm.RunString(`Promise.resolve(` + expr + `).then(__testResolve, __testReject)`); err != nil {
		t.Fatalf("RunString(%q): %v", expr, err)
	}
	// With synchronous fetch, all microtasks should flush during RunString.
	// Try a few extra no-op runs just in case.
	for i := 0; !settled && i < 20; i++ {
		vm.RunString(`(function(){})()`) //nolint:errcheck
	}
	if !settled {
		t.Fatalf("async expression %q did not settle", expr)
	}
	return result
}

// getLSKey reads a localStorage key from the Goja VM.
func getLSKey(t *testing.T, vm *goja.Runtime, key string) string {
	t.Helper()
	lsVal := vm.Get("localStorage")
	v := lsVal.ToObject(vm).Get(key)
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return ""
	}
	return v.String()
}

// setLSKey sets a localStorage key in the Goja VM.
func setLSKey(t *testing.T, vm *goja.Runtime, key, value string) {
	t.Helper()
	k, _ := json.Marshal(key)
	v, _ := json.Marshal(value)
	if _, err := vm.RunString(`localStorage.setItem(` + string(k) + `, ` + string(v) + `)`); err != nil {
		t.Fatalf("setLSKey(%q): %v", key, err)
	}
}

// ===== Tests =====

func TestSyncHTTP_fetchRemote_200(t *testing.T) {
	remoteProgress := map[string]interface{}{
		"node1": map[string]interface{}{"state": "done", "tasks": []interface{}{true, false}},
	}
	srv, reqs := newSyncServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(remoteProgress) //nolint:errcheck
	})

	vm := setupSyncVM(t, srv.URL, nil)
	result := runAsync(t, vm, `fetchRemote('roadmap1')`)

	// Verify GET request
	if len(*reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*reqs))
	}
	req := (*reqs)[0]
	if req.method != "GET" {
		t.Errorf("method = %q, want GET", req.method)
	}
	wantPath := "/test-device-uuid/roadmap1"
	if req.path != wantPath {
		t.Errorf("path = %q, want %q", req.path, wantPath)
	}

	// Verify returned data
	exported := result.Export()
	m, ok := exported.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not an object: %T", exported)
	}
	node1, ok := m["node1"].(map[string]interface{})
	if !ok {
		t.Fatalf("result.node1 missing or wrong type")
	}
	if node1["state"] != "done" {
		t.Errorf("node1.state = %v, want done", node1["state"])
	}
}

func TestSyncHTTP_fetchRemote_404(t *testing.T) {
	srv, reqs := newSyncServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	vm := setupSyncVM(t, srv.URL, nil)
	result := runAsync(t, vm, `fetchRemote('roadmap1')`)

	if len(*reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*reqs))
	}
	// fetchRemote returns {} on 404
	exported := result.Export()
	m, ok := exported.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not an object: %T %v", exported, result)
	}
	if len(m) != 0 {
		t.Errorf("expected empty object on 404, got %v", m)
	}
}

func TestSyncHTTP_fetchRemote_500(t *testing.T) {
	srv, reqs := newSyncServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	vm := setupSyncVM(t, srv.URL, nil)
	result := runAsync(t, vm, `fetchRemote('roadmap1')`)

	if len(*reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*reqs))
	}
	// fetchRemote returns null on non-ok, non-404 status
	if !goja.IsNull(result) {
		t.Errorf("expected null on 500, got %v", result)
	}
}

func TestSyncHTTP_flushPut_success(t *testing.T) {
	srv, reqs := newSyncServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})

	initialProgress := map[string]interface{}{
		"roadmap1": map[string]interface{}{
			"node1": map[string]interface{}{"state": "done", "tasks": []interface{}{}},
		},
	}
	vm := setupSyncVM(t, srv.URL, initialProgress)

	// dirty flag is set at start of flushPut
	runAsync(t, vm, `flushPut('roadmap1')`)

	if len(*reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*reqs))
	}
	req := (*reqs)[0]
	if req.method != "PUT" {
		t.Errorf("method = %q, want PUT", req.method)
	}
	wantPath := "/test-device-uuid/roadmap1"
	if req.path != wantPath {
		t.Errorf("path = %q, want %q", req.path, wantPath)
	}

	// Verify body is JSON of progress[roadmap1]
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(req.body), &body); err != nil {
		t.Fatalf("PUT body is not JSON: %v", err)
	}
	if _, ok := body["node1"]; !ok {
		t.Errorf("PUT body missing node1: %v", body)
	}

	// Dirty flag must be cleared on success
	if dirty := getLSKey(t, vm, "roadmapper:sync-dirty:roadmap1"); dirty != "" {
		t.Errorf("dirty flag should be cleared on PUT success, got %q", dirty)
	}
}

func TestSyncHTTP_flushPut_failure(t *testing.T) {
	srv, _ := newSyncServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	initialProgress := map[string]interface{}{
		"roadmap1": map[string]interface{}{
			"node1": map[string]interface{}{"state": "in-progress", "tasks": []interface{}{}},
		},
	}
	vm := setupSyncVM(t, srv.URL, initialProgress)

	runAsync(t, vm, `flushPut('roadmap1')`)

	// Dirty flag must remain when PUT fails
	if dirty := getLSKey(t, vm, "roadmapper:sync-dirty:roadmap1"); dirty == "" {
		t.Errorf("dirty flag should remain after PUT failure")
	}
}

func TestSyncHTTP_retryDirty(t *testing.T) {
	var putCount int
	srv, reqs := newSyncServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			putCount++
		}
		w.WriteHeader(204)
	})

	initialProgress := map[string]interface{}{
		"rm1": map[string]interface{}{
			"node1": map[string]interface{}{"state": "done", "tasks": []interface{}{}},
		},
		"rm2": map[string]interface{}{
			"node2": map[string]interface{}{"state": "in-progress", "tasks": []interface{}{}},
		},
	}
	vm := setupSyncVM(t, srv.URL, initialProgress)

	// Simulate two dirty roadmaps in localStorage
	setLSKey(t, vm, "roadmapper:sync-dirty:rm1", "1")
	setLSKey(t, vm, "roadmapper:sync-dirty:rm2", "1")

	runAsync(t, vm, `retryDirty()`)

	// Both dirty roadmaps should have triggered a PUT
	if putCount != 2 {
		t.Errorf("expected 2 PUTs from retryDirty, got %d", putCount)
	}

	// Verify paths include both roadmap IDs
	paths := make(map[string]bool)
	for _, r := range *reqs {
		if r.method == "PUT" {
			paths[r.path] = true
		}
	}
	for _, rmId := range []string{"rm1", "rm2"} {
		want := "/test-device-uuid/" + rmId
		if !paths[want] {
			t.Errorf("expected PUT to %q, got paths: %v", want, paths)
		}
	}

	// Dirty flags should be cleared
	for _, rmId := range []string{"rm1", "rm2"} {
		if dirty := getLSKey(t, vm, "roadmapper:sync-dirty:"+rmId); dirty != "" {
			t.Errorf("dirty flag for %q should be cleared, got %q", rmId, dirty)
		}
	}
}

func TestSyncHTTP_deviceIdPersists(t *testing.T) {
	var paths []string
	srv, _ := newSyncServer(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(404)
	})

	vm := setupSyncVM(t, srv.URL, nil)

	// First call generates and stores device ID
	runAsync(t, vm, `fetchRemote('rm1')`)
	// Second call should reuse the same device ID
	runAsync(t, vm, `fetchRemote('rm2')`)

	if len(paths) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(paths))
	}
	// Both paths should start with the same device ID segment
	seg0 := strings.Split(strings.TrimPrefix(paths[0], "/"), "/")[0]
	seg1 := strings.Split(strings.TrimPrefix(paths[1], "/"), "/")[0]
	if seg0 == "" || seg0 != seg1 {
		t.Errorf("device ID should persist across calls: %q vs %q", seg0, seg1)
	}
}
