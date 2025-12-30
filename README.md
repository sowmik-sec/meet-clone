# Meet Clone

## Structure

- `frontend/`: Next.js 14 Application (React, Tailwind, Cloudflare RealtimeKit)
- `backend/`: Go Modular Monolith (Gin, MongoDB, Hexagonal Architecture)

## Setup

1. **Start Infrastructure**
   ```bash
   docker-compose up -d
   ```

2. **Backend**
   ```bash
   cd backend
   # Copy .env.example if you have one, or set env vars
   export MONGODB_URL="mongodb://admin:password@localhost:27017"
   export SECRET_KEY="your-secret-key"
   go run cmd/api/main.go
   ```

3. **Frontend**
   ```bash
   cd frontend
   npm run dev
   ```

## Architecture

The backend follows a hexagonal architecture:
- `internal/modules/auth`: Authentication module
  - `domain`: Core business logic and entities
  - `application`: Use cases
  - `adapters`: HTTP handlers and Database repositories
