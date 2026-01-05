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
		Parameter: map[string]string{
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
		Parameter: map[string]string{
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
		Parameter: map[string]string{
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
		Parameter: map[string]string{
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
		Parameter: map[string]string{
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
