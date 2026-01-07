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
	"testing"
)

func TestCSV_Load(t *testing.T) {
	csvContent := `Name,Age,City
Alice,30,NYC
Bob,25,LA
Charlie,35,Chicago`

	if err := testFileAccess.Write("data.csv", []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	absPath, _ := testFileAccess.GetAbsPath("data.csv")
	parser := NewCSV(absPath, nil)
	doc, err := parser.Load(nil)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if doc.Content == "" {
		t.Error("content should not be empty")
	}
	if len(doc.Content) < 10 {
		t.Errorf("content too short: %q", doc.Content)
	}
}

func TestCSV_Load_WithFilenameMetadata(t *testing.T) {
	csvContent := `col1,col2,col3
val1,val2,val3`

	if err := testFileAccess.Write("Author_Title_2024.csv", []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	absPath, _ := testFileAccess.GetAbsPath("Author_Title_2024.csv")
	parser := NewCSV(absPath, nil)
	doc, err := parser.Load(nil)
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
