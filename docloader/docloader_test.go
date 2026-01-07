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

package docloader

import (
	"context"
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

// testFileAccess provides a unified FileAccess for all tests in the package
var testFileAccess = utils.NewFileAccess("")

func newDocLoader(t *testing.T) *DocLoader {
	return NewDocLoader(types.PluginCall{
		JobID:       "test-job",
		Workflow:    "test-workflow",
		Namespace:   "test-namespace",
		WorkingPath: testFileAccess.Workdir(),
		PluginName:  "",
		Version:     "",
		Params:      map[string]string{},
	}).(*DocLoader)
}

func TestDocLoader_NameTypeVersion(t *testing.T) {
	loader := newDocLoader(t)
	if loader.Name() != "docloader" {
		t.Errorf("Name() = %q, want %q", loader.Name(), "docloader")
	}
	if loader.Version() != "1.0" {
		t.Errorf("Version() = %q, want %q", loader.Version(), "1.0")
	}
}

func TestDocLoader_Run_MissingFilePath(t *testing.T) {
	loader := newDocLoader(t)
	req := &api.Request{
		Parameter: map[string]any{},
	}

	resp, _ := loader.Run(context.Background(), req)
	if resp.IsSucceed {
		t.Error("Run should fail when file_path is missing")
	}
}

func TestDocLoader_Run_FileNotFound(t *testing.T) {
	loader := newDocLoader(t)
	req := &api.Request{
		Parameter: map[string]any{"file_path": "nonexistent/file.pdf"},
	}

	resp, _ := loader.Run(context.Background(), req)
	if resp.IsSucceed {
		t.Error("Run should fail when file not found")
	}
}

func TestDocLoader_Run_UnsupportedFormat(t *testing.T) {
	loader := newDocLoader(t)

	if err := testFileAccess.Write("test.xyz", []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	req := &api.Request{
		Parameter: map[string]any{"file_path": "test.xyz"},
	}

	resp, _ := loader.Run(context.Background(), req)
	if resp.IsSucceed {
		t.Error("Run should fail for unsupported format")
	}
}

func TestDocLoader_Run_TextFile(t *testing.T) {
	loader := newDocLoader(t)

	content := `# Test Document

This is a test paragraph.`

	if err := testFileAccess.Write("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	req := &api.Request{
		Parameter: map[string]any{"file_path": "test.txt"},
	}

	resp, err := loader.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !resp.IsSucceed {
		t.Errorf("Run failed: %s", resp.Message)
	}

	result := resp.Results
	filePath, ok := result["file_path"].(string)
	if !ok || filePath != "test.txt" {
		t.Errorf("file_path = %v", filePath)
	}

	doc := result["document"].(map[string]any)
	if doc["content"] == "" {
		t.Error("document should contain content")
	}
}

func TestDocLoader_Run_MarkdownFile(t *testing.T) {
	loader := newDocLoader(t)

	content := `# Markdown Document Title

Some content here.`

	if err := testFileAccess.Write("document.md", []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	req := &api.Request{
		Parameter: map[string]any{"file_path": "document.md"},
	}

	resp, err := loader.Run(context.Background(), req)
	if err != nil || !resp.IsSucceed {
		t.Fatalf("Run failed: %v, %s", err, resp.Message)
	}

	doc := resp.Results["document"].(map[string]any)
	props := doc["properties"].(map[string]any)
	if props["title"] != "Markdown Document Title" {
		t.Errorf("title = %v, want %v", props["title"], "Markdown Document Title")
	}
}

func TestDocLoader_Run_HTMLFile(t *testing.T) {
	loader := newDocLoader(t)

	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>HTML Test</title>
    <meta name="author" content="Test Author">
</head>
<body>Content here</body>
</html>`

	if err := testFileAccess.Write("page.html", []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	req := &api.Request{
		Parameter: map[string]any{"file_path": "page.html"},
	}

	resp, err := loader.Run(context.Background(), req)
	if err != nil || !resp.IsSucceed {
		t.Fatalf("Run failed: %v, %s", err, resp.Message)
	}

	doc := resp.Results["document"].(map[string]any)
	props := doc["properties"].(map[string]any)
	if props["author"] != "Test Author" {
		t.Errorf("author = %v", props["author"])
	}
	if doc["content"] == "" {
		t.Error("document should contain content")
	}
}

func TestDocLoader_Run_DefaultTitle(t *testing.T) {
	loader := newDocLoader(t)

	// Content without any meaningful title (short line that would be skipped)
	content := `12345`

	if err := testFileAccess.Write("my_custom_file.txt", []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	req := &api.Request{
		Parameter: map[string]any{"file_path": "my_custom_file.txt"},
	}

	resp, _ := loader.Run(context.Background(), req)
	doc := resp.Results["document"].(map[string]any)
	props := doc["properties"].(map[string]any)
	if props["title"] != "my_custom_file" {
		t.Errorf("title = %v, want %v", props["title"], "my_custom_file")
	}
}
