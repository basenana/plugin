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

package text

import (
	"context"
	"testing"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/types"
	"github.com/basenana/plugin/utils"
	"go.uber.org/zap"
)

func init() {
	logger.SetLogger(zap.NewNop().Sugar())
}

type testContext struct {
	workdir string
	fa      *utils.FileAccess
}

func newTestContext(t *testing.T) *testContext {
	workdir := t.TempDir()
	return &testContext{
		workdir: workdir,
		fa:      utils.NewFileAccess(workdir),
	}
}

func (tc *testContext) newPlugin() *TextPlugin {
	return NewTextPlugin(types.PluginCall{
		JobID:       "test-job",
		Workflow:    "test-workflow",
		Namespace:   "test-namespace",
		WorkingPath: tc.workdir,
		PluginName:  "",
		Version:     "",
		Params:      map[string]string{},
	}).(*TextPlugin)
}

func TestTextPlugin_Name(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	if p.Name() != pluginName {
		t.Errorf("expected %s, got %s", pluginName, p.Name())
	}
}

func TestTextPlugin_Type(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestTextPlugin_Version(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	if p.Version() != pluginVersion {
		t.Errorf("expected %s, got %s", pluginVersion, p.Version())
	}
}

func TestTextPlugin_Run_Search_Found(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":  "search",
			"content": "hello world",
			"pattern": "world",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if resp.Results["result"] != true {
		t.Errorf("expected true, got %v", resp.Results["result"])
	}
}

func TestTextPlugin_Run_Search_NotFound(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":  "search",
			"content": "hello world",
			"pattern": "foo",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if resp.Results["result"] != false {
		t.Errorf("expected false, got %v", resp.Results["result"])
	}
}

func TestTextPlugin_Run_Replace(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":      "replace",
			"content":     "hello world",
			"pattern":     "world",
			"replacement": "go",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if resp.Results["result"] != "hello go" {
		t.Errorf("expected 'hello go', got '%v'", resp.Results["result"])
	}
}

func TestTextPlugin_Run_ReplaceAll(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":      "replace",
			"content":     "foo bar foo baz foo",
			"pattern":     "foo",
			"replacement": "qux",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if resp.Results["result"] != "qux bar qux baz qux" {
		t.Errorf("expected 'qux bar qux baz qux', got '%v'", resp.Results["result"])
	}
}

func TestTextPlugin_Run_Regex(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":  "regex",
			"content": "email: test@example.com",
			"pattern": `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if resp.Results["result"] != "test@example.com" {
		t.Errorf("expected 'test@example.com', got '%v'", resp.Results["result"])
	}
}

func TestTextPlugin_Run_Split(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":    "split",
			"content":   "apple,banana,orange",
			"delimiter": ",",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	result, ok := resp.Results["result"].([]string)
	if !ok {
		t.Fatal("result should be []string")
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}
	if result[0] != "apple" || result[1] != "banana" || result[2] != "orange" {
		t.Errorf("expected [apple, banana, orange], got %v", result)
	}
}

func TestTextPlugin_Run_Split_TrimSpaces(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":    "split",
			"content":   "  apple  ,  banana  ,  orange  ",
			"delimiter": ",",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}

	result, ok := resp.Results["result"].([]string)
	if !ok {
		t.Fatal("result should be []string")
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}
	if result[0] != "apple" || result[1] != "banana" || result[2] != "orange" {
		t.Errorf("expected [apple, banana, orange], got %v", result)
	}
}

func TestTextPlugin_Run_Join(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":    "join",
			"items":     "apple,banana,orange",
			"delimiter": "|",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if resp.Results["result"] != "apple|banana|orange" {
		t.Errorf("expected 'apple|banana|orange', got '%v'", resp.Results["result"])
	}
}

func TestTextPlugin_Run_MissingAction(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
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
	if resp.Message == "" || resp.Message != "action is required" {
		t.Errorf("expected 'action is required', got '%s'", resp.Message)
	}
}

func TestTextPlugin_Run_UnknownAction(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":  "unknown",
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
	if resp.Message == "" || resp.Message != "unknown action: unknown" {
		t.Errorf("expected 'unknown action: unknown', got '%s'", resp.Message)
	}
}

func TestTextPlugin_Run_Replace_MissingPattern(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":      "replace",
			"content":     "hello world",
			"replacement": "go",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message == "" || resp.Message != "pattern is required for replace action" {
		t.Errorf("expected 'pattern is required for replace action', got '%s'", resp.Message)
	}
}

func TestTextPlugin_Run_Split_MissingDelimiter(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":  "split",
			"content": "apple,banana,orange",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message == "" || resp.Message != "delimiter or pattern is required for split action" {
		t.Errorf("expected 'delimiter or pattern is required for split action', got '%s'", resp.Message)
	}
}

func TestTextPlugin_Run_Join_MissingDelimiter(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action": "join",
			"items":  "apple,banana,orange",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure, got success")
	}
	if resp.Message == "" || resp.Message != "delimiter is required for join action" {
		t.Errorf("expected 'delimiter is required for join action', got '%s'", resp.Message)
	}
}

func TestTextPlugin_Run_CustomResultKey(t *testing.T) {
	tc := newTestContext(t)
	p := tc.newPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"action":      "replace",
			"content":     "hello world",
			"pattern":     "world",
			"replacement": "go",
			"result_key":  "modified_text",
		},
	}

	resp, err := p.Run(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if resp.Results["modified_text"] != "hello go" {
		t.Errorf("expected 'hello go' in modified_text, got '%v'", resp.Results["modified_text"])
	}
}
