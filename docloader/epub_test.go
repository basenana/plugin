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
	"archive/zip"
	"context"
	"testing"

	"github.com/basenana/plugin/logger"
)

func createTestEPUB(t *testing.T, path string, title, author, content string) {
	t.Helper()

	w, err := testFileAccess.Create(path, 0644)
	if err != nil {
		t.Fatalf("Failed to create EPUB: %v", err)
	}
	defer w.Close()

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	if err := addZipFile(zipWriter, "META-INF/container.xml", containerXML); err != nil {
		t.Fatalf("Failed to add container.xml: %v", err)
	}

	opfContent := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="BookId">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:title>` + title + `</dc:title>
    <dc:creator>` + author + `</dc:creator>
    <dc:description>A test EPUB description</dc:description>
    <dc:subject>test,epub</dc:subject>
    <dc:publisher>Test Publisher</dc:publisher>
    <dc:date>2024-01-15T00:00:00Z</dc:date>
  </metadata>
  <manifest>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="chapter1"/>
  </spine>
</package>`
	if err := addZipFile(zipWriter, "OEBPS/content.opf", opfContent); err != nil {
		t.Fatalf("Failed to add content.opf: %v", err)
	}

	chapterContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Chapter 1</title></head>
<body><p>` + content + `</p></body>
</html>`
	if err := addZipFile(zipWriter, "OEBPS/chapter1.xhtml", chapterContent); err != nil {
		t.Fatalf("Failed to add chapter: %v", err)
	}
}

func addZipFile(zipWriter *zip.Writer, name, content string) error {
	w, err := zipWriter.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

func TestEPUB_Load(t *testing.T) {
	loader := newDocLoader(t)

	createTestEPUB(t, "test.epub", "Test Book", "Test Author", "Chapter content here")

	doc, err := loader.loadDocument(context.Background(), "test.epub")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// EPUB content should be extracted
	if doc.Content == "" {
		t.Error("content should not be empty")
	}
	// Title should fall back to filename
	if doc.Properties.Title == "" {
		t.Error("title should fall back to filename")
	}
}

func TestEPUB_Load_InvalidFile(t *testing.T) {
	if err := testFileAccess.Write("invalid.epub", []byte("not a valid epub"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := logger.IntoContext(context.Background(), logger.NewLogger("test"))
	parser := NewEPUB("invalid.epub", nil)
	_, err := parser.Load(ctx)
	if err == nil {
		t.Error("Load should fail for invalid EPUB")
	}
}
