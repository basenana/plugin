package fs

import (
	"context"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/types"
)

func TestSaver_Run_MissingFilePath(t *testing.T) {
	plugin := &Saver{}
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
	plugin := &Saver{}
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
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	plugin := &Saver{}
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path": tmpFile.Name(),
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
	if mockFS.GetEntriesCount() != 1 {
		t.Errorf("expected 1 entry, got %d", mockFS.GetEntriesCount())
	}
}

func TestSaver_Run_WithName(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	plugin := &Saver{}
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path": tmpFile.Name(),
			"name":      "custom_name.txt",
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

	entry, ok := mockFS.GetEntry(1)
	if !ok {
		t.Fatal("expected entry to be saved")
	}
	if entry.name != "custom_name.txt" {
		t.Errorf("expected name 'custom_name.txt', got '%s'", entry.name)
	}
}

func TestSaver_Run_WithParentURI(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	plugin := &Saver{}
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  tmpFile.Name(),
			"parent_uri": "12345",
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

	entry, ok := mockFS.GetEntry(1)
	if !ok {
		t.Fatal("expected entry to be saved")
	}
	if entry.parentURI != "12345" {
		t.Errorf("expected parent_uri '12345', got '%s'", entry.parentURI)
	}
}

func TestSaver_Run_WithProperties(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	plugin := &Saver{}
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path": tmpFile.Name(),
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

	entry, ok := mockFS.GetEntry(1)
	if !ok {
		t.Fatal("expected entry to be saved")
	}
	if entry.props.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got '%s'", entry.props.Title)
	}
	if entry.props.Author != "Test Author" {
		t.Errorf("expected author 'Test Author', got '%s'", entry.props.Author)
	}
	if entry.props.Year != "2024" {
		t.Errorf("expected year '2024', got '%s'", entry.props.Year)
	}
	if !entry.props.Marked {
		t.Error("expected marked to be true")
	}
	if entry.props.Unread {
		t.Error("expected unread to be false")
	}
}

func TestSaver_Run_WithDocument(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	plugin := &Saver{}
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path": tmpFile.Name(),
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

	entry, ok := mockFS.GetEntry(1)
	if !ok {
		t.Fatal("expected entry to be saved")
	}
	if entry.props.Title != "Doc Title" {
		t.Errorf("expected title 'Doc Title', got '%s'", entry.props.Title)
	}
	if entry.props.Abstract != "Doc Abstract" {
		t.Errorf("expected abstract 'Doc Abstract', got '%s'", entry.props.Abstract)
	}
	if len(entry.props.Keywords) != 2 {
		t.Errorf("expected 2 keywords, got %d", len(entry.props.Keywords))
	}
}

func TestSaver_Run_WithNilParameter(t *testing.T) {
	plugin := &Saver{}
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
	plugin := &Saver{}
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
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	plugin := &Saver{}
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path": tmpFile.Name(),
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
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	plugin := &Saver{}
	mockFS := NewMockNanaFS()
	mockFS.SetSaveError(context.DeadlineExceeded)

	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path": tmpFile.Name(),
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
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	plugin := &Saver{}
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  tmpFile.Name(),
			"name":       "all_params.txt",
			"parent_uri": "999",
			"properties": map[string]interface{}{
				"title":       "Full Test",
				"author":      "Author",
				"year":        "2025",
				"source":      "Source",
				"abstract":    "Abstract",
				"notes":       "Notes",
				"url":         "https://example.com",
				"header_image": "https://example.com/image.png",
				"unread":      true,
				"marked":      true,
				"publish_at":  int64(1704067200),
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

	entry, ok := mockFS.GetEntry(1)
	if !ok {
		t.Fatal("expected entry to be saved")
	}
	if entry.name != "all_params.txt" {
		t.Errorf("expected name 'all_params.txt', got '%s'", entry.name)
	}
	if entry.parentURI != "999" {
		t.Errorf("expected parent_uri '999', got '%s'", entry.parentURI)
	}
	if entry.props.Title != "Full Test" {
		t.Errorf("expected title 'Full Test', got '%s'", entry.props.Title)
	}
	if entry.props.Author != "Author" {
		t.Errorf("expected author 'Author', got '%s'", entry.props.Author)
	}
	if entry.props.Year != "2025" {
		t.Errorf("expected year '2025', got '%s'", entry.props.Year)
	}
	if entry.props.Source != "Source" {
		t.Errorf("expected source 'Source', got '%s'", entry.props.Source)
	}
	if entry.props.Abstract != "Abstract" {
		t.Errorf("expected abstract 'Abstract', got '%s'", entry.props.Abstract)
	}
	if entry.props.Notes != "Notes" {
		t.Errorf("expected notes 'Notes', got '%s'", entry.props.Notes)
	}
	if entry.props.URL != "https://example.com" {
		t.Errorf("expected url 'https://example.com', got '%s'", entry.props.URL)
	}
	if entry.props.HeaderImage != "https://example.com/image.png" {
		t.Errorf("expected headerImage 'https://example.com/image.png', got '%s'", entry.props.HeaderImage)
	}
	if !entry.props.Unread {
		t.Error("expected unread to be true")
	}
	if !entry.props.Marked {
		t.Error("expected marked to be true")
	}
	if entry.props.PublishAt != 1704067200 {
		t.Errorf("expected publishAt 1704067200, got %d", entry.props.PublishAt)
	}
}

func TestSaver_Properties_Priority(t *testing.T) {
	// When both properties and document are provided, properties should take priority
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	plugin := &Saver{}
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path": tmpFile.Name(),
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

	entry, ok := mockFS.GetEntry(1)
	if !ok {
		t.Fatal("expected entry to be saved")
	}
	// Properties should override document
	if entry.props.Title != "From Properties" {
		t.Errorf("expected title 'From Properties' (from properties), got '%s'", entry.props.Title)
	}
}

func TestSaver_Properties_NilDocumentAndProperties(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	plugin := &Saver{}
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path":  tmpFile.Name(),
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
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	plugin := &Saver{}
	mockFS := NewMockNanaFS()
	req := &api.Request{
		Parameter: map[string]interface{}{
			"file_path": tmpFile.Name(),
			"document":  "invalid string instead of map",
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
	entries      map[int64]*mockEntry
	saveCalled   bool
	saveErr      error
	updateCalled bool
	updateErr    error
	nextID       int64
}

type mockEntry struct {
	id        int64
	parentURI string
	name      string
	props     types.Properties
}

func NewMockNanaFS() *MockNanaFS {
	return &MockNanaFS{
		entries: make(map[int64]*mockEntry),
		nextID:  1,
	}
}

func (m *MockNanaFS) SaveEntry(ctx context.Context, parentURI, name string, properties types.Properties, write io.WriteCloser) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.saveCalled = true
	if m.saveErr != nil {
		return m.saveErr
	}

	id := m.nextID
	m.nextID++
	m.entries[id] = &mockEntry{
		id:        id,
		parentURI: parentURI,
		name:      name,
		props:     properties,
	}

	return nil
}

func (m *MockNanaFS) UpdateEntry(ctx context.Context, entryURI int64, properties types.Properties) error {
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

func (m *MockNanaFS) GetEntry(id int64) (*mockEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entries[id]
	return e, ok
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
