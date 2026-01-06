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
	"testing"
	"time"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"go.uber.org/zap"
)

func init() {
	// Initialize logger for tests
	logger.SetLogger(zap.NewNop().Sugar())
}

func newDelayPlugin() *DelayProcessPlugin {
	p := &DelayProcessPlugin{}
	p.logger = logger.NewPluginLogger(delayPluginName, "test-job")
	return p
}

func TestDelayPlugin_Name(t *testing.T) {
	p := newDelayPlugin()
	if p.Name() != delayPluginName {
		t.Errorf("expected %s, got %s", delayPluginName, p.Name())
	}
}

func TestDelayPlugin_Type(t *testing.T) {
	p := newDelayPlugin()
	if string(p.Type()) != "process" {
		t.Errorf("expected process, got %s", p.Type())
	}
}

func TestDelayPlugin_Version(t *testing.T) {
	p := newDelayPlugin()
	if p.Version() != delayPluginVersion {
		t.Errorf("expected %s, got %s", delayPluginVersion, p.Version())
	}
}

func TestDelayPlugin_DelayDuration(t *testing.T) {
	p := newDelayPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"delay": "10ms",
		},
	}

	start := time.Now()
	resp, err := p.Run(ctx, req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if elapsed < 10*time.Millisecond {
		t.Errorf("expected at least 10ms delay, got %v", elapsed)
	}
}

func TestDelayPlugin_DelayDuration_Short(t *testing.T) {
	p := newDelayPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"delay": "1ms",
		},
	}

	start := time.Now()
	resp, err := p.Run(ctx, req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if elapsed < 1*time.Millisecond {
		t.Errorf("expected at least 1ms delay, got %v", elapsed)
	}
}

func TestDelayPlugin_DelayDuration_Minutes(t *testing.T) {
	p := newDelayPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"delay": "1m30s",
		},
	}

	start := time.Now()
	resp, err := p.Run(ctx, req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if elapsed < 90*time.Second {
		t.Errorf("expected at least 90s delay, got %v", elapsed)
	}
}

func TestDelayPlugin_UntilRFC3339(t *testing.T) {
	p := newDelayPlugin()
	ctx := context.Background()

	// Set until to 500ms from now to ensure enough time for execution
	until := time.Now().Add(500 * time.Millisecond).Format(time.RFC3339)

	req := &api.Request{
		Parameter: map[string]any{
			"until": until,
		},
	}

	start := time.Now()
	resp, err := p.Run(ctx, req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	if elapsed < 150*time.Millisecond {
		t.Errorf("expected at least 150ms delay, got %v", elapsed)
	}
}

func TestDelayPlugin_Until_AlreadyPassed(t *testing.T) {
	p := newDelayPlugin()
	ctx := context.Background()

	// Set until to 1 second ago
	until := time.Now().Add(-1 * time.Second).Format(time.RFC3339)

	req := &api.Request{
		Parameter: map[string]any{
			"until": until,
		},
	}

	resp, err := p.Run(ctx, req)

	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success when until is in the past, got failure: %s", resp.Message)
	}
}

func TestDelayPlugin_MissingParameters(t *testing.T) {
	p := newDelayPlugin()
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
}

func TestDelayPlugin_InvalidDuration(t *testing.T) {
	p := newDelayPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"delay": "invalid",
		},
	}

	_, err := p.Run(ctx, req)

	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestDelayPlugin_InvalidUntil(t *testing.T) {
	p := newDelayPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"until": "invalid-timestamp",
		},
	}

	_, err := p.Run(ctx, req)

	if err == nil {
		t.Error("expected error for invalid until timestamp")
	}
}

func TestDelayPlugin_ZeroDelay(t *testing.T) {
	p := newDelayPlugin()
	ctx := context.Background()

	req := &api.Request{
		Parameter: map[string]any{
			"delay": "0s",
		},
	}

	start := time.Now()
	resp, err := p.Run(ctx, req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsSucceed {
		t.Errorf("expected success, got failure: %s", resp.Message)
	}
	// Zero delay should complete immediately
	if elapsed > 10*time.Millisecond {
		t.Errorf("expected minimal delay for 0s, got %v", elapsed)
	}
}

func TestDelayPlugin_ContextCancellation(t *testing.T) {
	p := newDelayPlugin()
	ctx, cancel := context.WithCancel(context.Background())

	req := &api.Request{
		Parameter: map[string]any{
			"delay": "10s",
		},
	}

	// Cancel context immediately
	cancel()

	resp, err := p.Run(ctx, req)

	if err != nil {
		t.Fatal(err)
	}
	if resp.IsSucceed {
		t.Error("expected failure due to context cancellation")
	}
}
