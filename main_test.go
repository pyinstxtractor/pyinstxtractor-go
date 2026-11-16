//go:build !gopherjs

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func extractEntries(t *testing.T, entryType byte, names ...string) string {
	t.Helper()
	workDirectory := t.TempDir()
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workDirectory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(originalDirectory) })

	payload := []byte("payload")
	archivePath := filepath.Join(workDirectory, "sample.exe")
	if err := os.WriteFile(archivePath, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archiveFile.Close()

	archive := PyInstArchive{
		inFilePath:         archivePath,
		fPtr:               archiveFile,
		pythonMajorVersion: 3,
		pythonMinorVersion: 11,
		gotPycMagic:        true,
	}
	for _, name := range names {
		archive.tableOfContents = append(archive.tableOfContents, CTOCEntry{
			DataSize:             uint(len(payload)),
			UncompressedDataSize: uint(len(payload)),
			TypeCompressedData:   entryType,
			Name:                 name,
		})
	}
	archive.ExtractFiles()
	return workDirectory
}

func assertMissing(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Lstat(path); err == nil {
			t.Errorf("wrote outside the extraction directory: %s", path)
		}
	}
}

func assertExists(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Lstat(path); err != nil {
			t.Errorf("expected extracted file %s: %v", path, err)
		}
	}
}

func TestExtractFilesKeepsRawEntriesInsideExtractionDirectoryWhenNamesEscape(t *testing.T) {
	outsideDirectory := t.TempDir()
	absoluteName := filepath.Join(outsideDirectory, "absolute")

	workDirectory := extractEntries(t, 'b', absoluteName, "../../dot_dot")

	extractionDirectory := filepath.Join(workDirectory, "sample.exe_extracted")
	assertMissing(t, absoluteName, filepath.Join(filepath.Dir(workDirectory), "dot_dot"))
	assertExists(
		t,
		filepath.Join(extractionDirectory, strings.TrimLeft(absoluteName, string(os.PathSeparator))),
		filepath.Join(extractionDirectory, "__", "__", "dot_dot"),
	)
}

func TestExtractFilesKeepsPycEntriesInsideExtractionDirectoryWhenNamesEscape(t *testing.T) {
	outsideDirectory := t.TempDir()
	absoluteName := filepath.Join(outsideDirectory, "absolute")

	workDirectory := extractEntries(t, 's', absoluteName, "../../dot_dot")

	extractionDirectory := filepath.Join(workDirectory, "sample.exe_extracted")
	assertMissing(t, absoluteName+".pyc", filepath.Join(filepath.Dir(workDirectory), "dot_dot.pyc"))
	assertExists(
		t,
		filepath.Join(extractionDirectory, strings.TrimLeft(absoluteName, string(os.PathSeparator))+".pyc"),
		filepath.Join(extractionDirectory, "__", "__", "dot_dot.pyc"),
	)
}

func TestExtractFilesKeepsEntryNamesWhenTheyAreAlreadySafe(t *testing.T) {
	workDirectory := extractEntries(t, 'b', "lib/python3.11/os.pyc", "base_library.zip")

	extractionDirectory := filepath.Join(workDirectory, "sample.exe_extracted")
	assertExists(
		t,
		filepath.Join(extractionDirectory, "lib", "python3.11", "os.pyc"),
		filepath.Join(extractionDirectory, "base_library.zip"),
	)
}
