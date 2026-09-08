package MarsService

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSmokeProperties(t *testing.T) {
	name := filepath.Join(t.TempDir(), "config.json")
	original := []byte(`{"enabled":true}`)
	if err := writePropertiesFile(name, original); err != nil {
		t.Fatal(err)
	}
	s := &MarsService{PropertyFileName: name}
	for _, invalid := range []string{"broken", "null", "[]"} {
		s.ModifyProperties(invalid)
	}
	got, err := os.ReadFile(name)
	if err != nil || string(got) != string(original) {
		t.Fatalf("設定遭覆蓋: %q %v", got, err)
	}
}
