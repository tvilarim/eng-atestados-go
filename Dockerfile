FROM golang:latest

WORKDIR /app

COPY . .

RUN apt-get update && apt-get install -y tesseract-ocr sqlite3 libsqlite3-dev

# Add the required package
RUN go get github.com/mattn/go-sqlite3

RUN go mod download
RUN go build -o main .

EXPOSE 8080

CMD ["./main"]

