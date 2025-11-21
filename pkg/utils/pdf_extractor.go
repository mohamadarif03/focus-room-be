package utils

import (
    "bytes"
    "fmt"
    "io"
    "mime/multipart"

    "github.com/ledongthuc/pdf" // Gunakan library ini
)

func ExtractTextFromPDF(file multipart.File, fileSize int64) (string, error) {
    // 1. Salin file ke bytes buffer karena library butuh ReaderAt
    buf := new(bytes.Buffer)
    if _, err := io.Copy(buf, file); err != nil {
        return "", fmt.Errorf("gagal copy buffer: %w", err)
    }

    // 2. Buat Reader
    readerAt := bytes.NewReader(buf.Bytes())
    r, err := pdf.NewReader(readerAt, int64(buf.Len()))
    if err != nil {
        return "", fmt.Errorf("gagal init pdf reader: %w", err)
    }

    // 3. Baca teks per halaman
    var text string
    totalPage := r.NumPage()

    for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
        p := r.Page(pageIndex)
        if p.V.IsNull() {
            continue
        }
        
        content, err := p.GetPlainText(nil)
        if err != nil {
            // Log error per halaman tapi jangan stop total, lanjut ke halaman berikutnya
            continue 
        }
        text += content + "\n"
    }

    if text == "" {
        return "", fmt.Errorf("teks kosong, kemungkinan PDF berupa gambar/scan")
    }

    return text, nil
}