package HttpService

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePublicStaticFileRejectsDirectoryListing(t *testing.T) {
	root := t.TempDir()
	publicFile := filepath.Join(root, "app.js")
	if err := os.WriteFile(publicFile, []byte("console.log('ok')"), 0o600); err != nil {
		t.Fatal(err)
	}
	if resolved, err := resolvePublicStaticFile(root, publicFile, ""); err != nil || resolved != publicFile {
		t.Fatalf("regular file should resolve, path=%q err=%v", resolved, err)
	}

	privateDirectory := filepath.Join(root, "db")
	if err := os.Mkdir(privateDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := resolvePublicStaticFile(root, privateDirectory, ""); err == nil {
		t.Fatal("directory without an index file must not be served")
	}

	publicDirectory := filepath.Join(root, "docs")
	if err := os.Mkdir(publicDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	indexFile := filepath.Join(publicDirectory, "index.html")
	if err := os.WriteFile(indexFile, []byte("<h1>docs</h1>"), 0o600); err != nil {
		t.Fatal(err)
	}
	if resolved, err := resolvePublicStaticFile(root, publicDirectory, ""); err != nil || resolved != indexFile {
		t.Fatalf("directory index should resolve, path=%q err=%v", resolved, err)
	}
}
