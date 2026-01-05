# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Basenana is a Go plugin system for workflow/file system operations. It provides a plugin architecture with two plugin types:
- **Source plugins**: Generate content/files
- **Process plugins**: Perform operations (e.g., delays)

## Commands

```bash
# Build the project
go build ./...

# Run tests
go test ./...
```

## Architecture

**Plugin Interface Hierarchy**:
```
Plugin (base)
├── ProcessPlugin: Run(ctx, *Request) (*Response, error)
    └── SourcePlugin: SourceInfo() (string, error)
```

**Key Components**:

1. **Registry** (`registry.go`): Thread-safe plugin manager with `ListPlugins()` and `Call()` methods
2. **API** (`api/request.go`): Request/Response types with JobID, Namespace, WorkingPath, Parameters
3. **Results** (`types/results.go`): Two implementations - `memoryBasedResults` and `fileBasedResults` (Gob-encoded)
4. **Logging**: Uses uber-go/zap, set via `SetLogger()`

**Built-in Plugins**:
- `DelayProcessPlugin`: Time-based delays (duration or RFC3339 timestamp)
- `ThreeBodyPlugin`: Source plugin generating timestamped files

## Conventions

- Plugins are singletons stored in the registry
- All plugin execution uses context for cancellation
- Results use Gob encoding for persistence