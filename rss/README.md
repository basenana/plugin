# RssSourcePlugin

Fetches RSS/Atom feeds and archives articles in specified format (url, html, rawhtml, webarchive).

## Type
SourcePlugin

## Version
1.0

## Name
`rss`

## Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `feed` | Yes | string | RSS feed URL |
| `file_type` | No | string | Output format: `url`, `html`, `rawhtml`, `webarchive` (default: `webarchive`) |
| `timeout` | No | integer | Download timeout in seconds (default: 120) |
| `clutter_free` | No | boolean | Remove clutter from HTML (default: `true`) |
| `header_*` | No | string | Custom HTTP headers (prefix with `header_`) |

## Output

```json
{
  "articles": [
    {
      "file_path": "<filename>",
      "size": <file-size-in-bytes>,
      "title": "<article-title>",
      "url": "<article-url>",
      "site_url": "<site-url>",
      "site_name": "<site-name>",
      "updated_at": "<RFC3339-timestamp>"
    },
    ...
  ]
}
```

### Article Structure

| Field | Type | Description |
|-------|------|-------------|
| `file_path` | string | Filename of the archived article |
| `size` | int64 | File size in bytes |
| `title` | string | Article title |
| `url` | string | Original article URL |
| `site_url` | string | Site URL of the feed |
| `site_name` | string | Site name of the feed |
| `updated_at` | string | Publication/update time in RFC3339 format |

## File Type Formats

| Format | Description |
|--------|-------------|
| `url` | Internet Shortcut file (.url) |
| `html` | Readable HTML file |
| `rawhtml` | Full HTML with clutter removal |
| `webarchive` | Web Archive format (.webarchive) |

## Usage Example

```yaml
# Fetch RSS feed with default settings
- name: rss
  parameters:
    feed: "https://example.com/feed.xml"
  working_path: "/path/to/output"

# Fetch with custom timeout
- name: rss
  parameters:
    feed: "https://example.com/feed.xml"
    timeout: 60
    file_type: "html"

# Fetch with custom headers
- name: rss
  parameters:
    feed: "https://example.com/feed.xml"
    header_User-Agent: "MyBot/1.0"
```

## Notes
- Uses persistent store to track already-processed articles to avoid duplicates
- Maximum 50 articles processed per feed
- For RSSHub feeds, automatically uses `html` format
- Custom headers are passed to the web packer
