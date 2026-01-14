/*
 Copyright 2023 NanaFS Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package agentic

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	fridaytools "github.com/basenana/friday/core/tools"
	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/utils"
	"go.uber.org/zap"
)

func init() {
	logger.SetLogger(zap.NewNop().Sugar())
}

func newTools(t *testing.T) (*utils.FileAccess, []*fridaytools.Tool) {
	workdir := t.TempDir()
	fileAccess := utils.NewFileAccess(workdir)
	tools := FileAccessTools(workdir)
	return fileAccess, tools
}

func getToolByName(tools []*fridaytools.Tool, name string) *fridaytools.Tool {
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	return nil
}

func getResultText(r *fridaytools.Result) string {
	if len(r.Content) == 0 {
		return ""
	}
	if tc, ok := r.Content[0].(fridaytools.TextContent); ok {
		return tc.Text
	}
	return ""
}

// ============ File Read Tests ============

func TestFileReadTool_Success(t *testing.T) {
	fa, tools := newTools(t)
	tool := getToolByName(tools, "file_read")
	if tool == nil {
		t.Fatal("file_read tool not found")
	}

	// Create test file
	content := "Hello, World!"
	if err := fa.Write("test.txt", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Execute tool
	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"path": "test.txt",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %s", getResultText(result))
	}
	if getResultText(result) != content {
		t.Errorf("expected %q, got %q", content, getResultText(result))
	}
}

func TestFileReadTool_MissingPath(t *testing.T) {
	_, tools := newTools(t)
	tool := getToolByName(tools, "file_read")
	if tool == nil {
		t.Fatal("file_read tool not found")
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing path")
	}
	if !strings.Contains(getResultText(result), "missing required parameter: path") {
		t.Errorf("expected 'missing required parameter: path', got %q", getResultText(result))
	}
}

func TestFileReadTool_FileNotFound(t *testing.T) {
	_, tools := newTools(t)
	tool := getToolByName(tools, "file_read")
	if tool == nil {
		t.Fatal("file_read tool not found")
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"path": "nonexistent.txt",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent file")
	}
}

// ============ File Write Tests ============

func TestFileWriteTool_Success(t *testing.T) {
	fa, tools := newTools(t)
	tool := getToolByName(tools, "file_write")
	if tool == nil {
		t.Fatal("file_write tool not found")
	}

	content := "Test content to write"

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"path":    "output.txt",
			"content": content,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %s", getResultText(result))
	}

	// Verify file content
	data, err := fa.Read("output.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("expected %q, got %q", content, string(data))
	}
}

func TestFileWriteTool_MissingPath(t *testing.T) {
	_, tools := newTools(t)
	tool := getToolByName(tools, "file_write")
	if tool == nil {
		t.Fatal("file_write tool not found")
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"content": "some content",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing path")
	}
	if !strings.Contains(getResultText(result), "missing required parameter: path") {
		t.Errorf("expected 'missing required parameter: path', got %q", getResultText(result))
	}
}

func TestFileWriteTool_MissingContent(t *testing.T) {
	_, tools := newTools(t)
	tool := getToolByName(tools, "file_write")
	if tool == nil {
		t.Fatal("file_write tool not found")
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"path": "file.txt",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing content")
	}
	if !strings.Contains(getResultText(result), "missing required parameter: content") {
		t.Errorf("expected 'missing required parameter: content', got %q", getResultText(result))
	}
}

// ============ File List Tests ============

func TestFileListTool_EmptyDirectory(t *testing.T) {
	_, tools := newTools(t)
	tool := getToolByName(tools, "file_list")
	if tool == nil {
		t.Fatal("file_list tool not found")
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %s", getResultText(result))
	}

	// Empty directory should return empty array
	var list []any
	if err := json.Unmarshal([]byte(getResultText(result)), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

func TestFileListTool_WithFiles(t *testing.T) {
	fa, tools := newTools(t)
	tool := getToolByName(tools, "file_list")
	if tool == nil {
		t.Fatal("file_list tool not found")
	}

	// Create test files
	if err := fa.Write("file1.txt", []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := fa.Write("file2.txt", []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %s", getResultText(result))
	}

	var list []map[string]any
	if err := json.Unmarshal([]byte(getResultText(result)), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 files, got %d", len(list))
	}

	// Verify file info structure
	expectedFiles := map[string]bool{"file1.txt": true, "file2.txt": true}
	for _, item := range list {
		name, ok := item["name"].(string)
		if !ok {
			t.Error("missing 'name' field in file info")
		}
		if !expectedFiles[name] {
			t.Errorf("unexpected file name: %s", name)
		}
		if _, ok := item["size"]; !ok {
			t.Error("missing 'size' field in file info")
		}
		if _, ok := item["modified"]; !ok {
			t.Error("missing 'modified' field in file info")
		}
		isDir, ok := item["is_dir"].(bool)
		if !ok || isDir {
			t.Error("'is_dir' should be false for files")
		}
	}
}

func TestFileListTool_Subdirectory(t *testing.T) {
	fa, tools := newTools(t)
	tool := getToolByName(tools, "file_list")
	if tool == nil {
		t.Fatal("file_list tool not found")
	}

	// Create subdirectory with files
	if err := fa.MkdirAll("subdir", 0755); err != nil {
		t.Fatal(err)
	}
	if err := fa.Write("subdir/nested.txt", []byte("nested content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"path": "subdir",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %s", getResultText(result))
	}

	var list []map[string]any
	if err := json.Unmarshal([]byte(getResultText(result)), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 file in subdir, got %d", len(list))
	}

	// Verify file info structure
	if len(list) > 0 {
		item := list[0]
		name, ok := item["name"].(string)
		if !ok || name != "nested.txt" {
			t.Errorf("expected name 'nested.txt', got %q", name)
		}
		if _, ok := item["size"]; !ok {
			t.Error("missing 'size' field in file info")
		}
		if _, ok := item["modified"]; !ok {
			t.Error("missing 'modified' field in file info")
		}
		isDir, ok := item["is_dir"].(bool)
		if !ok || isDir {
			t.Error("'is_dir' should be false for files")
		}
	}
}

func TestFileListTool_InvalidPath(t *testing.T) {
	_, tools := newTools(t)
	tool := getToolByName(tools, "file_list")
	if tool == nil {
		t.Fatal("file_list tool not found")
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"path": "../invalid",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for invalid path")
	}
}

// ============ File Parse Tests ============

func TestFileParseTool_TextFile(t *testing.T) {
	fa, tools := newTools(t)
	tool := getToolByName(tools, "file_parse")
	if tool == nil {
		t.Fatal("file_parse tool not found")
	}

	// Create test text file
	content := "This is a test document."
	if err := fa.Write("test.txt", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"path": "test.txt",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %s", getResultText(result))
	}
	if getResultText(result) != content {
		t.Errorf("expected %q, got %q", content, getResultText(result))
	}
}

func TestFileParseTool_MarkdownFile(t *testing.T) {
	fa, tools := newTools(t)
	tool := getToolByName(tools, "file_parse")
	if tool == nil {
		t.Fatal("file_parse tool not found")
	}

	// Create test markdown file
	content := "# Title\n\nSome content here."
	if err := fa.Write("test.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"path": "test.md",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %s", getResultText(result))
	}
}

func TestFileParseTool_HtmlFile(t *testing.T) {
	fa, tools := newTools(t)
	tool := getToolByName(tools, "file_parse")
	if tool == nil {
		t.Fatal("file_parse tool not found")
	}

	// Create test HTML file
	content := "<html><body><h1>Hello</h1></body></html>"
	if err := fa.Write("test.html", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"path": "test.html",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %s", getResultText(result))
	}
}

func TestFileParseTool_MissingPath(t *testing.T) {
	_, tools := newTools(t)
	tool := getToolByName(tools, "file_parse")
	if tool == nil {
		t.Fatal("file_parse tool not found")
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing path")
	}
	if !strings.Contains(getResultText(result), "missing required parameter: path") {
		t.Errorf("expected 'missing required parameter: path', got %q", getResultText(result))
	}
}

func TestFileParseTool_UnsupportedFormat(t *testing.T) {
	fa, tools := newTools(t)
	tool := getToolByName(tools, "file_parse")
	if tool == nil {
		t.Fatal("file_parse tool not found")
	}

	// Create file with unsupported extension
	if err := fa.Write("test.xyz", []byte("some content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"path": "test.xyz",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for unsupported format")
	}
	if !strings.Contains(getResultText(result), "unsupported file format") {
		t.Errorf("expected 'unsupported file format', got %q", getResultText(result))
	}
}

func TestFileParseTool_FileNotFound(t *testing.T) {
	_, tools := newTools(t)
	tool := getToolByName(tools, "file_parse")
	if tool == nil {
		t.Fatal("file_parse tool not found")
	}

	result, err := tool.Handler(context.Background(), &fridaytools.Request{
		Arguments: map[string]any{
			"path": "nonexistent.txt",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent file")
	}
}
