def split_text(text: str, chunk_size: int = 500, chunk_overlap: int = 50) -> list[str]:
    """按行边界优先、固定长度兜底的方式分段，相邻块保留重叠"""
    text = text.strip()
    if not text:
        return []

    lines = text.split("\n")
    chunks: list[str] = []
    current = ""
    for line in lines:
        candidate = line if not current else current + "\n" + line
        if len(candidate) <= chunk_size:
            current = candidate
            continue
        if current:
            chunks.append(current)
        # 超长单行直接硬切
        while len(line) > chunk_size:
            chunks.append(line[:chunk_size])
            line = line[chunk_size - chunk_overlap:]
        current = line
    if current:
        chunks.append(current)

    # 尾部重叠：将下一块开头 chunk_overlap 字符拼到上一块末尾（近似实现，保证信息不丢）
    overlapped: list[str] = []
    for i, c in enumerate(chunks):
        if i > 0 and chunk_overlap > 0:
            prev_tail = chunks[i - 1][-chunk_overlap:]
            c = prev_tail + c
        overlapped.append(c[:chunk_size])
    return [c for c in overlapped if c.strip()]
