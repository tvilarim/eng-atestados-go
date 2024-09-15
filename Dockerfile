# Etapa 1: Usar a imagem base do Golang mais recente com suporte a CGO
FROM golang:latest as builder

# Definir o diretório de trabalho dentro do container
WORKDIR /app

# Copiar os arquivos go.mod e go.sum e instalar dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar o código-fonte do projeto
COPY . .

# Instalar dependências do sistema (compilador C, SQLite3)
RUN apt-get update && apt-get install -y \
    build-essential \
    libsqlite3-dev \
    tesseract-ocr \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Instalar o pdfplumber usando pip
RUN pip3 install pdfplumber

# Compilar o código Go com suporte a CGO
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Etapa 2: Criar a imagem final com o binário compilado e as dependências
FROM debian:buster

# Instalar dependências necessárias para rodar o binário e o script Python
RUN apt-get update && apt-get install -y \
    libsqlite3-dev \
    tesseract-ocr \
    python3 \
    && rm -rf /var/lib/apt/lists/*

# Instalar o pdfplumber usando pip
RUN pip3 install pdfplumber

# Definir o diretório de trabalho
WORKDIR /app

# Copiar o binário e o código do Python da etapa de build
COPY --from=builder /app/main /app/
COPY --from=builder /app/extract_text.py /app/

# Expor a porta 8080
EXPOSE 8080

# Comando para rodar a aplicação
CMD ["/app/main"]
