# ArchivePlugin

Extracts archive files (zip, tar, gzip) to a destination directory.

## Type
ProcessPlugin

## Version
1.0

## Name
`archive`

## Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `file_path` | Yes | string | Path to the archive file |
| `format` | Yes | string | Archive format: `zip`, `tar`, or `gzip` |
| `dest_path` | No | string | Destination directory (default: `.`) |

## Output

```json
{
  "success": true
}
```

On failure, returns an error message.

## Usage Example

```yaml
# Extract a zip file
- name: archive
  parameters:
    file_path: "/path/to/archive.zip"
    format: "zip"
    dest_path: "/path/to/output"

# Extract a tar.gz file
- name: archive
  parameters:
    file_path: "/path/to/archive.tar.gz"
    format: "tar"
    dest_path: "/path/to/output"

# Extract a gzip file
- name: archive
  parameters:
    file_path: "/path/to/file.gz"
    format: "gzip"
```

## Notes
- For `tar` format, the file is expected to be gzip-compressed (.tar.gz or .tgz)
- For `gzip` format, the `.gz` extension is removed from the output filename
- Creates the destination directory if it does not exist
