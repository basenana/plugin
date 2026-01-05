# SavePlugin

Saves content to a file at the specified path.

## Type
ProcessPlugin

## Version
1.0

## Name
`save`

## Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `content` | Yes | string | The text content to write |
| `dest_path` | Yes | string | Destination file path |
| `mode` | No | string | File permissions in octal (default: `0644`) |

## Output

```json
{
  "success": true
}
```

On failure, returns an error message.

## Usage Example

```yaml
# Save content to a file with default permissions
- name: save
  parameters:
    content: "Hello, World!"
    dest_path: "/path/to/output.txt"

# Save with custom permissions
- name: save
  parameters:
    content: "#!/bin/bash\necho 'Hello'"
    dest_path: "/path/to/script.sh"
    mode: "0755"
```

## Notes
- Creates parent directories if they do not exist
- Overwrites the file if it already exists
- Permissions are interpreted as octal (e.g., "0644" = rw-r--r--)
