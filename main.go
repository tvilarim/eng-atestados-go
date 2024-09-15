package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB max size
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if err := validatePDF(file, handler); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Salvar o arquivo PDF temporariamente
	tempFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		http.Error(w, "Unable to save the file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name()) // Limpar o arquivo após o uso

	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Failed to save the file", http.StatusInternalServerError)
		return
	}

	// Chamar o script Python para extrair o texto
	extractedText, err := extractTextFromPDF(tempFile.Name())
	if err != nil {
		http.Error(w, "Failed to extract text from the PDF", http.StatusInternalServerError)
		return
	}

	// Salvar o texto extraído no banco de dados SQLite
	err = saveExtractedText(handler.Filename, extractedText)
	if err != nil {
		http.Error(w, "Failed to save extracted text to the database", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File %s uploaded and processed successfully!", handler.Filename)
}

func extractTextFromPDF(pdfPath string) (string, error) {
	cmd := exec.Command("python3", "extract_text.py", pdfPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("text extraction failed: %v", err)
	}
	return string(output), nil
}

func saveExtractedText(filename, text string) error {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS documents (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        filename TEXT,
        content TEXT
    )`)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO documents (filename, content) VALUES (?, ?)", filename, text)
	if err != nil {
		return fmt.Errorf("failed to insert data into the database: %v", err)
	}

	return nil
}

func showContentHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		http.Error(w, "Failed to connect to the database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT filename, content FROM documents")
	if err != nil {
		http.Error(w, "Failed to query the database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var htmlContent strings.Builder
	htmlContent.WriteString("<html><body><h1>Extracted PDF Contents</h1>")
	for rows.Next() {
		var filename, content string
		rows.Scan(&filename, &content)
		htmlContent.WriteString("<h2>" + filename + "</h2>")
		htmlContent.WriteString("<pre>" + content + "</pre>")
	}
	htmlContent.WriteString("</body></html>")

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlContent.String()))
}

func validatePDF(file multipart.File, handler *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".pdf" {
		return fmt.Errorf("Only PDF files are allowed")
	}

	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil {
		return fmt.Errorf("Unable to read file")
	}
	file.Seek(0, 0)

	mimeType := http.DetectContentType(buf)
	if mimeType != "application/pdf" {
		return fmt.Errorf("Only PDF files are allowed")
	}

	return nil
}

func main() {
	http.HandleFunc("/", uploadPageHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/content", showContentHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func uploadPageHandler(w http.ResponseWriter, r *http.Request) {
	html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Upload PDF</title>
        </head>
        <body>
            <h1>Upload PDF File</h1>
            <form enctype="multipart/form-data" action="/upload" method="post">
                <input type="file" name="pdf" accept="application/pdf" required>
                <input type="submit" value="Upload PDF">
            </form>
            <h2><a href="/content">View Extracted Content</a></h2>
        </body>
        </html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}
