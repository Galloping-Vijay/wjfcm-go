package requestlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wjfcm-go/internal/config"
)

func TestWriteRotatesAndFindsRequestLog(t *testing.T) {
	store := NewStore(config.LogConfig{
		RequestEnabled:      true,
		RequestPath:         t.TempDir(),
		RequestOutput:       "file",
		RequestLevel:        "info",
		RequestMaxBodyBytes: 1024,
		RequestMaxFileBytes: 220,
	})

	first := &Entry{
		RequestID: "req-one",
		Level:     "info",
		Method:    "GET",
		Path:      "/first",
		Request:   Payload{Body: strings.Repeat("a", 180)},
		Response:  Payload{Body: "ok"},
		Status:    200,
	}
	second := &Entry{
		RequestID: "req-two",
		Level:     "info",
		Method:    "GET",
		Path:      "/second",
		Request:   Payload{Body: strings.Repeat("b", 180)},
		Response:  Payload{Body: "ok"},
		Status:    200,
	}

	if err := store.write(first); err != nil {
		t.Fatalf("write first request log: %v", err)
	}
	if err := store.write(second); err != nil {
		t.Fatalf("write second request log: %v", err)
	}

	if got, _, err := store.Find(first.RequestID); err != nil || got.Path != first.Path {
		t.Fatalf("find first request log = (%+v, %v), want path %s", got, err, first.Path)
	}
	if got, _, err := store.Find(second.RequestID); err != nil || got.Path != second.Path {
		t.Fatalf("find second request log = (%+v, %v), want path %s", got, err, second.Path)
	}

	files := 0
	err := filepath.WalkDir(store.cfg.RequestPath, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d != nil && !d.IsDir() {
			files++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk request log dir: %v", err)
	}
	if files < 2 {
		t.Fatalf("rotated request log files = %d, want at least 2", files)
	}
}
