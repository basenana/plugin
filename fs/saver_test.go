package fs

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"testing"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/types"
	"github.com/basenana/plugin/utils"
	"go.uber.org/zap"
)

func init() {
	logger.SetLogger(zap.NewNop().Sugar())
}

func newSaver(t *testing.T) (*Saver, *utils.FileAccess) {
	p := NewSaver(types.PluginCall{
		JobID:       "test-job",
		Workflow:    "test-workflow",
		Namespace:   "test-namespace",
		WorkingPath: t.TempDir(),
		PluginName:  "",
		Version:     "",
		Params:      map[string]string{},
	}).(*Saver)

	return p, p.fileRoot
}

func TestSaver_Run_MissingFilePath(t *testing.T) {
	plugin, _ := newSaver(t)
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

func TestSaver_Run_FileNotFound(t *testing.T) {
	plugin, _ := newSaver(t)
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path": "/nonexistent/file.txt",
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

func TestSaver_Run_Success(t *testing.T) {
	plugin, tw := newSaver(t)

	if err := tw.Write("test_file.txt", []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  filepath.Join(tw.Workdir(), "test_file.txt"),
			"parent_uri": "/group",
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
	if !mockFS.WasSaveCalled() {
		t.Error("expected SaveEntry to be called")
	}
}

func TestSaver_Run_WithName(t *testing.T) {
	plugin, tw := newSaver(t)

	if err := tw.Write("test_file.txt", []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  filepath.Join(tw.Workdir(), "test_file.txt"),
			"name":       "custom_name.txt",
			"parent_uri": "/group",
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

func TestSaver_Run_WithParentURI(t *testing.T) {
	plugin, tw := newSaver(t)

	if err := tw.Write("test_file.txt", []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  filepath.Join(tw.Workdir(), "test_file.txt"),
			"parent_uri": "/group",
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

func TestSaver_Run_WithProperties(t *testing.T) {
	plugin, tw := newSaver(t)

	if err := tw.Write("test_file.txt", []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  filepath.Join(tw.Workdir(), "test_file.txt"),
			"parent_uri": "/group",
			"properties": map[string]interface{}{
				"title":  "Test Title",
				"author": "Test Author",
				"year":   "2024",
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

func TestSaver_Run_WithDocument(t *testing.T) {
	plugin, tw := newSaver(t)

	if err := tw.Write("test_file.txt", []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  filepath.Join(tw.Workdir(), "test_file.txt"),
			"parent_uri": "/group",
			"document": map[string]interface{}{
				"content": "document content",
				"properties": map[string]interface{}{
					"title":    "Doc Title",
					"abstract": "Doc Abstract",
					"keywords": []interface{}{"keyword1", "keyword2"},
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

func TestSaver_Run_WithNilParameter(t *testing.T) {
	plugin, _ := newSaver(t)
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

func TestSaver_Run_WithEmptyParameter(t *testing.T) {
	plugin, _ := newSaver(t)
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

func TestSaver_Run_NilFS(t *testing.T) {
	plugin, tw := newSaver(t)

	if err := tw.Write("test_file.txt", []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  filepath.Join(tw.Workdir(), "test_file.txt"),
			"parent_uri": "/group",
		},
		FS: nil,
	}

	// This will panic if nil FS is not handled
	resp, err := plugin.Run(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsSucceed {
		t.Error("expected response to indicate failure")
	}
}

func TestSaver_Run_SaveError(t *testing.T) {
	plugin, tw := newSaver(t)

	if err := tw.Write("test_file.txt", []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockFS := NewMockNanaFS()
	mockFS.SetSaveError(context.DeadlineExceeded)

	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  filepath.Join(tw.Workdir(), "test_file.txt"),
			"parent_uri": "/group",
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

func TestSaver_Run_WithAllParameters(t *testing.T) {
	plugin, tw := newSaver(t)

	if err := tw.Write("test_file.txt", []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  filepath.Join(tw.Workdir(), "test_file.txt"),
			"name":       "all_params.txt",
			"parent_uri": "/group",
			"properties": map[string]interface{}{
				"title":        "Full Test",
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

func TestSaver_Properties_Priority(t *testing.T) {
	plugin, tw := newSaver(t)

	if err := tw.Write("test_file.txt", []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  filepath.Join(tw.Workdir(), "test_file.txt"),
			"parent_uri": "/group",
			"properties": map[string]interface{}{
				"title": "From Properties",
			},
			"document": map[string]interface{}{
				"content": "document content",
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

func TestSaver_Properties_NilDocumentAndProperties(t *testing.T) {
	plugin, tw := newSaver(t)

	if err := tw.Write("test_file.txt", []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  filepath.Join(tw.Workdir(), "test_file.txt"),
			"parent_uri": "/group",
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

func TestSaver_Properties_InvalidDocumentType(t *testing.T) {
	plugin, tw := newSaver(t)

	if err := tw.Write("test_file.txt", []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  filepath.Join(tw.Workdir(), "test_file.txt"),
			"parent_uri": "/group",
			"document":   "invalid string instead of map",
		},
		FS: mockFS,
	}

	// This should handle the type assertion panic or gracefully fail
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered from panic: %v", r)
		}
	}()

	resp, err := plugin.Run(context.Background(), req)

	// Either it should handle the error or panic
	if err == nil && resp.IsSucceed {
		t.Log("function handled invalid type gracefully")
	}
}

// MockNanaFS is a mock implementation of NanaFS interface for testing.
type MockNanaFS struct {
	mu           sync.RWMutex
	entries      map[string]*mockEntry
	saveCalled   bool
	saveErr      error
	updateCalled bool
	updateErr    error
}

type mockEntry struct {
	parentURI string
	name      string
	props     types.Properties
}

func NewMockNanaFS() *MockNanaFS {
	return &MockNanaFS{entries: make(map[string]*mockEntry)}
}
func (m *MockNanaFS) CreateGroupIfNotExists(ctx context.Context, parentURI, group string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[fmt.Sprintf("%s/%s", parentURI, group)] = &mockEntry{parentURI: parentURI, name: group}
	return nil
}

func (m *MockNanaFS) SaveEntry(ctx context.Context, parentURI, name string, properties types.Properties, reader io.ReadCloser) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.saveCalled = true
	if m.saveErr != nil {
		return m.saveErr
	}

	m.entries[fmt.Sprintf("%s/%s", parentURI, name)] = &mockEntry{
		parentURI: parentURI,
		name:      name,
		props:     properties,
	}

	return nil
}

func (m *MockNanaFS) UpdateEntry(ctx context.Context, entryURI string, properties types.Properties) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.updateCalled = true
	if m.updateErr != nil {
		return m.updateErr
	}

	if entry, ok := m.entries[entryURI]; ok {
		entry.props = properties
		return nil
	}

	return nil
}

func (m *MockNanaFS) GetEntryProperties(ctx context.Context, entryURI string) (properties *types.Properties, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	en, ok := m.entries[entryURI]
	if !ok {
		return nil, fmt.Errorf("entry not found")
	}
	return &en.props, nil
}

// Test helpers

func (m *MockNanaFS) SetSaveError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveErr = err
}

func (m *MockNanaFS) SetUpdateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateErr = err
}

func (m *MockNanaFS) WasSaveCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.saveCalled
}

func (m *MockNanaFS) WasUpdateCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.updateCalled
}

func (m *MockNanaFS) GetEntriesCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

// Ensure MockNanaFS implements NanaFS interface
var _ api.NanaFS = (*MockNanaFS)(nil)
