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
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/basenana/plugin/api"
)

func TestArchivePlugin_Name(t *testing.T) {
	p := &ArchivePlugin{}
	if p.Name() != pluginName {
		t.Errorf("expected %s, got %s", pluginName, p.Name())
	}
}

func TestArchivePlugin_Type(t *testing.T) {
	p := &ArchivePlugin{}
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestArchivePlugin_Version(t *testing.T) {
	p := &ArchivePlugin{}
	if p.Version() != pluginVersion {
		t.Errorf("expected %s, got %s", pluginVersion, p.Version())
	}
}

func TestArchivePlugin_Run_Zip(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "test.zip")
	extractDir := filepath.Join(tmpDir, "extracted")

	w, err := os.Create(zipFile)
	if err != nil {
		t.Fatal(err)
	}

	zipWriter := zip.NewWriter(w)
	testContent := []byte("hello world")

	f, err := zipWriter.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write(testContent)
	if err != nil {
		t.Fatal(err)
	}

	f2, err := zipWriter.Create("subdir/nested.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f2.Write([]byte("nested content"))
	if err != nil {
		t.Fatal(err)
	}

	err = zipWriter.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": zipFile,
			"format":    "zip",
			"dest_path": extractDir,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content, err := os.ReadFile(filepath.Join(extractDir, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(content))
	}

	content, err = os.ReadFile(filepath.Join(extractDir, "subdir/nested.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "nested content" {
		t.Errorf("expected 'nested content', got '%s'", string(content))
	}
}

func TestArchivePlugin_Run_MissingFilePath(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
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

func TestArchivePlugin_Run_MissingFormat(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "test.zip")
	_, err := os.Create(zipFile)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": zipFile,
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

func TestArchivePlugin_Run_UnsupportedFormat(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.7z")
	_, err := os.Create(testFile)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": testFile,
			"format":    "7z",
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

func TestExtractGzip(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("gzip test content")
	gzipFile := filepath.Join(tmpDir, "test.gz")
	w, err := os.Create(gzipFile)
	if err != nil {
		t.Fatal(err)
	}

	gzipWriter := gzip.NewWriter(w)
	_, err = gzipWriter.Write(content)
	if err != nil {
		t.Fatal(err)
	}
	err = gzipWriter.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	err = extractGzip(gzipFile, extractDir)
	if err != nil {
		t.Fatal(err)
	}

	resultFile := filepath.Join(extractDir, "test")
	resultContent, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(resultContent) != string(content) {
		t.Errorf("expected '%s', got '%s'", string(content), string(resultContent))
	}
}

func TestExtractGzipTarExtension(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("tgz test content")
	tgzFile := filepath.Join(tmpDir, "test.tgz")
	w, err := os.Create(tgzFile)
	if err != nil {
		t.Fatal(err)
	}

	gzipWriter := gzip.NewWriter(w)
	_, err = gzipWriter.Write(content)
	if err != nil {
		t.Fatal(err)
	}
	err = gzipWriter.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	err = extractGzip(tgzFile, extractDir)
	if err != nil {
		t.Fatal(err)
	}

	resultFile := filepath.Join(extractDir, "test.tar")
	resultContent, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(resultContent) != string(content) {
		t.Errorf("expected '%s', got '%s'", string(content), string(resultContent))
	}
}

func TestExtractTar(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	extractDir := filepath.Join(tmpDir, "extracted")

	buf := &bytes.Buffer{}
	gw := gzip.NewWriter(buf)
	tw := tar.NewWriter(gw)

	testContent := []byte("tar test content")
	hdr := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: int64(len(testContent)),
	}
	err := tw.WriteHeader(hdr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tw.Write(testContent)
	if err != nil {
		t.Fatal(err)
	}
	err = tw.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = gw.Close()
	if err != nil {
		t.Fatal(err)
	}

	tarFile := filepath.Join(tmpDir, "test.tar.gz")
	err = os.WriteFile(tarFile, buf.Bytes(), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": tarFile,
			"format":    "tar",
			"dest_path": extractDir,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content, err := os.ReadFile(filepath.Join(extractDir, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "tar test content" {
		t.Errorf("expected 'tar test content', got '%s'", string(content))
	}
}

// Compression tests

func TestArchivePlugin_Compress_Zip(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()

	// Create source file
	sourceFile := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("compress test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	archiveDir := filepath.Join(tmpDir, "archives")
	archivePath := filepath.Join(archiveDir, "output.zip")

	req := &api.Request{
		Parameter: map[string]any{
			"action":       "compress",
			"source_path":  sourceFile,
			"format":       "zip",
			"dest_path":    archiveDir,
			"archive_name": "output.zip",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	// Verify zip file exists
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatal(err)
	}

	// Extract and verify content
	extractDir := filepath.Join(tmpDir, "extracted")
	req = &api.Request{
		Parameter: map[string]any{
			"file_path": archivePath,
			"format":    "zip",
			"dest_path": extractDir,
		},
	}

	resp, err = p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content, err := os.ReadFile(filepath.Join(extractDir, "source.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "compress test content" {
		t.Errorf("expected 'compress test content', got '%s'", string(content))
	}
}

func TestArchivePlugin_Compress_Tar(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()

	// Create source file
	sourceFile := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("tar compress test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	archiveDir := filepath.Join(tmpDir, "archives")
	archivePath := filepath.Join(archiveDir, "output.tar.gz")

	req := &api.Request{
		Parameter: map[string]any{
			"action":       "compress",
			"source_path":  sourceFile,
			"format":       "tar",
			"dest_path":    archiveDir,
			"archive_name": "output.tar.gz",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	// Verify tar.gz file exists
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatal(err)
	}

	// Extract and verify content
	extractDir := filepath.Join(tmpDir, "extracted")
	req = &api.Request{
		Parameter: map[string]any{
			"file_path": archivePath,
			"format":    "tar",
			"dest_path": extractDir,
		},
	}

	resp, err = p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content, err := os.ReadFile(filepath.Join(extractDir, "source.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "tar compress test" {
		t.Errorf("expected 'tar compress test', got '%s'", string(content))
	}
}

func TestArchivePlugin_Compress_Gzip(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()

	// Create source file
	sourceFile := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("gzip compress test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	archiveDir := filepath.Join(tmpDir, "archives")
	archivePath := filepath.Join(archiveDir, "output.gz")

	req := &api.Request{
		Parameter: map[string]any{
			"action":       "compress",
			"source_path":  sourceFile,
			"format":       "gzip",
			"dest_path":    archiveDir,
			"archive_name": "output.gz",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	// Verify gzip file exists
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatal(err)
	}

	// Extract and verify content
	extractDir := filepath.Join(tmpDir, "extracted")
	req = &api.Request{
		Parameter: map[string]any{
			"file_path": archivePath,
			"format":    "gzip",
			"dest_path": extractDir,
		},
	}

	resp, err = p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content, err := os.ReadFile(filepath.Join(extractDir, "output"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "gzip compress test" {
		t.Errorf("expected 'gzip compress test', got '%s'", string(content))
	}
}

func TestArchivePlugin_Compress_Directory(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()

	// Create source directory with files
	sourceDir := filepath.Join(tmpDir, "mydir")
	err := os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("content 1"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(sourceDir, "file2.txt"), []byte("content 2"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create subdirectory
	subDir := filepath.Join(sourceDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	archiveDir := filepath.Join(tmpDir, "archives")
	archivePath := filepath.Join(archiveDir, "dir.zip")

	req := &api.Request{
		Parameter: map[string]any{
			"action":       "compress",
			"source_path":  sourceDir,
			"format":       "zip",
			"dest_path":    archiveDir,
			"archive_name": "dir.zip",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	// Verify zip file exists
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatal(err)
	}

	// Extract and verify content
	extractDir := filepath.Join(tmpDir, "extracted")
	req = &api.Request{
		Parameter: map[string]any{
			"file_path": archivePath,
			"format":    "zip",
			"dest_path": extractDir,
		},
	}

	resp, err = p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content1, err := os.ReadFile(filepath.Join(extractDir, "mydir", "file1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content1) != "content 1" {
		t.Errorf("expected 'content 1', got '%s'", string(content1))
	}

	content2, err := os.ReadFile(filepath.Join(extractDir, "mydir", "file2.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content2) != "content 2" {
		t.Errorf("expected 'content 2', got '%s'", string(content2))
	}

	nested, err := os.ReadFile(filepath.Join(extractDir, "mydir", "subdir", "nested.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(nested) != "nested content" {
		t.Errorf("expected 'nested content', got '%s'", string(nested))
	}
}

func TestArchivePlugin_Compress_MissingSourcePath(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":  "compress",
			"format":  "zip",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message != "source_path is required for compression" {
		t.Errorf("expected 'source_path is required for compression', got '%s'", resp.Message)
	}
}

func TestArchivePlugin_Compress_DirectoryGzip(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()

	// Create source directory
	sourceDir := filepath.Join(tmpDir, "mydir")
	err := os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	archiveDir := filepath.Join(tmpDir, "archives")

	req := &api.Request{
		Parameter: map[string]any{
			"action":      "compress",
			"source_path": sourceDir,
			"format":      "gzip",
			"dest_path":   archiveDir,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure for gzip on directory, got success")
	}
}

func TestArchivePlugin_Compress_ReturnsResult(t *testing.T) {
	p := &ArchivePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()

	sourceFile := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"action":      "compress",
			"source_path": sourceFile,
			"format":      "zip",
			"dest_path":   tmpDir,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	if resp.Results == nil {
		t.Fatal("expected results, got nil")
	}
	if resp.Results["file_path"] == nil {
		t.Error("expected file_path in results")
	}
	if resp.Results["size"] == nil {
		t.Error("expected size in results")
	}
}

func TestGenerateArchiveName(t *testing.T) {
	tests := []struct {
		source   string
		format   string
		expected string
	}{
		{"file.txt", "zip", "file.txt.zip"},
		{"file.txt", "tar", "file.txt.tar.gz"},
		{"file.txt", "gzip", "file.txt.gz"},
		{"archive.zip", "zip", "archive.zip"},
		{"archive.tar.gz", "tar", "archive.tar.gz"},
		{"archive.gz", "gzip", "archive.gz"},
	}

	for _, tt := range tests {
		result := generateArchiveName(tt.source, tt.format)
		if result != tt.expected {
			t.Errorf("generateArchiveName(%s, %s) = %s, expected %s", tt.source, tt.format, result, tt.expected)
		}
	}
}
