# AIOPS Intelligent Operations Platform

[中文](README.md) | English

An intelligent operations (AIOps) platform for large-scale distributed systems. It unifies multi-source observability data — Prometheus metrics, ELK logs, and Pyroscope profiling — and combines it with LLM-powered agents to deliver conversational querying, alert rule evaluation, alert event management, alert noise reduction, scheduled inspection reports, root cause analysis, and an operations knowledge base.

## Features

- **AI Chat Assistant**: An LLM agent with a Function Calling loop that invokes read-only data source tools (Prometheus / ElasticSearch / Pyroscope). When ops-data keywords are detected, querying data sources is enforced before answering to prevent hallucination.
- **Alert Rule Evaluation Engine**: A built-in engine probes PromQL instant queries at each rule's execution interval. Every time series (label combination) is counted independently; alert/recovery events are generated once the rule's duration threshold is met. No Nightingale dependency required.
- **Alert Noise Reduction**: Two strategies — time-window aggregation (flap suppression) and topology suppression (child service alerts suppressed when an ancestor in the service tree is firing). Intercepted alerts are persisted separately for auditing.
- **Notification Records**: The full notification lifecycle (intercepted / no rule matched / severity filtered / delivery succeeded or failed) is persisted and traceable.
- **Dashboard**: KPIs, raw vs. post-denoise trends, severity distribution, service health scores, and more.
- **Root Cause Analysis**: An asynchronous Eino Graph workflow that collects evidence first, then synthesizes a structured conclusion (evidence chain / confidence / remediation suggestions).
- **Log Analysis**: Wait to implement.
- **Scheduled Inspections**: Cron-based scheduling with LLM-generated Markdown reports, delivered via DingTalk bot or Webhook.
- **Operations Knowledge Base**: Document parsing, chunking, and vectorization into Qdrant; the chat agent can search and cite it (optional component).
- **Global Search / i18n (Chinese, English) / RBAC**: User, role, and permission seeding.

## Tech Stack

| Layer | Technologies |
|-------|--------------|
| Frontend | Vue 3, Vite, ECharts, Vue Router (hash mode) |
| Backend | Go, Gin, GORM, Viper, CloudWeGo Eino, robfig/cron, JWT |
| Algorithm Service (optional) | Python, FastAPI, Qdrant Client |
| Storage | MySQL 8.0+, Qdrant (optional) |
| Monitoring Data Sources | Prometheus, ElasticSearch, Pyroscope |

## Project Structure

```
aiops/
├── backend/                 # Go backend service
│   ├── cmd/center/          # Entrypoint (routes, AutoMigrate, engine/scheduler startup)
│   ├── configs/             # Viper configuration (config.yaml)
│   ├── controllers/         # HTTP controllers
│   ├── internal/            # agent / alerting / chat / datasource / inspection / rca / notify / knowledge
│   ├── middleware/          # JWT auth middleware (SSE supports query token)
│   └── models/              # GORM models
├── frontend/                # Vue 3 + Vite frontend
│   ├── src/api/             # API client layer (JWT, 401 redirect to login)
│   ├── src/views/           # Page views
│   └── vite.config.js       # Dev proxy: /api -> localhost:8080
├── algorithm/               # Python algorithm service (KB parsing/embedding/retrieval, optional)
├── qdrant/                  # Qdrant deployment files (docker-compose.yml + config.yaml)
├── nginx.conf               # Reference Nginx config for production
└── sql/                     # Database initialization script (includes root user seed)
```

## Quick Start

### Prerequisites

| Dependency | Version |
|------------|---------|
| Go | 1.26+ |
| MySQL | 8.0+ |
| Node.js | 18+ |
| Docker | Required only for the knowledge base |
| Python | 3.10+, required only for the knowledge base |

### 1. Initialize the Database

```bash
mysql -u root -p -e "CREATE DATABASE aiops DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p aiops < sql/init.sql
```

`sql/init.sql` creates all table structures and seeds the default user, roles, permissions, notification templates, and noise-reduction policies.

> ⚠️ The default admin account is `root / root` (stored in plaintext). Change the password in **User Management** immediately after your first login.

### 2. Configure and Start the Backend

Edit `backend/configs/config.yaml` and set `database.dsn` to your MySQL connection string:

```yaml
database:
  dsn: "root:YOUR_PASSWORD@tcp(localhost:3306)/aiops?charset=utf8mb4&parseTime=True&loc=Local"
server:
  port: ":8080"
app:
  # Frontend URL used for the "full report" link in inspection notifications
  frontend_base_url: "http://localhost:5173"
knowledge:
  python_base_url: "http://localhost:9000"
  upload_dir: "./uploads/knowledge"
```

