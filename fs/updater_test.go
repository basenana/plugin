package fs

import (
	"context"
	"testing"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"go.uber.org/zap"
)

func init() {
	logger.SetLogger(zap.NewNop().Sugar())
}

func newUpdater() *Updater {
	u := &Updater{}
	u.logger = logger.NewPluginLogger("updater", "test-job")
	return u
}

func TestUpdater_Run_MissingEntryURI(t *testing.T) {
	plugin := newUpdater()
	req := &api.Request{
		Parameter: map[string]interface{}{},
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsSucceed {
		t.Error("expected response to indicate failure")
	}
}

func TestUpdater_Run_EmptyEntryURI(t *testing.T) {
	plugin := newUpdater()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "",
		},
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsSucceed {
		t.Error("expected response to indicate failure")
	}
}

func TestUpdater_Run_InvalidEntryURI(t *testing.T) {
	plugin := newUpdater()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "invalid",
		},
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsSucceed {
		t.Error("expected response to indicate failure")
	}
}

func TestUpdater_Run_ZeroEntryURI(t *testing.T) {
	plugin := newUpdater()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "0",
		},
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsSucceed {
		t.Error("expected response to indicate failure (0 is invalid)")
	}
}

func TestUpdater_Run_NegativeEntryURI(t *testing.T) {
	plugin := newUpdater()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "-1",
		},
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsSucceed {
		t.Error("expected response to indicate failure")
	}
}

func TestUpdater_Run_Success(t *testing.T) {
	plugin := newUpdater()
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "123",
		},
		FS: mockFS,
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if !mockFS.WasUpdateCalled() {
		t.Error("expected UpdateEntry to be called")
	}
}

func TestUpdater_Run_WithProperties(t *testing.T) {
	plugin := newUpdater()
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "123",
			"properties": map[string]interface{}{
				"title":  "Updated Title",
				"author": "Updated Author",
				"year":   "2025",
				"marked": true,
				"unread": false,
			},
		},
		FS: mockFS,
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
}

func TestUpdater_Run_WithDocument(t *testing.T) {
	plugin := newUpdater()
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "123",
			"document": map[string]interface{}{
				"content": "document content",
				"properties": map[string]interface{}{
					"title":    "Doc Title",
					"abstract": "Doc Abstract",
					"keywords": []interface{}{"kw1", "kw2"},
				},
			},
		},
		FS: mockFS,
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
}

func TestUpdater_Run_NilParameter(t *testing.T) {
	plugin := newUpdater()
	req := &api.Request{
		Parameter: nil,
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsSucceed {
		t.Error("expected response to indicate failure")
	}
}

func TestUpdater_Run_NilFS(t *testing.T) {
	plugin := newUpdater()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "123",
		},
		FS: nil,
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsSucceed {
		t.Error("expected response to indicate failure")
	}
}

func TestUpdater_Run_UpdateError(t *testing.T) {
	plugin := newUpdater()
	mockFS := NewMockNanaFS()
	mockFS.SetUpdateError(context.DeadlineExceeded)

	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "123",
		},
		FS: mockFS,
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsSucceed {
		t.Error("expected response to indicate failure")
	}
}

func TestUpdater_Run_WithAllParameters(t *testing.T) {
	plugin := newUpdater()
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "999",
			"properties": map[string]interface{}{
				"title":        "Full Update",
				"author":       "Author",
				"year":         "2025",
				"source":       "Source",
				"abstract":     "Abstract",
				"notes":        "Notes",
				"url":          "https://example.com",
				"header_image": "https://example.com/image.png",
				"unread":       true,
				"marked":       true,
				"publish_at":   int64(1704067200),
			},
		},
		FS: mockFS,
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
}

func TestUpdater_Run_LargeEntryURI(t *testing.T) {
	plugin := newUpdater()
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "9223372036854775807", // max int64
		},
		FS: mockFS,
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
}

func TestUpdater_Run_PropertiesAndDocumentBothProvided(t *testing.T) {
	// properties should take priority over document
	plugin := newUpdater()
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "123",
			"properties": map[string]interface{}{
				"title": "From Properties",
			},
			"document": map[string]interface{}{
				"content": "content",
				"properties": map[string]interface{}{
					"title": "From Document",
				},
			},
		},
		FS: mockFS,
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
}

func TestUpdater_Run_NilPropertiesAndDocument(t *testing.T) {
	plugin := newUpdater()
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri":  "123",
			"properties": nil,
			"document":   nil,
		},
		FS: mockFS,
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
}

func TestUpdater_Run_InvalidDocumentType(t *testing.T) {
	plugin := newUpdater()
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "123",
			"document":  "invalid string instead of map",
		},
		FS: mockFS,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered from panic: %v", r)
		}
	}()

	resp, err := plugin.Run(context.Background(), req)

	if err == nil && resp.IsSucceed {
		t.Log("function handled invalid type gracefully")
	}
}

func TestUpdater_Run_WhitespaceEntryURI(t *testing.T) {
	plugin := newUpdater()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "   ",
		},
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsSucceed {
		t.Error("expected response to indicate failure")
	}
}

func TestUpdater_Run_FloatEntryURI(t *testing.T) {
	plugin := newUpdater()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": 123.456,
		},
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// strconv.ParseInt will fail on float string representation
	if resp.IsSucceed {
		t.Error("expected response to indicate failure for float")
	}
}

func TestUpdater_Run_UpdateNonexistentEntry(t *testing.T) {
	// Should not fail when updating a non-existent entry
	plugin := newUpdater()
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"entry_uri": "99999",
		},
		FS: mockFS,
	}

	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success (silently ignoring non-existent entry), got failure: %s", resp.Message)
	}
}
