package web_test

import (
	"encoding/json"
	"testing"

	"github.com/dop251/goja"
)

type nodeProgress struct {
	State string `json:"state"`
	Tasks []bool `json:"tasks"`
}

func loadMergeJS(t *testing.T) string {
	return extractJSSection(t, "// TEST:MERGE_BEGIN", "// TEST:MERGE_END")
}

func runMerge(t *testing.T, vm *goja.Runtime, local, remote map[string]nodeProgress) map[string]nodeProgress {
	t.Helper()
	localJSON, _ := json.Marshal(local)
	remoteJSON, _ := json.Marshal(remote)
	val, err := vm.RunString(`mergeRoadmap(` + string(localJSON) + `, ` + string(remoteJSON) + `)`)
	if err != nil {
		t.Fatalf("mergeRoadmap call: %v", err)
	}
	raw, err := json.Marshal(val.Export())
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	var result map[string]nodeProgress
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	return result
}

func TestMergeRoadmap(t *testing.T) {
	vm := goja.New()
	if _, err := vm.RunString(loadMergeJS(t)); err != nil {
		t.Fatalf("load merge JS: %v", err)
	}

	tests := []struct {
		name   string
		local  map[string]nodeProgress
		remote map[string]nodeProgress
		want   map[string]nodeProgress
	}{
		{
			name:   "none vs in-progress → in-progress",
			local:  map[string]nodeProgress{"a": {State: "none"}},
			remote: map[string]nodeProgress{"a": {State: "in-progress", Tasks: []bool{true}}},
			want:   map[string]nodeProgress{"a": {State: "in-progress", Tasks: []bool{true}}},
		},
		{
			name:   "in-progress vs done → done",
			local:  map[string]nodeProgress{"a": {State: "in-progress"}},
			remote: map[string]nodeProgress{"a": {State: "done"}},
			want:   map[string]nodeProgress{"a": {State: "done"}},
		},
		{
			name:   "done vs in-progress → done",
			local:  map[string]nodeProgress{"a": {State: "done"}},
			remote: map[string]nodeProgress{"a": {State: "in-progress"}},
			want:   map[string]nodeProgress{"a": {State: "done"}},
		},
		{
			name:   "skipped vs done → done wins on equal rank",
			local:  map[string]nodeProgress{"a": {State: "skipped"}},
			remote: map[string]nodeProgress{"a": {State: "done"}},
			want:   map[string]nodeProgress{"a": {State: "done"}},
		},
		{
			name:   "done vs skipped → done",
			local:  map[string]nodeProgress{"a": {State: "done"}},
			remote: map[string]nodeProgress{"a": {State: "skipped"}},
			want:   map[string]nodeProgress{"a": {State: "done"}},
		},
		{
			name:   "skipped vs skipped → skipped",
			local:  map[string]nodeProgress{"a": {State: "skipped"}},
			remote: map[string]nodeProgress{"a": {State: "skipped"}},
			want:   map[string]nodeProgress{"a": {State: "skipped"}},
		},
		{
			name:   "tasks OR merge",
			local:  map[string]nodeProgress{"a": {State: "in-progress", Tasks: []bool{true, false, false}}},
			remote: map[string]nodeProgress{"a": {State: "in-progress", Tasks: []bool{false, true, false}}},
			want:   map[string]nodeProgress{"a": {State: "in-progress", Tasks: []bool{true, true, false}}},
		},
		{
			name:   "tasks OR merge with length difference (local longer)",
			local:  map[string]nodeProgress{"a": {State: "in-progress", Tasks: []bool{true, false, true}}},
			remote: map[string]nodeProgress{"a": {State: "in-progress", Tasks: []bool{false}}},
			want:   map[string]nodeProgress{"a": {State: "in-progress", Tasks: []bool{true, false, true}}},
		},
		{
			name:   "empty local → adopt remote",
			local:  map[string]nodeProgress{},
			remote: map[string]nodeProgress{"a": {State: "done"}},
			want:   map[string]nodeProgress{"a": {State: "done"}},
		},
		{
			name:   "empty remote → keep local",
			local:  map[string]nodeProgress{"a": {State: "in-progress"}},
			remote: map[string]nodeProgress{},
			want:   map[string]nodeProgress{"a": {State: "in-progress"}},
		},
		{
			name:   "disjoint node sets are merged",
			local:  map[string]nodeProgress{"a": {State: "done"}},
			remote: map[string]nodeProgress{"b": {State: "in-progress"}},
			want: map[string]nodeProgress{
				"a": {State: "done"},
				"b": {State: "in-progress"},
			},
		},
		{
			name:   "none vs none → none",
			local:  map[string]nodeProgress{"a": {State: "none"}},
			remote: map[string]nodeProgress{"a": {State: "none"}},
			want:   map[string]nodeProgress{"a": {State: "none"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runMerge(t, vm, tt.local, tt.remote)
			for id, wantNode := range tt.want {
				gotNode, ok := got[id]
				if !ok {
					t.Errorf("node %q missing in result", id)
					continue
				}
				if gotNode.State != wantNode.State {
					t.Errorf("node %q: state = %q, want %q", id, gotNode.State, wantNode.State)
				}
				if len(gotNode.Tasks) != len(wantNode.Tasks) {
					t.Errorf("node %q: tasks len = %d, want %d", id, len(gotNode.Tasks), len(wantNode.Tasks))
					continue
				}
				for i, wt := range wantNode.Tasks {
					if gotNode.Tasks[i] != wt {
						t.Errorf("node %q: tasks[%d] = %v, want %v", id, i, gotNode.Tasks[i], wt)
					}
				}
			}
			// 余分なノードがないか確認
			for id := range got {
				if _, ok := tt.want[id]; !ok {
					t.Errorf("unexpected node %q in result", id)
				}
			}
		})
	}
}
