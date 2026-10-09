# AIOPS Python 算法服务镜像（知识库解析/向量化/检索、日志 Drain 解析）
# PIP_INDEX_URL 可通过 --build-arg 覆盖（如海外服务器用 https://pypi.org/simple）
ARG PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple
FROM python:3.12-slim
ARG PIP_INDEX_URL
WORKDIR /srv
COPY requirements.txt ./
RUN pip install --no-cache-dir -i ${PIP_INDEX_URL} -r requirements.txt
COPY app ./app
EXPOSE 9000
# 2 个 worker：一个 worker 被慢的 embedding 请求占用时，另一个仍可响应日志解析/检索请求
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "9000", "--workers", "2"]
