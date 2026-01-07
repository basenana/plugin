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
	"testing"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/types"
	"go.uber.org/zap"
)

func init() {
	logger.SetLogger(zap.NewNop().Sugar())
}

func newFileOpPlugin(t *testing.T, workdir string) *FileOpPlugin {
	return NewFileOpPlugin(types.PluginCall{
		JobID:       "test-job",
		Workflow:    "test-workflow",
		Namespace:   "test-namespace",
		WorkingPath: workdir,
		PluginName:  "",
		Version:     "",
		Params:      map[string]string{},
	}).(*FileOpPlugin)
}

func TestFileOpPlugin_Name(t *testing.T) {
	p := newFileOpPlugin(t, t.TempDir())
	if p.Name() != pluginName {
		t.Errorf("expected %s, got %s", pluginName, p.Name())
	}
}

func TestFileOpPlugin_Type(t *testing.T) {
	p := newFileOpPlugin(t, t.TempDir())
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestFileOpPlugin_Version(t *testing.T) {
	p := newFileOpPlugin(t, t.TempDir())
	if p.Version() != pluginVersion {
		t.Errorf("expected %s, got %s", pluginVersion, p.Version())
	}
}

func TestFileOpPlugin_Run_Copy(t *testing.T) {
	workdir := t.TempDir()
	p := newFileOpPlugin(t, workdir)
	ctx := context.Background()

	// create source file using FileAccess
	err := p.fileRoot.Write("src.txt", []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"action": "cp",
			"src":    "src.txt",
			"dest":   "dest.txt",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	content, err := p.fileRoot.Read("dest.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(content))
	}
}

func TestFileOpPlugin_Run_Move(t *testing.T) {
	workdir := t.TempDir()
	p := newFileOpPlugin(t, workdir)
	ctx := context.Background()

	err := p.fileRoot.Write("src.txt", []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"action": "mv",
			"src":    "src.txt",
			"dest":   "dest.txt",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	if p.fileRoot.Exists("src.txt") {
		t.Error("source file should not exist after move")
	}

	if !p.fileRoot.Exists("dest.txt") {
		t.Error("dest file should exist after move")
	}

	content, err := p.fileRoot.Read("dest.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(content))
	}
}

func TestFileOpPlugin_Run_Remove(t *testing.T) {
	workdir := t.TempDir()
	p := newFileOpPlugin(t, workdir)
	ctx := context.Background()

	err := p.fileRoot.Write("to_delete.txt", []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"action": "rm",
			"src":    "to_delete.txt",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	if p.fileRoot.Exists("to_delete.txt") {
		t.Error("file should not exist after delete")
	}
}

func TestFileOpPlugin_Run_Rename(t *testing.T) {
	workdir := t.TempDir()
	p := newFileOpPlugin(t, workdir)
	ctx := context.Background()

	err := p.fileRoot.Write("old_name.txt", []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"action": "rename",
			"src":    "old_name.txt",
			"dest":   "new_name.txt",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	if p.fileRoot.Exists("old_name.txt") {
		t.Error("old file should not exist after rename")
	}

	if !p.fileRoot.Exists("new_name.txt") {
		t.Error("new file should exist after rename")
	}
}

func TestFileOpPlugin_Run_MissingAction(t *testing.T) {
	p := newFileOpPlugin(t, t.TempDir())
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
	p := newFileOpPlugin(t, t.TempDir())
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
