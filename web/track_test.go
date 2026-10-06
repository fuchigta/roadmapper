package web_test

import (
	"encoding/json"
	"testing"

	"github.com/dop251/goja"
)

// runTrackJS は app.js の Analytics セクションを Goja で実行し、
// roadmapperTrack に渡された呼び出しを JSON 配列で返す。
func runTrackJS(t *testing.T, search string, analyticsEvents bool, throwing bool, script string) []map[string]any {
	t.Helper()
	section := extractJSSection(t, "// TEST:TRACK_BEGIN", "// TEST:TRACK_END")

	vm := goja.New()
	vm.Set("cfg", map[string]any{"analyticsEvents": analyticsEvents})
	vm.Set("roadmapId", "rm1")
	vm.Set("throwing", throwing)
	// URLSearchParams の最小スタブ (?p=... の有無のみ判定)
	if _, err := vm.RunString(`
		var location = { search: ` + jsString(t, search) + ` };
		var URLSearchParams = function (s) { this.has = function (k) { return s.indexOf(k + '=') >= 0; }; };
		var calls = [];
		var window = {};
		window.roadmapperTrack = function (n, d) {
			calls.push({ name: n, data: d });
			if (throwing) throw new Error('boom');
		};
	`); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := vm.RunString(section + "\n" + script); err != nil {
		t.Fatalf("run: %v", err)
	}
	v, err := vm.RunString("JSON.stringify(calls)")
	if err != nil {
		t.Fatal(err)
	}
	var calls []map[string]any
	if err := json.Unmarshal([]byte(v.String()), &calls); err != nil {
		t.Fatal(err)
	}
	return calls
}

func jsString(t *testing.T, s string) string {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestTrack(t *testing.T) {
	tests := []struct {
		name      string
		search    string
		events    bool
		throwing  bool
		script    string
		wantCalls int
	}{
		{"events 有効で送信", "", true, false, `trackState('a', 'done')`, 1},
		{"analyticsEvents=false では呼ばれない", "", false, false, `track('node_open', {}); trackState('a', 'done')`, 0},
		{"シェアビューでは node_state を送らない", "?p=abc", true, false, `trackState('a', 'done')`, 0},
		{"シェアビューでも他イベントは送る", "?p=abc", true, false, `track('node_open', {})`, 1},
		{"アダプタの例外を外に出さない", "", true, true, `track('node_open', {}); trackState('a', 'done')`, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := runTrackJS(t, tt.search, tt.events, tt.throwing, tt.script)
			if len(calls) != tt.wantCalls {
				t.Fatalf("calls = %d, want %d: %v", len(calls), tt.wantCalls, calls)
			}
		})
	}
}

func TestTrackStatePayload(t *testing.T) {
	calls := runTrackJS(t, "", true, false, `trackState('n1', 'done')`)
	if len(calls) != 1 {
		t.Fatalf("calls = %d", len(calls))
	}
	data, _ := calls[0]["data"].(map[string]any)
	if calls[0]["name"] != "node_state" || data["roadmap"] != "rm1" || data["node"] != "n1" || data["state"] != "done" || len(data) != 3 {
		t.Errorf("unexpected payload: %v", calls[0])
	}
}
