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
	"context"
	"testing"

	"github.com/basenana/plugin/logger"
	"github.com/basenana/plugin/types"
)

func TestText_ExtractFileNameMetadata(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     types.Properties
	}{
		{
			name:     "author_title_year pattern",
			filename: "/path/to/ResearchTeam_StudyResults_2023.txt",
			want: types.Properties{
				Author: "ResearchTeam",
				Title:  "StudyResults",
				Year:   "2023",
			},
		},
		{
			name:     "author - title (year) pattern",
			filename: "/path/to/JaneSmith - Research Paper (2024).md",
			want: types.Properties{
				Author: "JaneSmith",
				Title:  "Research Paper",
				Year:   "2024",
			},
		},
		{
			name:     "author_title (year) pattern with space",
			filename: "/path/to/Author_Title (2024).txt",
			want: types.Properties{
				Author: "Author",
				Title:  "Title",
				Year:   "2024",
			},
		},
		{
			name:     "no match returns empty",
			filename: "/path/to/data_export.txt",
			want:     types.Properties{},
		},
		{
			name:     "year only in filename",
			filename: "/path/to/some_document_2022.txt",
			want: types.Properties{
				Author: "some",
				Title:  "document",
				Year:   "2022",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFileNameMetadata(tt.filename)

			if got.Author != tt.want.Author {
				t.Errorf("author = %q, want %q", got.Author, tt.want.Author)
			}
			if got.Title != tt.want.Title {
				t.Errorf("title = %q, want %q", got.Title, tt.want.Title)
			}
			if got.Year != tt.want.Year {
				t.Errorf("year = %q, want %q", got.Year, tt.want.Year)
			}
		})
	}
}

func TestText_ExtractContentMetadata_Title(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "extracts h1 title",
			content: "# My Document Title\n\nSome content here.",
			want:    "My Document Title",
		},
		{
			name:    "first non-empty line as title",
			content: "Just a regular line\n\nAnother line",
			want:    "Just a regular line",
		},
		{
			name:    "skips short lines without spaces",
			content: "abc\n\nSome paragraph here.",
			want:    "Some paragraph here.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := types.Properties{}
			got := extractTextContentMetadata(tt.content, props)

			if got.Title != tt.want {
				t.Errorf("title = %q, want %q", got.Title, tt.want)
			}
		})
	}
}

func TestText_ExtractContentMetadata_Abstract(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "extracts first paragraph",
			content: "First paragraph here.\n\nSecond paragraph here.\n\nThird paragraph.",
			want:    "First paragraph here.",
		},
		{
			name:    "skips markdown headers in abstract",
			content: "# Title\n\nFirst paragraph.\n\n## Subtitle\nSecond paragraph.",
			want:    "First paragraph.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := types.Properties{}
			got := extractTextContentMetadata(tt.content, props)

			if got.Abstract != tt.want {
				t.Errorf("abstract = %q, want %q", got.Abstract, tt.want)
			}
		})
	}
}

func TestText_Load(t *testing.T) {
	content := `# Test Document

This is a test paragraph.
This is another test paragraph.

More content here.`

	if err := testFileAccess.Write("test.txt", []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	absPath, _ := testFileAccess.GetAbsPath("test.txt")
	parser := NewText(absPath, nil)
	ctx := logger.IntoContext(context.Background(), logger.NewLogger("test"))
	doc, err := parser.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if doc.Properties.Title != "Test Document" {
		t.Errorf("title = %q, want %q", doc.Properties.Title, "Test Document")
	}
	if doc.Content == "" {
		t.Error("content should not be empty")
	}
}

func TestText_Load_WithFileNameMetadata(t *testing.T) {
	content := `Just some content without title.`

	if err := testFileAccess.Write("Author_Title_2024.txt", []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	absPath, _ := testFileAccess.GetAbsPath("Author_Title_2024.txt")
	parser := NewText(absPath, nil)
	ctx := logger.IntoContext(context.Background(), logger.NewLogger("test"))
	doc, err := parser.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if doc.Properties.Author != "Author" {
		t.Errorf("author = %q, want %q", doc.Properties.Author, "Author")
	}
	if doc.Properties.Title != "Title" {
		t.Errorf("title = %q, want %q", doc.Properties.Title, "Title")
	}
	if doc.Properties.Year != "2024" {
		t.Errorf("year = %q, want %q", doc.Properties.Year, "2024")
	}
}

func TestText_Load_Markdown(t *testing.T) {
	content := `# Markdown Title

Some **formatted** content.`

	if err := testFileAccess.Write("document.md", []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	absPath, _ := testFileAccess.GetAbsPath("document.md")
	parser := NewText(absPath, nil)
	ctx := logger.IntoContext(context.Background(), logger.NewLogger("test"))
	doc, err := parser.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if doc.Properties.Title != "Markdown Title" {
		t.Errorf("title = %q, want %q", doc.Properties.Title, "Markdown Title")
	}
}
