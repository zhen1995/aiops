import io

from docx import Document as DocxDocument
from pypdf import PdfReader


def parse_document(filename: str, data: bytes) -> str:
    """将文档字节解析为纯文本，按扩展名分发；不支持的类型抛 ValueError"""
    ext = filename.rsplit(".", 1)[-1].lower() if "." in filename else ""
    if ext in ("md", "markdown", "txt"):
        return data.decode("utf-8", errors="ignore")
    if ext == "docx":
        doc = DocxDocument(io.BytesIO(data))
        return "\n".join(p.text for p in doc.paragraphs if p.text.strip())
    if ext == "pdf":
        reader = PdfReader(io.BytesIO(data))
        return "\n".join(page.extract_text() or "" for page in reader.pages)
    raise ValueError(f"不支持的文件类型: {ext or filename}")
