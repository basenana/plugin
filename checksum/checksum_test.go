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
)

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

func TestChecksumPlugin_Run_MD5(t *testing.T) {
	p := &ChecksumPlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("hello world")
	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatal(err)
	}

	md5Sum := md5.Sum(content)
	expectedMD5 := hex.EncodeToString(md5Sum[:])

	req := &api.Request{
		Parameter: map[string]string{
			"file_path": testFile,
			"algorithm": "md5",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if resp.Results["hash"] != expectedMD5 {
		t.Errorf("expected %s, got %v", expectedMD5, resp.Results["hash"])
	}
}

func TestChecksumPlugin_Run_SHA256(t *testing.T) {
	p := &ChecksumPlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("hello world")
	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatal(err)
	}

	sha256Sum := sha256.Sum(content)
	expectedSHA256 := hex.EncodeToString(sha256Sum[:])

	req := &api.Request{
		Parameter: map[string]string{
			"file_path": testFile,
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
	if resp.Results["hash"] != expectedSHA256 {
		t.Errorf("expected %s, got %v", expectedSHA256, resp.Results["hash"])
	}
}

func TestChecksumPlugin_Run_DefaultAlgorithm(t *testing.T) {
	p := &ChecksumPlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("hello world")
	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatal(err)
	}

	md5Sum := md5.Sum(content)
	expectedMD5 := hex.EncodeToString(md5Sum[:])

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
	if resp.Results["hash"] != expectedMD5 {
		t.Errorf("expected %s, got %v", expectedMD5, resp.Results["hash"])
	}
}

func TestChecksumPlugin_Run_FileNotFound(t *testing.T) {
	p := &ChecksumPlugin{}
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]string{
			"file_path": "/nonexistent/file.txt",
			"algorithm": "md5",
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

func TestChecksumPlugin_Run_MissingFilePath(t *testing.T) {
	p := &ChecksumPlugin{}
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]string{
			"algorithm": "md5",
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

func TestChecksumPlugin_Run_UnsupportedAlgorithm(t *testing.T) {
	p := &ChecksumPlugin{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	err := os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &api.Request{
		Parameter: map[string]string{
			"file_path": testFile,
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
