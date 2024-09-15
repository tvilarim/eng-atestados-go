package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Salvar o arquivo PDF temporariamente
	file, _, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		http.Error(w, "Unable to save the file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Failed to save the file", http.StatusInternalServerError)
		return
	}

	// Extrair o texto do PDF
	extractedText, err := extractTextFromPDF(tempFile.Name())
	if err != nil {
		http.Error(w, "Failed to extract text from the PDF", http.StatusInternalServerError)
		return
	}

	// Salvar o texto extraído no banco de dados
	err = saveExtractedText("filename.pdf", extractedText)
	if err != nil {
		http.Error(w, "Failed to save extracted text to the database", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded and processed successfully!")
}

func extractTextFromPDF(pdfPath string) (string, error) {
	f, err := os.Open(pdfPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	reader, err := model.NewPdfReader(f)
	if err != nil {
		return "", err
	}

	var extractedText string
	numPages, err := reader.GetNumPages()
	if err != nil {
		return "", err
	}

	for i := 1; i <= numPages; i++ {
		page, err := reader.GetPage(i)
		if err != nil {
			return "", err
		}

		extractor := extractor.New(page)
		text, err := extractor.ExtractText()
		if err != nil {
			return "", err
		}
		extractedText += text
	}

	return extractedText, nil
}

func saveExtractedText(filename, text string) error {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS documents (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        filename TEXT,
        content TEXT
    )`)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO documents (filename, content) VALUES (?, ?)", filename, text)
	return err
}

func main() {
	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
