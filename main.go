package main

import (
	"database/sql"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form
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

	// Security measure: only allow PDF files by MIME type
	if err := validatePDF(file, handler); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Placeholder for OCR and data extraction logic
	// Placeholder for database logic

	fmt.Fprintf(w, "File %s uploaded and processed successfully!", handler.Filename)
}

func validatePDF(file multipart.File, handler *multipart.FileHeader) error {
	// Check file extension
	if ext := filepath.Ext(handler.Filename); ext != ".pdf" {
		return fmt.Errorf("Only PDF files are allowed")
	}

	// Check MIME type
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil {
		return fmt.Errorf("Unable to read file")
	}
	file.Seek(0, 0) // Reset the file pointer to the beginning

	mimeType := http.DetectContentType(buf)
	if mimeType != "application/pdf" {
		return fmt.Errorf("Only PDF files are allowed")
	}

	return nil
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
        </body>
        </html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func main() {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", uploadPageHandler)
	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
