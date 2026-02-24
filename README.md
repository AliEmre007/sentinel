# 🛡️ Sentinel

A high-performance, containerized URL Shortener microservice built with Go and PostgreSQL. 

Engineered with a strict adherence to modern backend best practices, the Twelve-Factor App methodology, and zero-downtime DevOps principles.

## 🏗️ Architecture & Tech Stack

* **Language:** Go (Golang 1.24+) utilizing modern native routing.
* **Database:** PostgreSQL (Relational persistence with parameterized queries).
* **Development Environment:** Docker Compose with [Air](https://github.com/cosmtrek/air) for live-reloading.
* **Production Build:** Multi-stage Dockerfile producing a distroless **6.65MB** static binary in a `scratch` container.
* **Testing:** Native Go `testing` and `httptest` packages for microsecond-fast unit and integration testing.

## ✨ Key Features

* **Collision-Resistant Generation:** Cryptographically safe random 6-character short codes using `math/rand/v2`.
* **Fail-Fast Configuration:** Environment variable injection (`.env`) for secure secrets management.
* **Strict Type Safety:** Robust JSON validation and schema enforcement.
* **Idempotent Migrations:** Auto-verifying database tables on server startup.
* **Micro-Footprint:** 98.7% image size reduction from Dev (511MB) to Prod (6.6MB) for instantaneous cloud scaling.

---

## 🚀 Quick Start (Local Development)

### 1. Set Up the Environment Vault
Create a `.env` file in the root directory to store your database credentials safely:
```env
DB_USER=admin
DB_PASSWORD=secretpassword
DB_NAME=sentinel_db
DATABASE_URL=postgres://admin:secretpassword@db:5432/sentinel_db?sslmode=disable

2. Boot the Infrastructure

Use Docker Compose to spin up the PostgreSQL database and the Go API with live-reloading enabled.
docker compose up -d
The API will be available on http://localhost:8080.

📖 API Reference
1. Health Check

Verify the API is running and responding.

    Endpoint: GET /health

    Response: 200 OK

curl http://localhost:8080/health

2. Shorten a URL

Submit a long URL to generate a unique 6-character short code.

    Endpoint: POST /shorten

    Payload: JSON containing original_url

    Response: 201 Created

curl -X POST http://localhost:8080/shorten \
-H "Content-Type: application/json" \
-d '{"original_url": "[https://www.sakarya.edu.tr](https://www.sakarya.edu.tr)"}'

# Returns: {"short_code":"UkBjqG","original_url":"[https://www.sakarya.edu.tr](https://www.sakarya.edu.tr)"}


3. Redirect to Original URL

Visit the short code in your browser, and the API will issue an HTTP 302 Found to redirect you.

    Endpoint: GET /{code}

    Example: http://localhost:8080/UkBjqG


🧪 Testing

Sentinel uses Go's native testing suite to ensure business logic and HTTP handlers function perfectly in memory, without network hardware overhead.

To execute the test suite inside the running API container:

docker exec -it sentinel-api go test -v

📦 Production Deployment

To build the ultra-secure, standalone machine-code image for cloud deployment (AWS, GCP, Kubernetes):


docker build -f Dockerfile.prod -t sentinel:production .


***

Once you have this saved in your directory, your local project is 100% complete. 

**Would you like me to walk you through the exact `git` commands to initialize this repository, commit this code (while safely ignoring the `.env` file), and push it up to GitHub?**


