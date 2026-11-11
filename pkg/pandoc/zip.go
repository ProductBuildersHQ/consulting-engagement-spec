package pandoc

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"time"
)

// NormalizeZip rewrites a ZIP container (such as a DOCX) so every entry
// carries the same modified time, t. Some Pandoc versions stamp the files
// they copy from the built-in reference document with the current time,
// which makes otherwise identical DOCX files differ. Entry order, names,
// compression methods, and contents are kept.
func NormalizeZip(data []byte, t time.Time) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("pandoc: read zip: %w", err)
	}
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, f := range r.File {
		hdr := &zip.FileHeader{Name: f.Name, Method: f.Method, Modified: t.UTC()}
		dst, err := w.CreateHeader(hdr)
		if err != nil {
			return nil, fmt.Errorf("pandoc: write %s: %w", f.Name, err)
		}
		src, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("pandoc: open %s: %w", f.Name, err)
		}
		_, copyErr := io.Copy(dst, src) //nolint:gosec // input is the DOCX Pandoc just wrote, bounded by the caller
		closeErr := src.Close()
		if copyErr != nil {
			return nil, fmt.Errorf("pandoc: copy %s: %w", f.Name, copyErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("pandoc: close %s: %w", f.Name, closeErr)
		}
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("pandoc: finish zip: %w", err)
	}
	return buf.Bytes(), nil
}
