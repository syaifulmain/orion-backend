# Orion Backend

## System Requirements

- **Go**: Version 1.27 or newer
- **MongoDB**: Version 6.0 or newer

---

## Environment Configuration (`.env`)

Copy the `.env.example` file to `.env`:

```bash
cp .env.example .env
```

Configure the following variables in `.env`:

```env
SERVER_NAME=orion-backend
SERVER_PORT=8080
SERVER_ENV=development
CORS_ALLOW_ORIGINS=*

# API Authentication configuration (Mandatory)
CORPUS_API_SECRET=your_32_bytes_or_longer_secret_key_here
API_KEY_STORE_PATH=data/api_keys.json

# MongoDB configuration
MONGODB_URI=mongodb://admin:166333@localhost:27017/orion_db?authSource=admin
```

> **Note:** When `SERVER_ENV=production`, the application bypasses the `.env` file and reads directly from the environment variables of the host system or container.

---

## Running the Application

### Using Makefile

- **Run development server:**
  ```bash
  make run
  ```

- **Compile binary executable:**
  ```bash
  make build
  ```

- **Tidy Go dependencies:**
  ```bash
  make tidy
  ```

- **Clean build directory:**
  ```bash
  make clean
  ```

- **Generate API Key:**
  ```bash
  # Generate a key valid for 30 days (default)
  make gen-key NAME=client-name

  # Or specify custom validity duration in days
  make gen-key NAME=my-client DAYS=90
  ```

---

## Authentication (Mandatory)

Authentication is **strictly mandatory** for all `/api/corpus/*` endpoints. `CORPUS_API_SECRET` must be configured in `.env` or system environment variables (minimum 32 bytes) for the server to start. Only `/api/health` remains publicly accessible without credentials.

### API Key Details & Structure
- Keys use the `cps_` prefix followed by 22 base64url characters (e.g. `cps_xYEpf3wJFwfQ-L6QeZXv1w`).
- Keys are HMAC-SHA256 authenticated against `CORPUS_API_SECRET`.
- Only HMAC digests and metadata are stored in `data/api_keys.json` (or `API_KEY_STORE_PATH`). The plaintext key is shown only once upon generation.
- Full interoperability with keys created by Python's `speech-corpus-collector`.

### Authorization Headers
Send your generated API key using either of the following HTTP headers:

```http
Authorization: Bearer <api_key>
```
or
```http
X-API-Key: <api_key>
```

If the key is missing, invalid, or expired, the server responds with `401 Unauthorized` and `WWW-Authenticate: Bearer`.

---

## API Documentation

| Method | Endpoint | Authentication | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/health` | No | Check API availability |
| `GET` | `/api/corpus` | Bearer | List, filter, and search records |
| `GET` | `/api/corpus/:id` | Bearer | Get record by ID |
| `GET` | `/api/corpus/export` | Bearer | Stream matching records as JSONL |

---

### 1. Health Check
- **Endpoint:** `GET /api/health`
- **Response:**
  ```json
  {
    "status": "ok"
  }
  ```

---

### 2. List & Filter Corpus
- **Endpoint:** `GET /api/corpus`
- **Headers:** `Authorization: Bearer <api_key>` or `X-API-Key: <api_key>`
- **Query Parameters:**
  | Parameter | Type | Default | Description |
  |---|---|---|---|
  | `q` | String | - | Case-insensitive search on `text` & `source` (Max: 200 characters). |
  | `split` | String | - | Filter subset: `train`, `eval`, or `test`. |
  | `source` | String | - | Filter source (exact match or prefix, e.g., `youtube` matches `youtube` and `youtube/...`). |
  | `license` | String | - | Filter license type (case-insensitive exact match, e.g., `CC-BY-3.0`). |
  | `limit` | Integer | `25` | Number of items per page (Max: `1000`). |
  | `offset` | Integer | `0` | Number of items to skip. |

- **Success Response (200 OK):**
  ```json
  {
    "data": [
      {
        "id": "005ad92162313049802f161b",
        "text": "xxxxxxxxxxxxxxxxxxxxxxxxxx",
        "source": "youtube/xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
        "license": "CC-BY-3.0",
        "split": "eval",
        "crawl_date": "2026-09-03"
      }
    ],
    "pagination": {
      "total": 1446,
      "limit": 25,
      "offset": 0,
      "has_more": true
    }
  }
  ```

- **cURL Example:**
  ```bash
  curl -H "Authorization: Bearer <api_key>" "http://localhost:8080/api/corpus?q=kesehatan&split=eval&limit=10"
  ```

---

### 3. Get Corpus by ID
- **Endpoint:** `GET /api/corpus/:id`
- **Headers:** `Authorization: Bearer <api_key>` or `X-API-Key: <api_key>`
- **Success Response (200 OK):**
  ```json
  {
    "data": {
      "id": "005ad92162313049802f161b",
      "text": "xxxxxxxxxxxxxxxxxxxxxxxxxx",
      "source": "youtube/xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
      "license": "CC-BY-3.0",
      "split": "eval",
      "crawl_date": "2026-09-03"
    }
  }
  ```

- **cURL Example:**
  ```bash
  curl -H "Authorization: Bearer <api_key>" "http://localhost:8080/api/corpus/005ad92162fa3049802f161b"
  ```

---

### 4. Streaming Export JSONL
- **Endpoint:** `GET /api/corpus/export`
- **Headers:**
  - `Authorization: Bearer <api_key>` or `X-API-Key: <api_key>`
  - `Content-Type: application/x-ndjson; charset=utf-8`
  - `Content-Disposition: attachment; filename="corpus-export.jsonl"`
- **Query Parameters:** Supports filter parameters (`q`, `split`, `source`, `license`). *Does not accept pagination parameters (`limit`, `offset`).*

- **cURL Example:**
  ```bash
  curl -N -H "Authorization: Bearer <api_key>" "http://localhost:8080/api/corpus/export?source=youtube" -o corpus-export.jsonl
  ```

---

## Project Structure

```text
orion-backend/
├── cmd/
│   ├── api/
│   │   └── main.go           # Application entry point
│   └── genkey/
│       └── main.go           # CLI utility for generating authenticated API keys
├── internal/
│   ├── auth/                 # API key authentication, HMAC verification & middleware
│   │   ├── auth.go
│   │   └── middleware.go
│   ├── config/
│   │   └── config.go         # Configuration loader (env / .env)
│   ├── corpus/               # Corpus domain module
│   │   ├── handler.go        # Fiber v3 HTTP handlers
│   │   ├── model.go          # Entity struct and query filter DTOs
│   │   ├── repository.go     # MongoDB repository queries and filters
│   │   ├── router.go         # HTTP endpoint route definitions
│   │   └── service.go        # Business logic and input validation
│   ├── database/
│   │   └── mongodb.go        # Official MongoDB Driver v2 connection setup
│   └── server/
│       └── server.go         # Fiber v3 server setup, CORS, Logger, and Recover
├── pkg/
│   └── response/             # Standardized JSON and pagination responses
│       └── response.go
├── Makefile                  # Build and management automation targets
├── eval.jsonl                # Source corpus dataset
├── .env                      # Active environment variables
├── .env.example              # Environment variables template
├── .gitignore
├── go.mod
└── go.sum
```
