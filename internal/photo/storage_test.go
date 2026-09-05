package photo

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidPhotoPathRejectsEscape(t *testing.T) {
	if _, err := ValidPhotoPath(`C:\outside\photo.jpg`, `C:\photos`); err == nil {
		t.Fatal("expected path escape to be rejected")
	}
}

func TestSafeUser(t *testing.T) {
	if got, want := safeUser(" Prem R. #7 "), "PremR7"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := safeUser("!@#$"); got != "USER" {
		t.Fatalf("got fallback %q, want USER", got)
	}
}

func TestNextSequenceIgnoresMalformedNames(t *testing.T) {
	folder := t.TempDir()
	prefix := "20260903_103712_Prem_"
	for _, name := range []string{prefix + "0001.jpg", prefix + "0007.jpg", prefix + "draft.jpg"} {
		if err := os.WriteFile(filepath.Join(folder, name), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	sequence, err := nextSequence(folder, prefix)
	if err != nil {
		t.Fatal(err)
	}
	if sequence != 8 {
		t.Fatalf("got sequence %d, want 8", sequence)
	}
}

func TestListReturnsNewestPhotoFirst(t *testing.T) {
	folder := t.TempDir()
	oldPath := filepath.Join(folder, "old.jpg")
	newPath := filepath.Join(folder, "new.JPG")
	for _, path := range []string{oldPath, newPath, filepath.Join(folder, "ignored.txt")} {
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	oldTime := time.Now().Add(-time.Hour)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	items, err := NewService().List(folder)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Name != "new.JPG" || items[1].Name != "old.jpg" {
		t.Fatalf("unexpected items: %#v", items)
	}
}

func TestDeleteManyRemovesSelectedPhotos(t *testing.T) {
	folder := t.TempDir()
	deletePath := filepath.Join(folder, "delete.jpg")
	keepPath := filepath.Join(folder, "keep.jpg")
	for _, path := range []string{deletePath, keepPath} {
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := NewService().DeleteMany([]string{deletePath}, folder); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(deletePath); !os.IsNotExist(err) {
		t.Fatalf("delete target still exists: %v", err)
	}
	if _, err := os.Stat(keepPath); err != nil {
		t.Fatalf("expected unselected file to remain: %v", err)
	}
}
