# Etapa 1: Usar a imagem base do Golang mais recente com suporte a CGO
FROM golang:1.18-alpine AS builder

# Instalar dependências do sistema (compilador C, SQLite3, Tesseract, libc-dev, e linux-headers)
RUN apk add --no-cache gcc musl-dev sqlite-dev tesseract-ocr libc-dev linux-headers

# Definir o diretório de trabalho dentro do container
WORKDIR /app

# Copiar o arquivo go.mod e go.sum para o container
COPY go.mod ./
COPY go.sum ./

# Baixar e instalar as dependências Go
RUN go mod download

# Copiar o código-fonte do projeto para o container
COPY . .

# Compilar o código Go com suporte a CGO (para usar SQLite3)
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
