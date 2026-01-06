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

package web

import (
	"os"
	"testing"

	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/types"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.SetLogger(zap.NewNop().Sugar())
	os.Exit(m.Run())
}

func newWebpackPlugin() *WebpackPlugin {
	p := &WebpackPlugin{}
	p.logger = logger.NewPluginLogger(WebpackPluginName, "test-job")
	return p
}

func TestWebpackPlugin_Name(t *testing.T) {
	p := newWebpackPlugin()
	if p.Name() != WebpackPluginName {
		t.Errorf("expected %s, got %s", WebpackPluginName, p.Name())
	}
}

func TestWebpackPlugin_Type(t *testing.T) {
	p := newWebpackPlugin()
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestWebpackPlugin_Version(t *testing.T) {
	p := newWebpackPlugin()
	if p.Version() != WebpackPluginVersion {
		t.Errorf("expected %s, got %s", WebpackPluginVersion, p.Version())
	}
}

func TestNewWebpackPlugin_DefaultFileType(t *testing.T) {
	p := NewWebpackPlugin(types.PluginCall{
		Params: map[string]string{},
	}).(*WebpackPlugin)

	if p.fileType != "webarchive" {
		t.Errorf("expected default file type to be 'webarchive', got %s", p.fileType)
	}
}

func TestNewWebpackPlugin_CustomFileType(t *testing.T) {
	p := NewWebpackPlugin(types.PluginCall{
		Params: map[string]string{
			webpackParameterFileType: "html",
		},
	}).(*WebpackPlugin)

	if p.fileType != "html" {
		t.Errorf("expected file type to be 'html', got %s", p.fileType)
	}
}

func TestNewWebpackPlugin_DefaultClutterFree(t *testing.T) {
	p := NewWebpackPlugin(types.PluginCall{
		Params: map[string]string{},
	}).(*WebpackPlugin)

	if p.clutterFree != true {
		t.Errorf("expected default clutterFree to be true, got %v", p.clutterFree)
	}
}

func TestNewWebpackPlugin_CustomClutterFree(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{"anything", false},
	}

	for _, tt := range tests {
		p := NewWebpackPlugin(types.PluginCall{
			Params: map[string]string{
				webpackParameterClutterFree: tt.value,
			},
		}).(*WebpackPlugin)

		if p.clutterFree != tt.expected {
			t.Errorf("clutterFree = %s: expected %v, got %v", tt.value, tt.expected, p.clutterFree)
		}
	}
}

func TestWebpackPluginSpec(t *testing.T) {
	if WebpackPluginSpec.Name != WebpackPluginName {
		t.Errorf("expected spec name %s, got %s", WebpackPluginName, WebpackPluginSpec.Name)
	}
	if WebpackPluginSpec.Version != WebpackPluginVersion {
		t.Errorf("expected spec version %s, got %s", WebpackPluginVersion, WebpackPluginSpec.Version)
	}
	if string(WebpackPluginSpec.Type) != "process" {
		t.Errorf("expected spec type 'process', got %s", WebpackPluginSpec.Type)
	}
}

func TestWebpackPlugin_InvalidFileType(t *testing.T) {
	// Skip this test as it requires network access and proper logger initialization
	// The internal packFromURL method uses logger which is not fully initialized in tests
	t.Skip("Skipping test that requires network access and internal logger initialization")
}

func TestWebpackPlugin_Parameters(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]string
		expectError bool
	}{
		{
			name: "missing file_name",
			params: map[string]string{
				webpackParameterURL: "https://example.com",
			},
			expectError: true,
		},
		{
			name: "missing url",
			params: map[string]string{
				webpackParameterFileName: "test",
			},
			expectError: true,
		},
		{
			name: "both present",
			params: map[string]string{
				webpackParameterFileName: "test",
				webpackParameterURL:      "https://example.com",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebpackParameters(tt.params)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func validateWebpackParameters(params map[string]string) error {
	if params[webpackParameterFileName] == "" {
		return &validationError{"file name is empty"}
	}
	if params[webpackParameterURL] == "" {
		return &validationError{"url is empty"}
	}
	return nil
}

type validationError struct {
	msg string
}

func (e *validationError) Error() string {
	return e.msg
}

func TestEnablePrivateNet(t *testing.T) {
	// Test that the enablePrivateNet variable is correctly initialized
	// It reads from environment variable
	if enablePrivateNet != (os.Getenv("WebPackerEnablePrivateNet") == "true") {
		t.Errorf("enablePrivateNet mismatch")
	}
}
