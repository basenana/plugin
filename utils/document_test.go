package utils

import (
	"strings"
	"testing"
)

func TestGenerateContentAbstract(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		contains    []string // result should contain these substrings
		notContains []string // result should NOT contain these substrings
		maxLen      int      // max allowed length for result
	}{
		{
			name: "removes script content",
			input: `<html><body>
<script>var x = "should not appear in abstract";</script>
<p>This is a paragraph that should be extracted.</p>
</body></html>`,
			contains:    []string{"This is a paragraph"},
			notContains: []string{"should not appear", "should not appear in abstract"},
		},
		{
			name: "removes style content",
			input: `<html><body>
<style>.foo { color: red; }</style>
<p>Style should be removed.</p>
</body></html>`,
			contains:    []string{"Style should be removed"},
			notContains: []string{"color: red", ".foo"},
		},
		{
			name: "removes nav content",
			input: `<html><body>
<nav>Navigation links home about contact</nav>
<p>Main content paragraph.</p>
</body></html>`,
			contains:    []string{"Main content paragraph"},
			notContains: []string{"Navigation", "home", "about", "contact"},
		},
		{
			name: "removes header footer",
			input: `<html><body>
<header>Site header</header>
<article><p>Article content here.</p></article>
<footer>Site footer</footer>
</body></html>`,
			contains:    []string{"Article content here"},
			notContains: []string{"Site header", "Site footer"},
		},
		{
			name: "removes aside",
			input: `<html><body>
<aside>Sidebar advertisement</aside>
<p>Main content.</p>
</body></html>`,
			contains:    []string{"Main content"},
			notContains: []string{"Sidebar", "advertisement"},
		},
		{
			name: "extracts from article tag",
			input: `<html><body>
<article><p>First paragraph in article.</p><p>Second paragraph in article.</p></article>
</body></html>`,
			contains: []string{"First paragraph", "Second paragraph"},
		},
		{
			name: "extracts from section tag",
			input: `<html><body>
<section><p>Content inside section.</p></section>
</body></html>`,
			contains: []string{"Content inside section"},
		},
		{
			name: "extracts from li tag",
			input: `<html><body>
<ul><li>List item one</li><li>List item two</li></ul>
</body></html>`,
			contains: []string{"List item one", "List item two"},
		},
		{
			name: "extracts from td th tag",
			input: `<html><body>
<table><tr><th>Header</th><td>Data</td></tr></table>
</body></html>`,
			contains: []string{"Header", "Data"},
		},
		{
			name: "handles noscript",
			input: `<html><body>
<noscript>This page requires JavaScript</noscript>
<p>Fallback content.</p>
</body></html>`,
			contains:    []string{"Fallback content"},
			notContains: []string{"requires JavaScript"},
		},
		{
			name: "handles iframe",
			input: `<html><body>
<iframe src="https://example.com"></iframe>
<p>Iframe content below.</p>
</body></html>`,
			contains: []string{"Iframe content below"},
		},
		{
			name: "removes inline CSS classes",
			input: `<html><body>
<p class="some-class" style="color: red;">Paragraph with inline styles.</p>
</body></html>`,
			contains:    []string{"Paragraph with inline styles"},
			notContains: []string{"some-class", "color: red"},
		},
		{
			name:     "handles empty input",
			input:    "",
			contains: nil,
		},
		{
			name:     "handles plain text",
			input:    "Just plain text without HTML tags.",
			contains: []string{"Just plain text"},
		},
		{
			name:    "truncates at 400 characters",
			input:   `<p>` + strings.Repeat("word ", 100) + `</p>`,
			maxLen:  400,
		},
		{
			name:    "handles unicode correctly",
			input:   `<p>中文测试内容中文测试内容中文测试内容中文测试内容中文测试内容中文测试内容中文测试内容中文测试内容</p>`,
			maxLen:  400,
		},
		{
			name:    "falls back to body text",
			input:   `<html><body><div>Direct body content here</div></body></html>`,
			contains: []string{"Direct body content here"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateContentAbstract(tt.input)

			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("expected result to contain %q, got %q", s, result)
				}
			}

			for _, s := range tt.notContains {
				if strings.Contains(result, s) {
					t.Errorf("expected result NOT to contain %q, but got %q", s, result)
				}
			}

			if tt.maxLen > 0 && len(result) > tt.maxLen {
				t.Errorf("expected result length <= %d, got %d", tt.maxLen, len(result))
			}
		})
	}
}
