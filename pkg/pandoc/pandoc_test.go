package pandoc

import (
	"archive/zip"
	"bytes"
	"errors"
	"slices"
	"testing"
	"time"
)

func TestCommand(t *testing.T) {
	args, env, err := Command(Options{Input: "in.md", Output: "out.pdf", Target: TargetPDF, Date: "2026-03-02"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(args, "--pdf-engine="+DefaultPDFEngine) || args[len(args)-1] != "in.md" {
		t.Fatalf("args = %v", args)
	}
	if !slices.Contains(env, "SOURCE_DATE_EPOCH=1772409600") || !slices.Contains(env, "FORCE_SOURCE_DATE=1") {
		t.Fatalf("env = %v", env)
	}

	args, _, err = Command(Options{Input: "in.md", Output: "out.docx", Target: TargetDOCX, ReferenceDoc: "ref.docx"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(args, "--to=docx") || !slices.Contains(args, "--reference-doc=ref.docx") {
		t.Fatalf("args = %v", args)
	}
}

func TestCommandErrors(t *testing.T) {
	if _, _, err := Command(Options{Input: "in.md", Output: "out.html", Target: "html"}); !errors.Is(err, ErrUnknownTarget) {
		t.Fatalf("error = %v", err)
	}
	if _, _, err := Command(Options{Input: "in.md", Output: "out.pdf", Target: TargetPDF, Date: "March 2"}); err == nil {
		t.Fatal("want a date error")
	}
	if _, err := TargetFromPath("x.txt"); !errors.Is(err, ErrUnknownTarget) {
		t.Fatalf("error = %v", err)
	}
}

func TestNormalizeZip(t *testing.T) {
	build := func(mod time.Time) []byte {
		var buf bytes.Buffer
		w := zip.NewWriter(&buf)
		for _, name := range []string{"a.xml", "b.xml"} {
			f, err := w.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate, Modified: mod})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Write([]byte("<x>" + name + "</x>")); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}
	pin := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	a, err := NormalizeZip(build(time.Date(2026, 3, 2, 11, 45, 52, 0, time.UTC)), pin)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NormalizeZip(build(time.Date(2026, 3, 2, 11, 45, 53, 0, time.UTC)), pin)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("normalized archives differ")
	}
	r, err := zip.NewReader(bytes.NewReader(a), int64(len(a)))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.File) != 2 || r.File[0].Name != "a.xml" || !r.File[1].Modified.Equal(pin) {
		t.Fatalf("unexpected entries: %+v", r.File)
	}
	if _, err := NormalizeZip([]byte("not a zip"), pin); err == nil {
		t.Fatal("want an error")
	}
}
