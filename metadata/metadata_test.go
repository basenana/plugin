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

package metadata

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMetadataPlugin_Name(t *testing.T) {
	p := &MetadataPlugin{}
	if p.Name() != pluginName {
		t.Errorf("expected %s, got %s", pluginName, p.Name())
	}
}

func TestMetadataPlugin_Type(t *testing.T) {
	p := &MetadataPlugin{}
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestMetadataPlugin_Version(t *testing.T) {
	p := &MetadataPlugin{}
	if p.Version() != pluginVersion {
		t.Errorf("expected %s, got %s", pluginVersion, p.Version())
	}
}

func TestMetadataPlugin_Run_File(t *testing.T) {
	p := &MetadataPlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]string{
			"file_path": testFile,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	results := resp.Results
	if results == nil {
		t.Fatal("results should not be nil")
	}
	if results["size"] != int64(len(content)) {
		t.Errorf("expected size %d, got %v", len(content), results["size"])
	}
	if results["is_dir"] != false {
		t.Errorf("expected is_dir false, got %v", results["is_dir"])
	}
	if results["mode"] == "" {
		t.Error("mode should not be empty")
	}
	if results["modified"] == "" {
		t.Error("modified should not be empty")
	}
}

func TestMetadataPlugin_Run_Directory(t *testing.T) {
	p := &MetadataPlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")

	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]string{
			"file_path": testDir,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	results := resp.Results
	if results == nil {
		t.Fatal("results should not be nil")
	}
	if results["is_dir"] != true {
		t.Errorf("expected is_dir true, got %v", results["is_dir"])
	}
}

func TestMetadataPlugin_Run_FileNotFound(t *testing.T) {
	p := &MetadataPlugin{}
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]string{
			"file_path": "/nonexistent/file.txt",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message == "" || resp.Message != "file not found: /nonexistent/file.txt" {
		t.Errorf("expected 'file not found: /nonexistent/file.txt', got '%s'", resp.Message)
	}
}

func TestMetadataPlugin_Run_MissingFilePath(t *testing.T) {
	p := &MetadataPlugin{}
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]string{},
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
