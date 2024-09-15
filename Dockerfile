# Etapa 1: Usar a imagem base do Golang mais recente com suporte a CGO
FROM golang:1.18-alpine AS builder

# Instalar dependências do sistema (compilador C, SQLite3, Tesseract, Python)
RUN apk add --no-cache gcc musl-dev sqlite-dev tesseract-ocr python3 py3-pip python3-virtualenv

# Definir o diretório de trabalho dentro do container
WORKDIR /app

# Criar um novo módulo Go dentro do container
RUN go mod init eng-atestados-go

# Criar o ambiente virtual Python e ativá-lo
RUN python3 -m venv /app/venv
ENV PATH="/app/venv/bin:$PATH"

# Instalar o pdfplumber no ambiente virtual
RUN pip install pdfplumber

# Copiar o código-fonte do projeto
COPY . .

# Gerar o arquivo go.sum automaticamente ao baixar as dependências
RUN go mod tidy

# Compilar o código Go com suporte a CGO
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Etapa 2: Criar a imagem final com o binário compilado e as dependências
FROM alpine:latest

# Instalar dependências necessárias para rodar o binário e o script Python
RUN apk add --no-cache sqlite-libs tesseract-ocr python3 py3-pip

# Copiar o ambiente virtual Python da etapa de build
COPY --from=builder /app/venv /app/venv

# Definir o ambiente virtual como padrão
ENV PATH="/app/venv/bin:$PATH"

# Definir o diretório de trabalho
WORKDIR /app

# Copiar o binário e o código do Python da etapa de build
COPY --from=builder /app/main /app/
COPY --from=builder /app/extract_text.py /app/

# Expor a porta 8080
EXPOSE 8080

# Comando para rodar a aplicação
CMD ["/app/main"]
