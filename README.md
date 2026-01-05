# Plugin Examples

This directory contains example plugins demonstrating how to implement the two plugin types in the Basenana system.

## Plugin Types

- [ProcessPlugin Example](#processplugin-example) - `process.go`
- [SourcePlugin Example](#sourceplugin-example) - `source.go`

---

## ProcessPlugin Example

**File:** `process.go`

This example demonstrates how to implement a `ProcessPlugin` that performs an action and returns a response.

### Interface Definition

```go
type Plugin interface {
    Name() string
    Type() types.PluginType
    Version() string
}

type ProcessPlugin interface {
    Plugin
    Run(ctx context.Context, request *api.Request) (*api.Response, error)
}
```

### Implementation Structure

```go
type DelayProcessPlugin struct{}

func (d *DelayProcessPlugin) Name() string    { return "delay" }
func (d *DelayProcessPlugin) Type() types.PluginType { return types.TypeProcess }
func (d *DelayProcessPlugin) Version() string { return "1.0" }

func (d *DelayProcessPlugin) Run(ctx context.Context, request *api.Request) (*api.Response, error) {
    // Get parameters
    delayStr := api.GetParameter("delay", request, "")

    // Perform action
    // ...

    // Return response
    return api.NewResponse(), nil
}
```

### Key Points

1. **Name()**: Returns the unique plugin identifier
2. **Type()**: Returns `types.TypeProcess` for ProcessPlugin
3. **Version()**: Returns semantic version string
4. **Run()**: Main execution method that:
   - Receives a context and request
   - Extracts parameters using `api.GetParameter()`
   - Performs the plugin's action
   - Returns a response or error

### Parameter Access

```go
// Get string parameter with default value
value := api.GetParameter("key", request, "default")

// Access request fields
workingPath := request.WorkingPath
parameters := request.Parameter
```

---

## SourcePlugin Example

**File:** `source.go`

This example demonstrates how to implement a `SourcePlugin` that generates content/files and extends ProcessPlugin.

### Interface Definition

```go
type SourcePlugin interface {
    ProcessPlugin
    SourceInfo() (string, error)
}
```

### Implementation Structure

```go
type ThreeBodyPlugin struct{}

func (d *ThreeBodyPlugin) Name() string    { return "three_body" }
func (d *ThreeBodyPlugin) Type() types.PluginType { return types.TypeSource }
func (d *ThreeBodyPlugin) Version() string { return "1.0" }

// SourcePlugin specific method
func (d *ThreeBodyPlugin) SourceInfo() (string, error) {
    return "internal.FileGenerator", nil
}

// Inherits Run() from ProcessPlugin
func (d *ThreeBodyPlugin) Run(ctx context.Context, request *api.Request) (*api.Response, error) {
    // Generate content/files
    // ...

    // Return result
    return api.NewResponseWithResult(map[string]any{
        "file_path": "output.txt",
        "size":      1024,
    }), nil
}
```

### Key Points

1. **SourcePlugin extends ProcessPlugin**: Inherits all ProcessPlugin methods
2. **SourceInfo()**: Returns a category identifier for the source type (format: `category.Name`)
3. **Run()**: Typically generates files or content in the `WorkingPath`

---

## Response Types

### Successful Response

```go
// Empty response
api.NewResponse()

// Response with result data
api.NewResponseWithResult(map[string]any{
    "key": "value",
})
```

### Failed Response

```go
api.NewFailedResponse("error message")
```

---

## Example: DelayProcessPlugin (Full Implementation)

```go
package plugin

import (
    "context"
    "fmt"
    "time"
    "github.com/basenana/plugin/api"
    "github.com/basenana/plugin/types"
)

type DelayProcessPlugin struct{}

func (d *DelayProcessPlugin) Name() string    { return "delay" }
func (d *DelayProcessPlugin) Type() types.PluginType { return types.TypeProcess }
func (d *DelayProcessPlugin) Version() string { return "1.0" }

func (d *DelayProcessPlugin) Run(ctx context.Context, request *api.Request) (*api.Response, error) {
    delayStr := api.GetParameter("delay", request, "")
    untilStr := api.GetParameter("until", request, "")

    var until time.Time
    switch {
    case delayStr != "":
        duration, _ := time.ParseDuration(delayStr)
        until = time.Now().Add(duration)
    case untilStr != "":
        until, _ = time.Parse(untilStr, time.RFC3339)
    default:
        return api.NewFailedResponse("unknown action"), nil
    }

    if time.Now().Before(until) {
        timer := time.NewTimer(until.Sub(time.Now()))
        defer timer.Stop()
        select {
        case <-timer.C:
            return api.NewResponse(), nil
        case <-ctx.Done():
            return api.NewFailedResponse(ctx.Err().Error()), nil
        }
    }

    return api.NewResponse(), nil
}
```

---

## Example: ThreeBodyPlugin (Full Implementation)

```go
package plugin

import (
    "context"
    "fmt"
    "os"
    "path"
    "time"
    "github.com/basenana/plugin/api"
    "github.com/basenana/plugin/types"
)

type ThreeBodyPlugin struct{}

func (d *ThreeBodyPlugin) Name() string    { return "three_body" }
func (d *ThreeBodyPlugin) Type() types.PluginType { return types.TypeSource }
func (d *ThreeBodyPlugin) Version() string { return "1.0" }
func (d *ThreeBodyPlugin) SourceInfo() (string, error) {
    return "internal.FileGenerator", nil
}

func (d *ThreeBodyPlugin) Run(ctx context.Context, request *api.Request) (*api.Response, error) {
    if request.WorkingPath == "" {
        return nil, fmt.Errorf("workdir is empty")
    }

    timestamp := time.Now().Unix()
    filePath := path.Join(request.WorkingPath, fmt.Sprintf("3_body_%d.txt", timestamp))
    fileData := []byte(fmt.Sprintf("%d - Do not answer!\n", timestamp))

    err := os.WriteFile(filePath, fileData, 0655)
    if err != nil {
        return api.NewFailedResponse(fmt.Sprintf("file generate failed: %s", err)), nil
    }

    return api.NewResponseWithResult(map[string]any{
        "file_path": path.Base(filePath),
        "size":      int64(len(fileData)),
    }), nil
}
```
