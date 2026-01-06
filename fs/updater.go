package fs

import (
	"context"
	"strconv"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/types"
)

const (
	updatePluginName    = "update"
	updatePluginVersion = "1.0"
)

var UpdatePluginSpec = types.PluginSpec{
	Name:    updatePluginName,
	Version: updatePluginVersion,
	Type:    types.TypeProcess,
}

type Updater struct{}

func (p *Updater) Name() string           { return updatePluginName }
func (p *Updater) Type() types.PluginType { return types.TypeProcess }
func (p *Updater) Version() string        { return updatePluginVersion }

func (p *Updater) Run(ctx context.Context, request *api.Request) (*api.Response, error) {
	entryURI := api.GetStringParameter("entry_uri", request, "")
	if entryURI == "" {
		return api.NewFailedResponse("entry_uri is required"), nil
	}

	id, err := strconv.ParseInt(entryURI, 10, 64)
	if err != nil {
		return api.NewFailedResponse("invalid entry_uri: " + entryURI), nil
	}

	props := buildProperties(request)

	if request.FS == nil {
		return api.NewFailedResponse("file system is not available"), nil
	}
	if err := request.FS.UpdateEntry(ctx, id, props); err != nil {
		return api.NewFailedResponse("failed to update entry: " + err.Error()), nil
	}

	return api.NewResponse(), nil
}
