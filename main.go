package main

import (
	"database/sql"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

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

	// Validate the uploaded file
	if err := validatePDF(file, handler); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Placeholder for OCR and data extraction logic
	// Here we assume that the content extraction is a simple string for demonstration
	extractedContent := "This is a placeholder for the extracted content from the PDF."

	// Store the extracted content in the database
	_, err = db.Exec("INSERT INTO pdf_contents (filename, content) VALUES (?, ?)", handler.Filename, extractedContent)
	if err != nil {
		http.Error(w, "Failed to store content in the database", http.StatusInternalServerError)
		return
	}

	// Redirect to the content display page
	http.Redirect(w, r, "/show", http.StatusSeeOther)
}

func validatePDF(file multipart.File, handler *multipart.FileHeader) error {
	// Normalize the file extension to lowercase
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".pdf" {
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

	// Check for the PDF header
	if !isPDFHeader(buf) {
		return fmt.Errorf("The file does not appear to be a valid PDF")
	}

	return nil
}

func isPDFHeader(buf []byte) bool {
	// Check if the file starts with '%PDF-' which is the standard PDF header
	return len(buf) >= 4 && string(buf[:4]) == "%PDF"
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

func showContentHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT filename, content FROM pdf_contents ORDER BY id DESC LIMIT 1")
	if err != nil {
		http.Error(w, "Failed to retrieve content from the database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var filename, content string
	if rows.Next() {
		err := rows.Scan(&filename, &content)
		if err != nil {
			http.Error(w, "Failed to scan database row", http.StatusInternalServerError)
			return
		}
	}

	html := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>PDF Content</title>
        </head>
        <body>
            <h1>Content of %s</h1>
            <p>%s</p>
        </body>
        </html>`, filename, content)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the table for storing PDF contents
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS pdf_contents (id INTEGER PRIMARY KEY, filename TEXT, content TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", uploadPageHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/show", showContentHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
