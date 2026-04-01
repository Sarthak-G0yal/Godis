FROM python:3.11-slim

WORKDIR /app

RUN pip install --no-cache-dir uv

COPY pyproject.toml uv.lock .python-version ./
RUN uv sync --frozen --no-dev

COPY . .

EXPOSE 8501

ENTRYPOINT ["uv", "run", "--no-sync", "streamlit", "run", "app.py", "--server.port=8501", "--server.address=0.0.0.0"]
