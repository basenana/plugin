# 🔌 Plugin

This is the plugin repository for [NanaFS](https://github.com/basenana/nanafs), providing extensible plugins for
workflow and file system operations.

The `process.go` and `source.go` files in the root directory are example implementations demonstrating how to develop
plugins. All built-in plugins are organized in subdirectories.

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
