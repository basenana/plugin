package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func setupTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "fileaccess_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return tmpDir
}

func cleanupTestDir(t *testing.T, dir string) {
	os.RemoveAll(dir)
}

func TestNewFileAccess(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	if fa.workdir != filepath.Clean(dir) {
		t.Errorf("expected workdir %s, got %s", filepath.Clean(dir), fa.workdir)
	}
}

func TestNewFileAccess_EmptyWorkdir(t *testing.T) {
	fa := NewFileAccess("")
	if fa.workdir != "." {
		t.Errorf("expected workdir '.', got %s", fa.workdir)
	}
}

func TestRead(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	testFile := filepath.Join(dir, "test.txt")
	testData := []byte("hello world")

	err := os.WriteFile(testFile, testData, 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	data, err := fa.Read("test.txt")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("expected %s, got %s", testData, data)
	}
}

func TestRead_NestedPath(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	nestedDir := filepath.Join(dir, "nested")
	err := os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	testFile := filepath.Join(nestedDir, "file.txt")
	testData := []byte("nested content")
	err = os.WriteFile(testFile, testData, 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	data, err := fa.Read("nested/file.txt")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("expected %s, got %s", testData, data)
	}
}

func TestWrite(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	testData := []byte("write test")

	err := fa.Write("newfile.txt", testData, 0644)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "newfile.txt"))
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("expected %s, got %s", testData, data)
	}
}

func TestWrite_NestedPath(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	testData := []byte("nested write")

	err := fa.MkdirAll("subdir/nested", 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	err = fa.Write("subdir/nested/file.txt", testData, 0644)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "subdir/nested/file.txt"))
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("expected %s, got %s", testData, data)
	}
}

func TestStat(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	testFile := filepath.Join(dir, "stat_test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	info, err := fa.Stat("stat_test.txt")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Name() != "stat_test.txt" {
		t.Errorf("expected name 'stat_test.txt', got %s", info.Name())
	}
	if info.IsDir() {
		t.Error("expected file, not directory")
	}
}

func TestStat_NotExists(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	_, err := fa.Stat("nonexistent.txt")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestMkdirAll(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	err := fa.MkdirAll("newdir/subdir", 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "newdir/subdir"))
	if err != nil {
		t.Fatalf("failed to stat created dir: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestRename(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	srcFile := filepath.Join(dir, "src.txt")
	err := os.WriteFile(srcFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	err = fa.Rename("src.txt", "dst.txt")
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "src.txt")); err == nil {
		t.Error("source file should not exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "dst.txt")); err != nil {
		t.Errorf("destination file should exist: %v", err)
	}
}

func TestRemove(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	testFile := filepath.Join(dir, "remove_test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err = fa.Remove("remove_test.txt")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if _, err := os.Stat(testFile); err == nil {
		t.Error("file should be removed")
	}
}

func TestExists(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	testFile := filepath.Join(dir, "exists_test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if !fa.Exists("exists_test.txt") {
		t.Error("expected file to exist")
	}
	if fa.Exists("nonexistent.txt") {
		t.Error("expected file to not exist")
	}
}

func TestGetAbsPath(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	absPath, err := fa.GetAbsPath("subdir/file.txt")
	if err != nil {
		t.Fatalf("GetAbsPath failed: %v", err)
	}
	expected := filepath.Join(dir, "subdir/file.txt")
	if absPath != expected {
		t.Errorf("expected %s, got %s", expected, absPath)
	}
}

func TestCopy(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	srcFile := filepath.Join(dir, "src_copy.txt")
	testData := []byte("copy content")
	err := os.WriteFile(srcFile, testData, 0644)
	if err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	err = fa.Copy("dst_copy.txt", "src_copy.txt", 0644)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "dst_copy.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("expected %s, got %s", testData, data)
	}
}

func TestValidatePath_PathTraversal(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)

	testCases := []string{
		"../../../etc/passwd",
		"..\\..\\..\\etc\\passwd",
		"subdir/../../etc/passwd",
		"subdir/..%/etc/passwd",
	}

	for _, tc := range testCases {
		err := fa.ValidatePath(tc)
		if err == nil {
			t.Errorf("expected error for path traversal: %s", tc)
		}
	}
}

func TestValidatePath_AbsolutePath(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)

	testCases := []string{
		"/etc/passwd",
		"/home/user/file.txt",
	}

	for _, tc := range testCases {
		err := fa.ValidatePath(tc)
		if err == nil {
			t.Errorf("expected error for absolute path: %s", tc)
		}
	}

	// Windows paths only relevant on Windows
	if runtime.GOOS == "windows" {
		err := fa.ValidatePath("C:\\Windows\\System32")
		if err == nil {
			t.Errorf("expected error for Windows absolute path")
		}
	}
}

func TestValidatePath_NullCharacter(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)

	testCases := []string{
		"file\x00.txt",
		"dir/\x00/file",
	}

	for _, tc := range testCases {
		_, err := fa.GetAbsPath(tc)
		if err == nil {
			t.Errorf("expected error for null character injection: %s", tc)
		}
	}
}

func TestValidatePath_EmptyPath(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)

	_, err := fa.GetAbsPath("")
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestRead_PathTraversalBlocked(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)

	_, err := fa.Read("../../../etc/passwd")
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestWrite_PathTraversalBlocked(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)

	err := fa.Write("../../../etc/passwd", []byte("malicious"), 0644)
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestWorkdir(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	fa := NewFileAccess(dir)
	if fa.Workdir() != filepath.Clean(dir) {
		t.Errorf("expected workdir %s, got %s", filepath.Clean(dir), fa.Workdir())
	}
}
