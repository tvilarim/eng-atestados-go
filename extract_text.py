import sys
import pdfplumber

def extract_text(pdf_path):
    with pdfplumber.open(pdf_path) as pdf:
        text = ""
        for page in pdf.pages:
            text += page.extract_text() + "\n"
    return text

if __name__ == "__main__":
    pdf_path = sys.argv[1]
    text = extract_text(pdf_path)
    print(text)
