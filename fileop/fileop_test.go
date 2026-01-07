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

package fileop

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

func newFileOpPlugin(workdir string) *FileOpPlugin {
	p := &FileOpPlugin{}
	p.logger = logger.NewPluginLogger(pluginName, "test-job")
	p.fileRoot = utils.NewFileAccess(workdir)
	return p
}

func newFileOpPluginWithTmpDir(t *testing.T) *FileOpPlugin {
	return newFileOpPlugin(t.TempDir())
}

func TestFileOpPlugin_Name(t *testing.T) {
	p := newFileOpPluginWithTmpDir(t)
	if p.Name() != pluginName {
		t.Errorf("expected %s, got %s", pluginName, p.Name())
	}
}

func TestFileOpPlugin_Type(t *testing.T) {
	p := newFileOpPluginWithTmpDir(t)
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestFileOpPlugin_Version(t *testing.T) {
	p := newFileOpPluginWithTmpDir(t)
	if p.Version() != pluginVersion {
		t.Errorf("expected %s, got %s", pluginVersion, p.Version())
	}
}

func TestFileOpPlugin_Run_Copy(t *testing.T) {
	tmpDir := t.TempDir()
	p := newFileOpPlugin(tmpDir)
	ctx := context.Background()

	srcFile := filepath.Join(tmpDir, "src.txt")
	destFile := filepath.Join(tmpDir, "dest.txt")

	err := os.WriteFile(srcFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"action": "cp",
			"src":    srcFile,
			"dest":   destFile,
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
	if string(content) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(content))
	}
}

func TestFileOpPlugin_Run_Move(t *testing.T) {
	tmpDir := t.TempDir()
	p := newFileOpPlugin(tmpDir)
	ctx := context.Background()

	srcFile := filepath.Join(tmpDir, "src.txt")
	destFile := filepath.Join(tmpDir, "dest.txt")

	err := os.WriteFile(srcFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"action": "mv",
			"src":    srcFile,
			"dest":   destFile,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	if _, err = os.Stat(srcFile); !os.IsNotExist(err) {
		t.Errorf("source file should not exist after move")
	}

	if _, err = os.Stat(destFile); err != nil {
		t.Errorf("dest file should exist: %v", err)
	}

	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(content))
	}
}

func TestFileOpPlugin_Run_Remove(t *testing.T) {
	tmpDir := t.TempDir()
	p := newFileOpPlugin(tmpDir)
	ctx := context.Background()

	srcFile := filepath.Join(tmpDir, "to_delete.txt")

	err := os.WriteFile(srcFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"action": "rm",
			"src":    srcFile,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	if _, err = os.Stat(srcFile); !os.IsNotExist(err) {
		t.Errorf("file should not exist after delete")
	}
}

func TestFileOpPlugin_Run_Rename(t *testing.T) {
	tmpDir := t.TempDir()
	p := newFileOpPlugin(tmpDir)
	ctx := context.Background()

	srcFile := filepath.Join(tmpDir, "old_name.txt")
	destFile := filepath.Join(tmpDir, "new_name.txt")

	err := os.WriteFile(srcFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"action": "rename",
			"src":    srcFile,
			"dest":   destFile,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	if _, err = os.Stat(srcFile); !os.IsNotExist(err) {
		t.Errorf("old file should not exist after rename")
	}

	if _, err = os.Stat(destFile); err != nil {
		t.Errorf("new file should exist: %v", err)
	}
}

func TestFileOpPlugin_Run_MissingAction(t *testing.T) {
	p := newFileOpPluginWithTmpDir(t)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"src": "somefile.txt",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message == "" || resp.Message != "action is required" {
		t.Errorf("expected 'action is required', got '%s'", resp.Message)
	}
}

func TestFileOpPlugin_Run_UnknownAction(t *testing.T) {
	p := newFileOpPluginWithTmpDir(t)
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action": "unknown",
			"src":    "somefile.txt",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message == "" || resp.Message != "unknown action: unknown" {
		t.Errorf("expected 'unknown action: unknown', got '%s'", resp.Message)
	}
}

func TestResolvePath(t *testing.T) {
	// ResolvePath function has been moved to utils/file.go as FileAccess method
	// Tests are now in utils/file_test.go
}