Configuration can also be overridden via environment variables (recommended for production, to avoid committing credentials):

| Environment Variable | Description | Default |
|----------------------|-------------|---------|
| `AIOPS_DATABASE_DSN` / `DATABASE_DSN` | MySQL DSN | See config.yaml |
| `AIOPS_SERVER_PORT` / `SERVER_PORT` | Server port | `:8080` |
| `AIOPS_APP_FRONTEND_BASE_URL` / `FRONTEND_BASE_URL` | Frontend URL | `http://localhost:5173` |

Start the backend:

```bash
cd backend
go mod tidy
go run ./cmd/center
```

GORM AutoMigrate (tables / columns) runs automatically on startup, and alert rules are backfilled with a default execution interval.

### 3. Start the Frontend

```bash
cd frontend
npm install
npm run dev        # Dev server at http://localhost:5173, /api proxied to http://localhost:8080
```

Open http://localhost:5173 and log in with `root / root`.

### 4. Configure an LLM

After logging in, go to the **LLM Config** page and add an OpenAI-compatible chat model, then set it as the default (used by the chat assistant, root cause analysis, and inspection reports). Changes take effect immediately without a restart.

## Optional: Enable the Knowledge Base

The knowledge base requires three components: **Qdrant**, the **Python parsing/embedding service**, and a **default embedding model configuration**.

```bash
# 1. Start Qdrant (config: qdrant/config.yaml; dashboard at http://<ip>:6333/dashboard)
cd qdrant
docker compose up -d

# 2. Start the Python service (port 9000; the Go backend connects to it by default)
cd ../algorithm
pip install -r requirements.txt
uvicorn app.main:app --port 9000

# 3. In "LLM Config", add a config with model_type=embedding and set it as default
```

The embedding model must expose an **Embeddings API** (any OpenAI-compatible provider works). Recommendations:

| Provider | Base URL | Model | Notes |
|----------|----------|-------|-------|
| Alibaba Cloud Model Studio | `https://dashscope.aliyuncs.com/compatible-mode/v1` | `text-embedding-v3` | Direct access in China, strong Chinese performance, pay-as-you-go |
| SiliconFlow | `https://api.siliconflow.cn/v1` | `BAAI/bge-m3` | Free quota for individual developers |
| OpenAI | `https://api.openai.com/v1` | `text-embedding-3-small` | Requires network access to OpenAI and an API key |

> Note: **DeepSeek's official API does not provide an Embeddings endpoint**, and Kimi coding subscription keys cannot be used for embedding either — misconfiguration causes document indexing to fail (see `kb_document.error_msg` for the error). Embedding dimensions vary by model (768/1024/1536); the collection is created automatically with the actual dimension of the model in use on first ingestion. If you switch models, clear the Qdrant data and re-upload your documents.

## Production Deployment

1. **Frontend build**: `cd frontend && npm run build`, then publish the contents of `frontend/dist/` to your Nginx site directory.
2. **Backend build**: `cd backend && go build -o center ./cmd/center`, then run the binary with configuration injected via environment variables.
3. **Nginx**: see `nginx.conf` in the repository root:

   ```nginx
   root /opt/aiops/web;                    # Frontend dist directory
   location /api/ {
       proxy_pass http://127.0.0.1:8080;   # Backend service
       proxy_buffering off;                # Required for SSE streaming
       proxy_read_timeout 600s;
       proxy_set_header X-Accel-Buffering no;
       # ... see nginx.conf for the rest
   }
   ```

   SSE (chat streaming, RCA polling) relies on `proxy_buffering off` and `X-Accel-Buffering no` — do not omit them.
4. **Knowledge base (optional)**: Deploy Qdrant (`cd qdrant && docker compose up -d`) and the Python service (uvicorn) on the server. Adjust the volume mount path in `qdrant/docker-compose.yml` (`/data/qdrant:/qdrant/storage`) to match your server layout.
5. **Frontend URL**: Set `AIOPS_APP_FRONTEND_BASE_URL` to the real frontend address so links in inspection report notifications are reachable.

## Testing

```bash
cd algorithm && pytest      # Python algorithm service unit tests
cd backend && go test ./... # Go tests (as added)
```

## Security Notes

- `backend/configs/config.yaml` ships with a placeholder DSN. **Inject the database password via environment variables in production** — never commit real credentials to the repository.
- LLM configuration API keys are stored in the database and masked in API responses.
- The default `root / root` account is for initial setup only — change it immediately.
- JWT tokens are stored in frontend localStorage; enable HTTPS in production deployments.

## License

[Apache License 2.0](LICENSE)
