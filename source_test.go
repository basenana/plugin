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

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/utils"
	"go.uber.org/zap"
)

func init() {
	// Initialize logger for tests
	logger.SetLogger(zap.NewNop().Sugar())
}

func newThreeBodyPlugin(workdir string) *ThreeBodyPlugin {
	p := &ThreeBodyPlugin{}
	p.logger = logger.NewPluginLogger(the3BodyPluginName, "test-job")
	p.fileRoot = utils.NewFileAccess(workdir)
	return p
}

func newThreeBodyPluginWithTmpDir(t *testing.T) *ThreeBodyPlugin {
	return newThreeBodyPlugin(t.TempDir())
}

func TestThreeBodyPlugin_Name(t *testing.T) {
	p := newThreeBodyPluginWithTmpDir(t)
	if p.Name() != the3BodyPluginName {
		t.Errorf("expected %s, got %s", the3BodyPluginName, p.Name())
	}
}

func TestThreeBodyPlugin_Type(t *testing.T) {
	p := newThreeBodyPluginWithTmpDir(t)
	if string(p.Type()) != "source" {
		t.Errorf("expected source, got %s", p.Type())
	}
}

func TestThreeBodyPlugin_Version(t *testing.T) {
	p := newThreeBodyPluginWithTmpDir(t)
	if p.Version() != the3BodyPluginVersion {
		t.Errorf("expected %s, got %s", the3BodyPluginVersion, p.Version())
	}
}

func TestThreeBodyPlugin_SourceInfo(t *testing.T) {
	p := newThreeBodyPluginWithTmpDir(t)
	info, err := p.SourceInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info != "internal.FileGenerator" {
		t.Errorf("expected 'internal.FileGenerator', got %s", info)
	}
}

func TestThreeBodyPlugin_Run(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "threebody_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	p := newThreeBodyPlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	filePath, ok := resp.Results["file_path"].(string)
	if !ok {
		t.Fatal("expected file_path in response")
	}

	// Verify file exists
	fullPath := filepath.Join(tmpDir, filePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("expected file to exist at %s", fullPath)
	}

	// Verify file content
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}

	// Should contain timestamp and "Do not answer!"
	if !strings.Contains(string(data), "Do not answer!") {
		t.Errorf("expected 'Do not answer!' in file content, got %s", string(data))
	}

	// Verify size
	size, ok := resp.Results["size"].(int64)
	if !ok {
		t.Fatal("expected size in result")
	}
	if size != int64(len(data)) {
		t.Errorf("expected size %d, got %d", len(data), size)
	}
}

func TestThreeBodyPlugin_MissingWorkingPath(t *testing.T) {
	// When workdir is empty, FileAccess defaults to "."
	// So this test verifies the plugin handles empty workdir gracefully
	p := newThreeBodyPlugin("")
	ctx := context.Background()

	req := &api.Request{}

	_, err := p.Run(ctx, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestThreeBodyPlugin_Run_Multiple(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "threebody_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	p := newThreeBodyPlugin(tmpDir)
	ctx := context.Background()

	// Run multiple times
	for i := 0; i < 3; i++ {
		req := &api.Request{}

		resp, err := p.Run(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !resp.IsSucceed {
			t.Errorf("run %d: expected success, got failure: %s", i, resp.Message)
		}

		filePath, ok := resp.Results["file_path"].(string)
		if !ok {
			t.Fatal("expected file_path in result")
		}

		// Verify file exists
		fullPath := filepath.Join(tmpDir, filePath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("run %d: expected file to exist at %s", i, fullPath)
		}
	}
}

func TestThreeBodyPlugin_FileNaming(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "threebody_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	p := newThreeBodyPlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	filePath, ok := resp.Results["file_path"].(string)
	if !ok {
		t.Fatal("expected file_path in result")
	}

	// File should be named like "3_body_<timestamp>.txt"
	if !strings.HasPrefix(filePath, "3_body_") {
		t.Errorf("expected file name to start with '3_body_', got %s", filePath)
	}
	if !strings.HasSuffix(filePath, ".txt") {
		t.Errorf("expected file name to end with '.txt', got %s", filePath)
	}
}

func TestThreeBodyPlugin_FileContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "threebody_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	p := newThreeBodyPlugin(tmpDir)
	ctx := context.Background()

	req := &api.Request{}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	filePath := resp.Results["file_path"].(string)
	fullPath := filepath.Join(tmpDir, filePath)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}

	// Content should be "<timestamp> - Do not answer!\n"
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "Do not answer!") {
		t.Errorf("expected 'Do not answer!' in content, got %s", lines[0])
	}
}
