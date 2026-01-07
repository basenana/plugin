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
	"testing"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/types"
	"go.uber.org/zap"
)

func init() {
	logger.SetLogger(zap.NewNop().Sugar())
}

func newMetadataPlugin(t *testing.T, workdir string) *MetadataPlugin {
	return NewMetadataPlugin(types.PluginCall{
		JobID:       "test-job",
		Workflow:    "test-workflow",
		Namespace:   "test-namespace",
		WorkingPath: workdir,
		PluginName:  "",
		Version:     "",
		Params:      map[string]string{},
	}).(*MetadataPlugin)
}

func TestMetadataPlugin_Name(t *testing.T) {
	p := newMetadataPlugin(t, t.TempDir())
	if p.Name() != pluginName {
		t.Errorf("expected %s, got %s", pluginName, p.Name())
	}
}

func TestMetadataPlugin_Type(t *testing.T) {
	p := newMetadataPlugin(t, t.TempDir())
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestMetadataPlugin_Version(t *testing.T) {
	p := newMetadataPlugin(t, t.TempDir())
	if p.Version() != pluginVersion {
		t.Errorf("expected %s, got %s", pluginVersion, p.Version())
	}
}

func TestMetadataPlugin_Run_File(t *testing.T) {
	workdir := t.TempDir()
	p := newMetadataPlugin(t, workdir)
	ctx := context.Background()

	content := []byte("test content")
	err := p.fileRoot.Write("test.txt", content, 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": "test.txt",
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
	workdir := t.TempDir()
	p := newMetadataPlugin(t, workdir)
	ctx := context.Background()

	err := p.fileRoot.MkdirAll("testdir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": "testdir",
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
	p := newMetadataPlugin(t, t.TempDir())
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": "nonexistent/file.txt",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message == "" {
		t.Error("expected error message")
	}
}

func TestMetadataPlugin_Run_MissingFilePath(t *testing.T) {
	p := newMetadataPlugin(t, t.TempDir())
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{},
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
