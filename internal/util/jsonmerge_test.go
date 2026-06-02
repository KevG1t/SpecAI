package util

import (
	"encoding/json"
	"testing"
)

func TestMergeJSON(t *testing.T) {
	tests := []struct {
		name    string
		dst     []byte
		src     map[string]any
		want    map[string]any
		wantErr bool
	}{
		{
			name: "empty dst becomes src",
			dst:  nil,
			src:  map[string]any{"key": "value"},
			want: map[string]any{"key": "value"},
		},
		{
			name:    "empty dst empty bytes becomes src",
			dst:     []byte{},
			src:     map[string]any{"a": 1.0},
			want:    map[string]any{"a": 1.0},
		},
		{
			name: "nested merge preserves user keys",
			dst:  mustMarshal(t, map[string]any{"user": map[string]any{"name": "alice", "role": "admin"}}),
			src:  map[string]any{"user": map[string]any{"email": "alice@example.com"}},
			want: map[string]any{"user": map[string]any{"name": "alice", "role": "admin", "email": "alice@example.com"}},
		},
		{
			name: "non-map src key overwrites dst",
			dst:  mustMarshal(t, map[string]any{"count": 1.0}),
			src:  map[string]any{"count": 2.0},
			want: map[string]any{"count": 2.0},
		},
		{
			name: "corrupted dst starts fresh with src",
			dst:  []byte("not valid json {{{{"),
			src:  map[string]any{"fresh": true},
			want: map[string]any{"fresh": true},
		},
		{
			name: "idempotent re-merge",
			dst:  mustMarshal(t, map[string]any{"mcp": map[string]any{"sdd-memory": map[string]any{"command": "sdd-memory", "type": "local"}}}),
			src:  map[string]any{"mcp": map[string]any{"sdd-memory": map[string]any{"command": "sdd-memory", "type": "local"}}},
			want: map[string]any{"mcp": map[string]any{"sdd-memory": map[string]any{"command": "sdd-memory", "type": "local"}}},
		},
		{
			name: "top-level src key not in dst is added",
			dst:  mustMarshal(t, map[string]any{"existing": "yes"}),
			src:  map[string]any{"new": "yes"},
			want: map[string]any{"existing": "yes", "new": "yes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MergeJSON(tt.dst, tt.src)
			if (err != nil) != tt.wantErr {
				t.Fatalf("MergeJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			var gotMap map[string]any
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Fatalf("result is not valid JSON: %v", err)
			}
			gotBytes, _ := json.Marshal(gotMap)
			wantBytes, _ := json.Marshal(tt.want)
			if string(gotBytes) != string(wantBytes) {
				t.Errorf("MergeJSON() = %s, want %s", gotBytes, wantBytes)
			}
		})
	}
}

func TestDeepMerge(t *testing.T) {
	tests := []struct {
		name string
		dst  map[string]any
		src  map[string]any
		want map[string]any
	}{
		{
			name: "simple key added",
			dst:  map[string]any{"a": 1},
			src:  map[string]any{"b": 2},
			want: map[string]any{"a": 1, "b": 2},
		},
		{
			name: "nested maps merged recursively",
			dst:  map[string]any{"top": map[string]any{"a": 1}},
			src:  map[string]any{"top": map[string]any{"b": 2}},
			want: map[string]any{"top": map[string]any{"a": 1, "b": 2}},
		},
		{
			name: "src map overwrites non-map dst value",
			dst:  map[string]any{"key": "string"},
			src:  map[string]any{"key": map[string]any{"nested": true}},
			want: map[string]any{"key": map[string]any{"nested": true}},
		},
		{
			name: "empty src no-ops",
			dst:  map[string]any{"a": 1},
			src:  map[string]any{},
			want: map[string]any{"a": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeepMerge(tt.dst, tt.src)
			gotBytes, _ := json.Marshal(tt.dst)
			wantBytes, _ := json.Marshal(tt.want)
			if string(gotBytes) != string(wantBytes) {
				t.Errorf("DeepMerge() dst = %s, want %s", gotBytes, wantBytes)
			}
		})
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("mustMarshal: %v", err)
	}
	return b
}
