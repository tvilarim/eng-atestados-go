package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"

    _ "github.com/mattn/go-sqlite3"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }
    file, _, err := r.FormFile("pdf")
    if err != nil {
        http.Error(w, "Failed to upload file", http.StatusInternalServerError)
        return
    }
    defer file.Close()

    // Placeholder for OCR and data extraction logic

    // Placeholder for database logic

    fmt.Fprintf(w, "File uploaded and processed successfully!")
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
    // Placeholder for fetching and displaying data from the database
}

func main() {
    db, err := sql.Open("sqlite3", "./data.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    http.HandleFunc("/upload", uploadHandler)
    http.HandleFunc("/data", dataHandler)

    log.Println("Server started on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

