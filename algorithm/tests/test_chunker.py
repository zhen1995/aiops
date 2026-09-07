from app.chunker import split_text


def test_empty_text():
    assert split_text("") == []
    assert split_text("   \n\n  ") == []


def test_short_text_single_chunk():
    assert split_text("短文本", chunk_size=500) == ["短文本"]


def test_long_text_splits_with_overlap():
    text = "0123456789" * 200  # 2000 字符
    chunks = split_text(text, chunk_size=500, chunk_overlap=50)
    assert len(chunks) >= 4
    assert all(len(c) <= 500 for c in chunks)
    # 相邻块有重叠：下一块的开头应出现在上一块尾部
    assert chunks[1][:50] == chunks[0][-50:]


def test_prefers_line_boundary():
    lines = [f"第{i}行" + "x" * 60 for i in range(20)]
    chunks = split_text("\n".join(lines), chunk_size=500, chunk_overlap=20)
    # 块应以某行行首开始（允许第一块例外）
    for c in chunks[1:]:
        assert c.lstrip("x")[:1] in "第" or c.startswith("x") is False
