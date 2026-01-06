# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Basenana is a Go plugin system for workflow/file system operations. It provides a plugin architecture with two plugin types:
- **Source plugins**: Generate content/files (Type: "source")
- **Process plugins**: Perform operations like delays, file handling (Type: "process")

## Commands

```bash
# Build the project
go build ./...

# Run tests
go test ./...
```

## Architecture

### Plugin Interface Hierarchy

```
Plugin (base interface)
├── Name() string
├── Type() types.PluginType
├── Version() string
│
├── ProcessPlugin: Run(ctx, *Request) (*Response, error)
│   └── SourcePlugin: SourceInfo() (string, error)
```

### Key Components

| File | Purpose |
|------|---------|
| `registry.go` | Thread-safe plugin manager with `ListPlugins()` and `Call()` methods |
| `api/request.go` | Request/Response types with JobID, Namespace, WorkingPath, Parameters |
| `api/fs.go` | NanaFS interface for file system operations |
| `types/spec.go` | PluginSpec and PluginCall types |

### Request/Response API

```go
// Request fields
type Request struct {
    JobID       string              // Job identifier
    Namespace   string              // Plugin namespace
    WorkingPath string              // Working directory for file operations
    PluginName  string              // Name of the plugin being called
    Parameter   map[string]string   // Plugin parameters
    Store       PersistentStore     // Persistent storage interface
    FS          NanaFS              // File system interface
}

// Response types
api.NewResponse()                          // Empty success response
api.NewResponseWithResult(map[string]any)  // Success with result data
api.NewFailedResponse("error message")     // Failure response

// Parameter access
api.GetParameter("key", request, "default")  // Get string parameter
```

## Built-in Plugins

### delay (Process)
Pauses execution for a specified duration.

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `delay` | Yes* | - | Duration (e.g., "5s", "1m30s") |
| `until` | Yes* | - | RFC3339 timestamp |

*Either `delay` or `until` must be provided.

### three_body (Source)
Generates a timestamped file in the working directory.

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| None |

**Result**: Returns `file_path` and `size`.

### archive (Process)
Extracts archive files (zip, tar, gzip).

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `file_path` | Yes | - | Path to archive file |
| `format` | Yes | - | Archive format: `zip`, `tar`, `gzip` |
| `dest_path` | No | `.` | Destination directory |

### save (Process)
Saves content to a file.

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `content` | Yes | - | File content |
| `dest_path` | Yes | - | Destination file path |
| `mode` | No | `0644` | File permission (octal) |

### fssaver (Process)
Saves files to NanaFS with metadata.

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `file_path` | Yes | - | Source file path |
| `name` | No | filename | Entry name |
| `parent_uri` | No | - | Parent entry URI |
| `title` | No | - | Entry title |
| `author` | No | - | Author name |
| `year` | No | - | Publication year |
| `source` | No | - | Source URL |
| `abstract` | No | - | Abstract content |
| `keywords` | No | - | Comma-separated keywords |
| `url` | No | - | Source URL |
| `header_image` | No | - | Header image URL |
| `unread` | No | `false` | Mark as unread |
| `marked` | No | `false` | Mark as starred |

**Result**: Returns `saved`, `name`, `parentUri`.

### checksum (Process)
Computes file checksums.

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `file_path` | Yes | - | Path to file |
| `algorithm` | No | `md5` | Hash algorithm: `md5`, `sha256` |

**Result**: Returns `hash`.

### fileop (Process)
File operations: copy, move, rename, delete.

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `action` | Yes | - | Action: `cp`, `mv`, `rm`, `rename` |
| `src` | Yes | - | Source path |
| `dest` | Yes* | - | Destination path (*required for `cp`, `mv`, `rename`) |

### text (Process)
Text manipulation operations.

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `action` | Yes | - | Action: `search`, `replace`, `regex`, `split`, `join` |
| `content` | Yes* | - | Input text (*not required for `join`) |
| `result_key` | No | `result` | Result key name |

**Actions**:
- `search`: Check if content contains pattern
  - `pattern`: Search pattern
- `replace`: Replace text
  - `pattern`: Search pattern
  - `replacement`: Replacement text
  - `count`: Max replacements (-1 for all)
- `regex`: Extract first regex match
  - `pattern`: Regular expression
