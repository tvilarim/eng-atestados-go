# Etapa 1: Usar a imagem base do Golang mais recente com suporte a CGO
FROM golang:1.18-alpine AS builder

# Instalar dependências do sistema (compilador C, SQLite3, Tesseract)
RUN apk add --no-cache gcc musl-dev sqlite-dev tesseract-ocr

# Definir o diretório de trabalho dentro do container
WORKDIR /app

# Copiar o arquivo go.mod e gerar o go.sum diretamente no container
COPY go.mod ./
COPY go.sum ./

# Baixar dependências e instalar módulos Go
RUN go mod tidy
RUN go get github.com/mattn/go-sqlite3
RUN go get github.com/unidoc/unipdf/v3/extractor
RUN go get github.com/unidoc/unipdf/v3/model

# Copiar o código-fonte do projeto para o container
COPY . .

# Compilar o código Go com suporte a CGO
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Etapa 2: Criar a imagem final com o binário compilado e as dependências
FROM alpine:latest

# Instalar dependências necessárias para rodar o binário
RUN apk add --no-cache sqlite-libs tesseract-ocr

# Definir o diretório de trabalho
WORKDIR /app

# Copiar o binário da etapa de build
COPY --from=builder /app/main /app/

# Expor a porta 8080
EXPOSE 8080

# Comando para rodar a aplicação
CMD ["/app/main"]
