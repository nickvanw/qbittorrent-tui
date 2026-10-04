package api

import (
	"encoding/json"
	"testing"
)

func TestCategoryDownloadPathUnmarshal(t *testing.T) {
	tests := map[string]string{
		`{"name":"a","savePath":"/s","download_path":"/incomplete"}`: "/incomplete",
		`{"name":"a","savePath":"/s","download_path":false}`:         "",
		`{"name":"a","savePath":"/s","download_path":null}`:          "",
		`{"name":"a","savePath":"/s"}`:                               "",
	}
	for in, want := range tests {
		var c Category
		if err := json.Unmarshal([]byte(in), &c); err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if string(c.DownloadPath) != want {
			t.Errorf("%s: got %q want %q", in, c.DownloadPath, want)
		}
	}
}
