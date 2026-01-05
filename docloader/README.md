# DocLoader

Loads and parses document files (PDF, TXT, MD, HTML, EPUB, webarchive).

## Type
ProcessPlugin

## Version
1.0

## Name
`docloader`

## Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `file_path` | Yes | string | Path to document file |

## Supported Formats

| Extension | Format |
|-----------|--------|
| `.pdf` | PDF Document |
| `.txt` | Plain Text |
| `.md`, `.markdown` | Markdown |
| `.html`, `.htm` | HTML |
| `.webarchive` | Web Archive |
| `.epub` | EPUB |

## Output

```json
{
  "file_path": "<original-path>",
  "document": {
    "title": "<document-title>",
    "content": "<extracted-text>",
    "properties": {
      "author": "<author>",
      "created": "<timestamp>",
      ...
    }
  }
}
```

## Usage Example

```yaml
# Load a PDF document
- name: docloader
  parameters:
    file_path: "/path/to/document.pdf"

# Load an HTML file
- name: docloader
  parameters:
    file_path: "/path/to/page.html"

# Load a markdown file
- name: docloader
  parameters:
    file_path: "/path/to/readme.md"
```

## Notes
- The `file_path` is relative to the working path
- If no title is found in the document, the filename (without extension) is used
- Properties vary by document format
- PDF properties may include: author, creator, producer, creation date, etc.
