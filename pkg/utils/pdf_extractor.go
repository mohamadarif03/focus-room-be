package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/ledongthuc/pdf"
)

func ExtractTextFromPDF(file multipart.File, fileSize int64) (string, error) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return "", fmt.Errorf("gagal menyalin file ke buffer: %w", err)
	}

	readerAt := bytes.NewReader(buf.Bytes())

	r, err := pdf.NewReader(readerAt, fileSize)
	if err != nil {
		return "", fmt.Errorf("gagal init pdf reader: %w", err)
	}

	var textBuilder string

	numPages := r.NumPage()
	for i := 1; i <= numPages; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}

		content, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}

		textBuilder += content + "\n"
	}

	if textBuilder == "" {
		return "", fmt.Errorf("gagal mengekstrak teks: hasil kosong (mungkin PDF berupa gambar/scan)")
	}

	return textBuilder, nil
}
