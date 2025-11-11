package utils

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"

	"github.com/dslipak/pdf"
)

func ExtractTextFromPDF(file multipart.File, fileSize int64) (string, error) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return "", err
	}

	readerAt := bytes.NewReader(buf.Bytes())
	r, err := pdf.NewReader(readerAt, fileSize)
	if err != nil {
		return "", err
	}

	var allText strings.Builder
	numPages := r.NumPage()

	for i := 1; i <= numPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		allText.WriteString(text)
		allText.WriteString("\n")
	}

	return allText.String(), nil
}