- `split`: Split text by delimiter
  - `delimiter` or `pattern`: Split separator
- `join`: Join items with delimiter
  - `delimiter`: Join separator
  - `items`: Comma-separated items

### metadata (Process)
Get file metadata.

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `file_path` | Yes | - | Path to file |

**Result**: Returns `size`, `modified`, `mode`, `is_dir`.

### rss (Source)
Sync RSS/Atom feeds and archive articles.

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `feed` | Yes | - | RSS/Atom feed URL |
| `file_type` | No | `webarchive` | Archive format: `url`, `html`, `rawhtml`, `webarchive` |
| `timeout` | No | `120` | Download timeout (seconds) |
| `clutter_free` | No | `true` | Enable clutter-free mode |
| `header_*` | No | - | Custom HTTP headers |

**Result**: Returns `articles` array with `file_path`, `size`, `title`, `url`, `updated_at`.

### docloader (Process)
Loads and parses documents, extracting metadata and content.

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `file_path` | Yes | - | Path to document file |

**Supported formats**:
- PDF (`.pdf`)
- Text (`.txt`, `.md`, `.markdown`)
- HTML (`.html`, `.htm`, `.webarchive`)
- EPUB (`.epub`)
- CSV (`.csv`)

**Result**: Returns `document` map with fields:
| Field | Type | Description |
|-------|------|-------------|
| `content` | string | Document text content |
| `title` | string | Document title |
| `author` | string | Author name |
| `abstract` | string | Document abstract/summary |
| `keywords` | string | Keywords (comma-separated) |
| `source` | string | Source/publisher |
| `publish_at` | string | Publish timestamp (Unix) |
| `header_image` | string | Header image URL (HTML only) |
| `year` | string | Publication year |

## How to Add a New Plugin

### 1. Create Plugin File

Create a new file in a subdirectory (e.g., `myplugin/myplugin.go`):

```go
package myplugin

import (
    "context"

    "github.com/basenana/plugin/api"
    "github.com/basenana/plugin/types"
)

const (
    pluginName    = "myplugin"
    pluginVersion = "1.0"
)

var PluginSpec = types.PluginSpec{
    Name:    pluginName,
    Version: pluginVersion,
    Type:    types.TypeProcess,  // or types.TypeSource
}

type MyPlugin struct{}

func (p *MyPlugin) Name() string           { return pluginName }
func (p *MyPlugin) Type() types.PluginType { return types.TypeProcess }
func (p *MyPlugin) Version() string        { return pluginVersion }

func (p *MyPlugin) Run(ctx context.Context, request *api.Request) (*api.Response, error) {
    // Get parameters
    param := api.GetParameter("param_key", request, "default")

    // Perform action
    // ...

    // Return response
    return api.NewResponse(), nil
}

func NewMyPlugin() *MyPlugin {
    return &MyPlugin{}
}
```

### 2. Register Plugin in Init()

Edit `registry.go` and add registration in `Init()` function:

```go
func Init() (Manager, error) {
    r := &registry{
        plugins: map[string]*pluginInfo{},
        logger:  logger.NewLogger("registry"),
    }

    r.Register(fssaver.PluginSpec.Name, fssaver.PluginSpec, &fssaver.FSSaver{})
    // Add your plugin registration
    r.Register(myplugin.PluginSpec.Name, myplugin.PluginSpec, myplugin.NewMyPlugin())

    return &manager{r: r}, nil
}
```

### 3. For SourcePlugin

If creating a SourcePlugin (generates files), implement `SourceInfo()` method:

```go
type SourcePlugin interface {
    ProcessPlugin
    SourceInfo() (string, error)  // Returns category identifier like "category.PluginName"
}

func (p *MyPlugin) SourceInfo() (string, error) {
    return "category.MyGenerator", nil
}
```

### 4. Key Conventions

- Plugins are singletons stored in the registry
- All plugin execution uses context for cancellation
- Use `api.NewFailedResponse()` for user-facing errors (returns Response, not error)
- Return actual errors only for exceptional conditions
- Access working directory via `request.WorkingPath`
- Use `api.GetParameter()` to access plugin parameters

## Logging

Uses uber-go/zap. Initialize via `logger.SetLogger()` in `logger/logger.go`.