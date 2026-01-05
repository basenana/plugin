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
	"errors"
	"fmt"
	"sync"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/types"
	"go.uber.org/zap"
)

var (
	ErrNotFound = errors.New("PluginNotFound")
)

type Manager interface {
	ListPlugins() []types.PluginSpec
	Call(ctx context.Context, ps types.PluginCall, req *api.Request) (resp *api.Response, err error)
}

type manager struct {
	r *registry
}

func (m *manager) ListPlugins() []types.PluginSpec {
	infos := m.r.List()
	var result = make([]types.PluginSpec, 0, len(infos))
	for _, i := range infos {
		result = append(result, i.spec)
	}
	return result
}

func (m *manager) Call(ctx context.Context, ps types.PluginCall, req *api.Request) (resp *api.Response, err error) {
	var plugin Plugin
	plugin, err = m.r.BuildPlugin(ps)
	if err != nil {
		return nil, err
	}

	runnablePlugin, ok := plugin.(ProcessPlugin)
	if !ok {
		return nil, fmt.Errorf("not process plugin")
	}
	return runnablePlugin.Run(ctx, req)
}

type Plugin interface {
	Name() string
	Type() types.PluginType
	Version() string
}

func Init() (Manager, error) {
	r := &registry{
		plugins: map[string]*pluginInfo{},
		logger:  logger.NewLogger("registry"),
	}

	return &manager{r: r}, nil
}

type registry struct {
	plugins map[string]*pluginInfo
	mux     sync.RWMutex
	logger  *zap.SugaredLogger
}

func (r *registry) BuildPlugin(ps types.PluginCall) (Plugin, error) {
	r.mux.RLock()
	p, ok := r.plugins[ps.PluginName]
	if !ok {
		r.mux.RUnlock()
		r.logger.Warnw("build plugin failed", "plugin", ps.PluginName)
		return nil, ErrNotFound
	}
	r.mux.RUnlock()
	return p.singleton, nil
}

func (r *registry) Register(pluginName string, spec types.PluginSpec, singleton Plugin) {
	r.mux.Lock()
	r.plugins[pluginName] = &pluginInfo{
		singleton: singleton,
		spec:      spec,
		buildIn:   true,
	}
	r.mux.Unlock()
}

func (r *registry) List() []*pluginInfo {
	var result = make([]*pluginInfo, 0, len(r.plugins))
	r.mux.Lock()
	for _, p := range r.plugins {
		result = append(result, p)
	}
	r.mux.Unlock()
	return result
}

type pluginInfo struct {
	singleton Plugin
	spec      types.PluginSpec
	disable   bool
	buildIn   bool
}
