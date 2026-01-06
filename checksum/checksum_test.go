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

package checksum

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.SetLogger(zap.NewNop().Sugar())
	os.Exit(m.Run())
}

func newChecksumPlugin(algorithm string) *ChecksumPlugin {
	p := &ChecksumPlugin{algorithm: algorithm}
	p.logger = logger.NewPluginLogger(pluginName, "test-job")
	return p
}

func TestChecksumPlugin_Name(t *testing.T) {
	p := &ChecksumPlugin{}
	if p.Name() != pluginName {
		t.Errorf("expected %s, got %s", pluginName, p.Name())
	}
}

func TestChecksumPlugin_Type(t *testing.T) {
	p := &ChecksumPlugin{}
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestChecksumPlugin_Version(t *testing.T) {
	p := &ChecksumPlugin{}
	if p.Version() != pluginVersion {
		t.Errorf("expected %s, got %s", pluginVersion, p.Version())
	}
}

func TestChecksumPlugin_MD5(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checksum_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	content := "hello world"
	filePath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	// Calculate expected MD5
	hash := md5.Sum([]byte(content))
	expected := hex.EncodeToString(hash[:16])

	p := newChecksumPlugin("md5")
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": filePath,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	result, ok := resp.Results["hash"].(string)
	if !ok {
		t.Fatal("expected hash in results")
	}
	if result != expected {
		t.Errorf("expected hash %s, got %s", expected, result)
	}
}

func TestChecksumPlugin_SHA256(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checksum_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	content := "hello world"
	filePath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	// Calculate expected SHA256 using sha256.New().Sum(nil)
	h := sha256.New()
	h.Write([]byte(content))
	expected := hex.EncodeToString(h.Sum(nil))

	p := newChecksumPlugin("sha256")
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": filePath,
			"algorithm": "sha256",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	result, ok := resp.Results["hash"].(string)
	if !ok {
		t.Fatal("expected hash in results")
	}
	if result != expected {
		t.Errorf("expected hash %s, got %s", expected, result)
	}
}

func TestChecksumPlugin_MissingFilePath(t *testing.T) {
	p := newChecksumPlugin("md5")
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

func TestChecksumPlugin_FileNotFound(t *testing.T) {
	p := newChecksumPlugin("md5")
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
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
}

func TestChecksumPlugin_InvalidAlgorithm(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checksum_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(filePath, []byte("content"), 0644)

	p := newChecksumPlugin("sha512")
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": filePath,
			"algorithm": "sha512",
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

func TestChecksumPlugin_EmptyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checksum_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "empty.txt")
	os.WriteFile(filePath, []byte(""), 0644)

	p := newChecksumPlugin("md5")
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": filePath,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	result, ok := resp.Results["hash"].(string)
	if !ok {
		t.Fatal("expected hash in results")
	}
	// MD5 of empty string
	if result != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Errorf("expected d41d8cd98f00b204e9800998ecf8427e, got %s", result)
	}
}

func TestChecksumPlugin_LargeFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "checksum_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a 1MB file
	filePath := filepath.Join(tmpDir, "large.txt")
	content := make([]byte, 1024*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}
	os.WriteFile(filePath, content, 0644)

	p := newChecksumPlugin("md5")
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"file_path": filePath,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	result, ok := resp.Results["hash"].(string)
	if !ok {
		t.Fatal("expected hash in results")
	}
	if len(result) != 32 {
		t.Errorf("expected 32 character hash, got %d", len(result))
	}
}

func TestComputeHash_MD5(t *testing.T) {
	content := []byte("test content")
	filePath := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(filePath, content, 0644)

	hash, err := computeHash(filePath, "md5")
	if err != nil {
		t.Fatal(err)
	}

	md5Hash := md5.Sum(content)
	expected := hex.EncodeToString(md5Hash[:16])
	if hash != expected {
		t.Errorf("expected %s, got %s", expected, hash)
	}
}

func TestComputeHash_SHA256(t *testing.T) {
	content := []byte("test content")
	filePath := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(filePath, content, 0644)

	hash, err := computeHash(filePath, "sha256")
	if err != nil {
		t.Fatal(err)
	}

	h := sha256.New()
	h.Write(content)
	sha256Hash := h.Sum(nil)
	expected := hex.EncodeToString(sha256Hash)
	if hash != expected {
		t.Errorf("expected %s, got %s", expected, hash)
	}
}

func TestComputeHash_InvalidAlgorithm(t *testing.T) {
	content := []byte("test content")
	filePath := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(filePath, content, 0644)

	_, err := computeHash(filePath, "invalid")
	if err == nil {
		t.Error("expected error for invalid algorithm")
	}
}
