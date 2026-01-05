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

package save

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/basenana/plugin/api"
)

func TestSavePlugin_Name(t *testing.T) {
	p := &SavePlugin{}
	if p.Name() != pluginName {
		t.Errorf("expected %s, got %s", pluginName, p.Name())
	}
}

func TestSavePlugin_Type(t *testing.T) {
	p := &SavePlugin{}
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestSavePlugin_Version(t *testing.T) {
	p := &SavePlugin{}
	if p.Version() != pluginVersion {
		t.Errorf("expected %s, got %s", pluginVersion, p.Version())
	}
}

func TestSavePlugin_Run_SaveContent(t *testing.T) {
	p := &SavePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	destFile := filepath.Join(tmpDir, "test.txt")

	req := &api.Request{
		Parameter: map[string]string{
			"content":   "hello world",
			"dest_path": destFile,
			"mode":      "0644",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(content))
	}
}

func TestSavePlugin_Run_DefaultMode(t *testing.T) {
	p := &SavePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	destFile := filepath.Join(tmpDir, "test.txt")

	req := &api.Request{
		Parameter: map[string]string{
			"content":   "hello world",
			"dest_path": destFile,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(content))
	}
}

func TestSavePlugin_Run_CreateParentDir(t *testing.T) {
	p := &SavePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	destFile := filepath.Join(tmpDir, "subdir", "nested", "test.txt")

	req := &api.Request{
		Parameter: map[string]string{
			"content":   "nested content",
			"dest_path": destFile,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "nested content" {
		t.Errorf("expected 'nested content', got '%s'", string(content))
	}
}

func TestSavePlugin_Run_MissingDestPath(t *testing.T) {
	p := &SavePlugin{}
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]string{
			"content": "hello world",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message == "" || resp.Message != "dest_path is required" {
		t.Errorf("expected 'dest_path is required', got '%s'", resp.Message)
	}
}

func TestSavePlugin_Run_InvalidMode(t *testing.T) {
	p := &SavePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	destFile := filepath.Join(tmpDir, "test.txt")

	req := &api.Request{
		Parameter: map[string]string{
			"content":   "hello world",
			"dest_path": destFile,
			"mode":      "invalid",
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

func TestSavePlugin_Run_EmptyContent(t *testing.T) {
	p := &SavePlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	destFile := filepath.Join(tmpDir, "empty.txt")

	req := &api.Request{
		Parameter: map[string]string{
			"content":   "",
			"dest_path": destFile,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	info, err := os.Stat(destFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Errorf("expected size 0, got %d", info.Size())
	}
}

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		workingDir string
		want       string
	}{
		{
			name:       "absolute path",
			path:       "/absolute/path/file.txt",
			workingDir: "/working",
			want:       "/absolute/path/file.txt",
		},
		{
			name:       "relative path",
			path:       "file.txt",
			workingDir: "/working",
			want:       "/working/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolvePath(tt.path, tt.workingDir)
			if err != nil {
				t.Errorf("ResolvePath() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ResolvePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "normal path",
			path:    "foo/bar.txt",
			want:    "foo/bar.txt",
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			path:    "../etc/passwd",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
