/*
 Copyright 2023 NanaFS Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package docloader

import (
	"strings"
	"testing"
)

func TestHTML_ExtractMetadata(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Test Document Title</title>
    <meta name="author" content="Test Author">
    <meta name="description" content="This is a test description">
    <meta name="keywords" content="go,testing,unit-test">
</head>
<body>
    <p>Content here</p>
</body>
</html>`

	if err := testFileAccess.Write("test.html", []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to create test HTML file: %v", err)
	}

	absPath, _ := testFileAccess.GetAbsPath("test.html")
	got := extractHTMLMetadata(absPath)

	if got.Title != "Test Document Title" {
		t.Errorf("title = %q, want %q", got.Title, "Test Document Title")
	}
	if got.Author != "Test Author" {
		t.Errorf("author = %q, want %q", got.Author, "Test Author")
	}
	if got.Abstract != "This is a test description" {
		t.Errorf("abstract = %q, want %q", got.Abstract, "This is a test description")
	}
	if len(got.Keywords) != 3 || got.Keywords[0] != "go" || got.Keywords[1] != "testing" || got.Keywords[2] != "unit-test" {
		t.Errorf("keywords = %v, want [go, testing, unit-test]", got.Keywords)
	}
}

func TestHTML_ExtractMetadata_OGTags(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Page Title</title>
    <meta property="og:title" content="OG Title">
    <meta property="og:description" content="OG Description">
    <meta property="og:image" content="https://example.com/image.jpg">
    <meta property="og:site_name" content="Example Site">
</head>
<body>Test</body>
</html>`

	if err := testFileAccess.Write("og_test.html", []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to create test HTML file: %v", err)
	}

	absPath, _ := testFileAccess.GetAbsPath("og_test.html")
	got := extractHTMLMetadata(absPath)

	// OG tags should override empty title tag
	if got.Title != "OG Title" {
		t.Errorf("title = %q, want %q", got.Title, "OG Title")
	}
	if got.Abstract != "OG Description" {
		t.Errorf("abstract = %q, want %q", got.Abstract, "OG Description")
	}
	if got.HeaderImage != "https://example.com/image.jpg" {
		t.Errorf("headerImage = %q, want %q", got.HeaderImage, "https://example.com/image.jpg")
	}
	if got.Source != "Example Site" {
		t.Errorf("source = %q, want %q", got.Source, "Example Site")
	}
}

func TestHTML_ExtractMetadata_DCTags(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <meta name="dc.creator" content="DC Author">
    <meta name="dc.description" content="DC Description">
    <meta name="dc.subject" content="tag1,tag2,tag3">
    <meta name="dc.publisher" content="DC Publisher">
</head>
<body>Test</body>
</html>`

	if err := testFileAccess.Write("dc_test.html", []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to create test HTML file: %v", err)
	}

	absPath, _ := testFileAccess.GetAbsPath("dc_test.html")
	got := extractHTMLMetadata(absPath)

	if got.Author != "DC Author" {
		t.Errorf("author = %q, want %q", got.Author, "DC Author")
	}
	if got.Abstract != "DC Description" {
		t.Errorf("abstract = %q, want %q", got.Abstract, "DC Description")
	}
	if len(got.Keywords) != 3 || got.Keywords[0] != "tag1" || got.Keywords[1] != "tag2" || got.Keywords[2] != "tag3" {
		t.Errorf("keywords = %v, want [tag1, tag2, tag3]", got.Keywords)
	}
	if got.Source != "DC Publisher" {
		t.Errorf("source = %q, want %q", got.Source, "DC Publisher")
	}
}

func TestHTML_TitleFromTag(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>HTML Tag Title</title>
</head>
<body>Test content</body>
</html>`

	if err := testFileAccess.Write("title_test.html", []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to create test HTML file: %v", err)
	}

	absPath, _ := testFileAccess.GetAbsPath("title_test.html")
	got := extractHTMLMetadata(absPath)

	if got.Title != "HTML Tag Title" {
		t.Errorf("title = %q, want %q", got.Title, "HTML Tag Title")
	}
}

func TestHTML_Load(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Load Test</title>
    <meta name="author" content="Test Author">
</head>
<body>
    <p>This is the content body.</p>
</body>
</html>`

	if err := testFileAccess.Write("load_test.html", []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to create test HTML file: %v", err)
	}

	absPath, _ := testFileAccess.GetAbsPath("load_test.html")
	parser := NewHTML(absPath, nil)
	doc, err := parser.Load(nil)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if doc.Properties.Title != "Load Test" {
		t.Errorf("title = %q, want %q", doc.Properties.Title, "Load Test")
	}
	if doc.Properties.Author != "Test Author" {
		t.Errorf("author = %q, want %q", doc.Properties.Author, "Test Author")
	}
	if doc.Content == "" {
		t.Error("content should not be empty")
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"removes simple tags", "<p>Hello World</p>", "Hello World"},
		{"removes nested tags", "<div><p>Content</p></div>", "Content"},
		{"removes script tags", "<script>alert('xss')</script><p>Safe</p>", "Safe"},
		{"removes style tags", "<style>.hidden{display:none}</style><p>Visible</p>", "Visible"},
		{"converts br tags to newlines", "Line1<br/>Line2", "Line1\nLine2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHTMLTags(tt.input)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("stripHTMLTags(%q) should contain %q, got %q", tt.input, tt.contains, got)
			}
		})
	}
}
