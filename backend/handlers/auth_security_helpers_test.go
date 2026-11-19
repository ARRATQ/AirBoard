package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"net/textproto"
	"testing"
)

// buildFileHeader construit un *multipart.FileHeader comme celui que reçoit UploadAvatar.
func buildFileHeader(t *testing.T, declaredType string, content []byte) *multipart.FileHeader {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="avatar"; filename="a.png"`)
	h.Set("Content-Type", declaredType)
	part, err := w.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatal(err)
	}
	return req.MultipartForm.File["avatar"][0]
}

func TestSniffImageType(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	jpeg := []byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00")
	html := []byte("<html><script>alert(1)</script></html>")

	tests := []struct {
		name         string
		declaredType string
		content      []byte
		wantType     string
		wantErr      bool
	}{
		{"real PNG", "image/png", png, "image/png", false},
		{"real JPEG declared as PNG uses the real type", "image/png", jpeg, "image/jpeg", false},
		{"HTML disguised as PNG is rejected", "image/png", html, "", true},
		{"empty file is rejected", "image/png", nil, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sniffImageType(buildFileHeader(t, tt.declaredType, tt.content))
			if (err != nil) != tt.wantErr {
				t.Fatalf("sniffImageType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantType {
				t.Errorf("sniffImageType() = %q, want %q", got, tt.wantType)
			}
		})
	}
}
