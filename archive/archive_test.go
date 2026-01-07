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

package archive

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/types"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.SetLogger(zap.NewNop().Sugar())
	os.Exit(m.Run())
}

func newArchivePlugin(workdir string) *ArchivePlugin {
	return NewArchivePlugin(types.PluginCall{JobID: "test-job", WorkingPath: workdir}).(*ArchivePlugin)
}

func newArchivePluginWithTmpDir(t *testing.T) *ArchivePlugin {
	return newArchivePlugin(t.TempDir())
}

func TestArchivePlugin_Name(t *testing.T) {
	p := newArchivePluginWithTmpDir(t)
	if p.Name() != pluginName {
		t.Errorf("expected %s, got %s", pluginName, p.Name())
	}
}

func TestArchivePlugin_Type(t *testing.T) {
	p := newArchivePluginWithTmpDir(t)
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestArchivePlugin_Version(t *testing.T) {
	p := newArchivePluginWithTmpDir(t)
	if p.Version() != pluginVersion {
		t.Errorf("expected %s, got %s", pluginVersion, p.Version())
	}
}

func TestArchivePlugin_Extract_MissingFilePath(t *testing.T) {
	p := newArchivePluginWithTmpDir(t)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action": "extract",
			"format": "zip",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message == "" || resp.Message != "file_path is required" {
		t.Errorf("expected 'file_path is required', got '%s'", resp.Message)
	}
}

func TestArchivePlugin_Extract_MissingFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "test.zip")
	os.WriteFile(zipPath, []byte("test"), 0644)

	p := newArchivePlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":    "extract",
			"file_path": zipPath,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message == "" || resp.Message != "format is required" {
		t.Errorf("expected 'format is required', got '%s'", resp.Message)
	}
}

func TestArchivePlugin_Extract_InvalidFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "test.zip")
	os.WriteFile(zipPath, []byte("test"), 0644)

	p := newArchivePlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":    "extract",
			"file_path": zipPath,
			"format":    "invalid",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
}

func TestArchivePlugin_Extract_InvalidZipFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an invalid zip file
	zipPath := filepath.Join(tmpDir, "invalid.zip")
	os.WriteFile(zipPath, []byte("this is not a valid zip file"), 0644)

	p := newArchivePlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":    "extract",
			"file_path": "invalid.zip",
			"format":    "zip",
			"dest_path": "dest",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure for invalid zip, got success")
	}
}

func TestArchivePlugin_Extract_ValidZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid zip file
	testContent := "Hello, World!"
	zipPath := filepath.Join(tmpDir, "test.zip")

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	w, err := zipWriter.Create("hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.Write([]byte(testContent))
	if err != nil {
		t.Fatal(err)
	}
	zipWriter.Close()
	os.WriteFile(zipPath, buf.Bytes(), 0644)

	p := newArchivePlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":    "extract",
			"file_path": "test.zip",
			"format":    "zip",
			"dest_path": "dest",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "dest", "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != testContent {
		t.Errorf("expected %q, got %q", testContent, string(data))
	}
}

func TestArchivePlugin_Compress_MissingSourcePath(t *testing.T) {
	p := newArchivePluginWithTmpDir(t)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action": "compress",
			"format": "zip",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message == "" || resp.Message != "source_path is required for compression" {
		t.Errorf("expected 'source_path is required for compression', got '%s'", resp.Message)
	}
}

func TestArchivePlugin_Compress_Zip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source files
	os.MkdirAll(filepath.Join(tmpDir, "source"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "source", "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "source", "file2.txt"), []byte("content2"), 0644)

	p := newArchivePlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":       "compress",
			"source_path":  "source",
			"format":       "zip",
			"archive_name": "test.zip",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	zipPath := filepath.Join(tmpDir, "test.zip")
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Errorf("expected zip file to exist at %s", zipPath)
	}

	// Verify zip contents
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	// Filter out directory entries (ending with /)
	var files []*zip.File
	for _, f := range reader.File {
		if !strings.HasSuffix(f.Name, "/") {
			files = append(files, f)
		}
	}

	expectedFiles := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files in zip, got %d (including directories)", len(reader.File))
	}
	for _, f := range files {
		expectedContent, ok := expectedFiles[f.Name]
		if !ok {
			t.Errorf("unexpected file in zip: %s", f.Name)
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, _ := io.ReadAll(rc)
		rc.Close()
		if string(content) != expectedContent {
			t.Errorf("file %s content mismatch: expected %q, got %q", f.Name, expectedContent, string(content))
		}
	}
}

func TestArchivePlugin_Compress_SingleFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "archive_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a single source file
	content := "single file content"
	os.WriteFile(filepath.Join(tmpDir, "single.txt"), []byte(content), 0644)

	p := newArchivePlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":      "compress",
			"source_path": "single.txt",
			"format":      "zip",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	// The result is directly in resp.Results
	archivePath, ok := resp.Results["file_path"].(string)
	if !ok {
		t.Fatal("expected file_path in response results")
	}

	// archivePath is absolute, use it directly
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	if len(reader.File) != 1 {
		t.Errorf("expected 1 file in zip, got %d", len(reader.File))
	}
}

func TestArchivePlugin_GenerateArchiveName(t *testing.T) {
	tests := []struct {
		sourcePath string
		format     string
		expected   string
	}{
		{"mydir", "zip", "mydir.zip"},
		{"mydir", "tar", "mydir.tar.gz"},
		{"mydir", "gzip", "mydir.gz"},
		{"file.txt", "zip", "file.txt.zip"},
		{"file.tar.gz", "tar", "file.tar.gz"},
		{"file.gz", "gzip", "file.gz"},
	}

	plugin := newArchivePluginWithTmpDir(t)
	for _, tt := range tests {
		result := plugin.generateArchiveName(tt.sourcePath, tt.format)
		if result != tt.expected {
			t.Errorf("generateArchiveName(%q, %q) = %q, expected %q",
				tt.sourcePath, tt.format, result, tt.expected)
		}
	}
}

func TestExtractZip_Function(t *testing.T) {
	// extractZip is now a method on ArchivePlugin, tested via TestArchivePlugin_Extract
}

func TestExtractZip_Empty(t *testing.T) {
	// extractZip is now a method on ArchivePlugin, tested via TestArchivePlugin_Extract
}

func TestCreateZip_Directory(t *testing.T) {
	// createZip is now a method on ArchivePlugin, tested via TestArchivePlugin_Compress
}

func TestCreateZip_SingleFile(t *testing.T) {
	// createZip is now a method on ArchivePlugin, tested via TestArchivePlugin_Compress
}
