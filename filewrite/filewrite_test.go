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

package filewrite

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/utils"
	"go.uber.org/zap"
)

func init() {
	logger.SetLogger(zap.NewNop().Sugar())
}

func newFileWritePlugin(workdir string) *FileWritePlugin {
	p := &FileWritePlugin{}
	p.logger = logger.NewPluginLogger(pluginName, "test-job")
	p.fileRoot = utils.NewFileAccess(workdir)
	return p
}

func newFileWritePluginWithTmpDir(t *testing.T) *FileWritePlugin {
	return newFileWritePlugin(t.TempDir())
}

func TestFileWritePlugin_Name(t *testing.T) {
	p := newFileWritePluginWithTmpDir(t)
	if p.Name() != pluginName {
		t.Errorf("expected %s, got %s", pluginName, p.Name())
	}
}

func TestFileWritePlugin_Type(t *testing.T) {
	p := newFileWritePluginWithTmpDir(t)
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestFileWritePlugin_Version(t *testing.T) {
	p := newFileWritePluginWithTmpDir(t)
	if p.Version() != pluginVersion {
		t.Errorf("expected %s, got %s", pluginVersion, p.Version())
	}
}

func TestFileWritePlugin_Run_SaveContent(t *testing.T) {
	tmpDir := t.TempDir()
	p := newFileWritePlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"content":   "hello world",
			"dest_path": "test.txt",
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

	content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(content))
	}
}

func TestFileWritePlugin_Run_DefaultMode(t *testing.T) {
	tmpDir := t.TempDir()
	p := newFileWritePlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"content":   "hello world",
			"dest_path": "test.txt",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(content))
	}
}

func TestFileWritePlugin_Run_CreateParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	p := newFileWritePlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"content":   "nested content",
			"dest_path": "subdir/nested/test.txt",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "subdir/nested/test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "nested content" {
		t.Errorf("expected 'nested content', got '%s'", string(content))
	}
}

func TestFileWritePlugin_Run_MissingDestPath(t *testing.T) {
	p := newFileWritePluginWithTmpDir(t)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
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

func TestFileWritePlugin_Run_InvalidMode(t *testing.T) {
	tmpDir := t.TempDir()
	p := newFileWritePlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"content":   "hello world",
			"dest_path": "test.txt",
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

func TestFileWritePlugin_Run_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	p := newFileWritePlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"content":   "",
			"dest_path": "empty.txt",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	info, err := os.Stat(filepath.Join(tmpDir, "empty.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Errorf("expected size 0, got %d", info.Size())
	}
}

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		workingDir  string
		want        string
		wantErr     bool
	}{
		{
			name:        "absolute path outside workdir",
			path:        "/absolute/path/file.txt",
			workingDir:  "/working",
			wantErr:     true,
		},
		{
			name:        "relative path",
			path:        "file.txt",
			workingDir:  "/working",
			want:        "/working/file.txt",
			wantErr:     false,
		},
		{
			name:        "absolute path within workdir",
			path:        "/working/dir/file.txt",
			workingDir:  "/working",
			want:        "/working/dir/file.txt",
			wantErr:     false,
		},
		{
			name:        "path traversal rejected",
			path:        "../outside.txt",
			workingDir:  "/working",
			wantErr:     true,
		},
		{
			name:        "empty path rejected",
			path:        "",
			workingDir:  "/working",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fa := utils.NewFileAccess(tt.workingDir)
			got, err := fa.GetAbsPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAbsPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetAbsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	// SanitizePath function has been moved to utils/file.go as FileAccess.ValidatePath
	// Tests are now in utils/file_test.go
}
