package web_test

import (
	"encoding/json"
	"testing"

	"github.com/dop251/goja"
)

// calcRoadmapProgress / isLocked (app.js の TEST:PROGRESS_BEGIN〜END) を Goja で検証する。
func TestCalcRoadmapProgress_drafts(t *testing.T) {
	section := extractJSSection(t, "// TEST:PROGRESS_BEGIN", "// TEST:PROGRESS_END")

	type node = map[string]any
	nodes := func(extra map[string]node) map[string]any {
		m := map[string]any{
			"__order": []string{"a", "b", "d", "o"},
			"a":       node{"type": "required"},
			"b":       node{"type": "required"},
			"d":       node{"type": "required", "draft": true},
			"o":       node{"type": "optional"},
		}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}
	allDone := map[string]any{"r": map[string]any{
		"a": map[string]string{"state": "done"},
		"b": map[string]string{"state": "none"},
		"d": map[string]string{"state": "done"}, // 既存の進捗は残るが計算では無視する
	}}

	tests := []struct {
		name     string
		cfg      map[string]any
		data     map[string]any
		progress map[string]any
		indexIDs []string // ROADMAP_NODE_IDS (インデックスページ)
		want     float64
	}{
		{"公開ビルド: 下書きは分母・分子から除外", map[string]any{}, nodes(nil), allDone, nil, 50},
		{"公開ビルド: 下書きのみ完了でも 0%", map[string]any{}, nodes(nil),
			map[string]any{"r": map[string]any{"d": map[string]string{"state": "done"}}}, nil, 0},
		{"プレビュー: 下書きも通常ノード", map[string]any{"draftPreview": true}, nodes(nil), allDone, nil, 67},
		{"下書きなし", map[string]any{}, map[string]any{
			"__order": []string{"a", "b"}, "a": node{"type": "required"}, "b": node{"type": "required"},
		}, allDone, nil, 50},
		{"インデックス: Go 側で除外済みの ID 一覧を使う", map[string]any{}, map[string]any{}, allDone, []string{"a", "b"}, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := goja.New()
			j := func(v any) string { b, _ := json.Marshal(v); return string(b) }
			pre := `var window = {}; var cfg = ` + j(tt.cfg) + `; var nodeData = ` + j(tt.data) +
				`; var progress = ` + j(tt.progress) + `;`
			if tt.indexIDs != nil {
				pre += `window.ROADMAP_NODE_IDS = {r: ` + j(tt.indexIDs) + `};`
			}
			if _, err := vm.RunString(pre + section); err != nil {
				t.Fatalf("load: %v", err)
			}
			v, err := vm.RunString(`calcRoadmapProgress('r')`)
			if err != nil {
				t.Fatal(err)
			}
			if got := v.ToFloat(); got != tt.want {
				t.Errorf("progress = %v, want %v", got, tt.want)
			}
		})
	}
}
