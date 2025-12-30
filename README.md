# Meet Clone

A real-time video conferencing application built with Go, Next.js, Cloudflare Calls API, and MongoDB.

## 🎯 Features

- 🎥 Real-time video and audio conferencing (up to 10 participants)
- 💬 Live chat during meetings
- 🔒 JWT-based authentication
- 👥 Participant management
- ⚡ WebRTC powered by Cloudflare Calls
- 🏗️ Hexagonal architecture backend

## 🛠️ Tech Stack

### Backend
- **Go 1.21+** with hexagonal architecture
- **MongoDB** for data persistence
- **WebSocket** for real-time events
- **JWT** authentication
- **Gorilla Mux** for routing

### Frontend
- **Next.js 14+** with App Router
- **TypeScript** for type safety
- **Tailwind CSS** for styling
- **Zustand** for state management
- **WebRTC** for media streaming

## 📁 Project Structure

```
meet-clone/
├── backend/           # Go backend
│   ├── cmd/
│   │   └── api/       # Entry point
│   ├── internal/
│   │   ├── core/      # Business logic
│   │   ├── adapters/  # Infrastructure
│   │   ├── config/    # Configuration
│   │   └── pkg/       # Utilities
│   └── go.mod
├── frontend/          # Next.js frontend
│   ├── src/
│   │   ├── app/       # Pages
│   │   ├── components/# UI components
│   │   ├── lib/       # Utilities
│   │   └── hooks/     # Custom hooks
│   └── package.json
├── docker-compose.yml
└── README.md
```

## 🚀 Quick Start

### Prerequisites

- **Go** 1.21 or higher
- **Node.js** 18 or higher
- **Docker** and Docker Compose
- **MongoDB** (or use Docker)

### 1. Clone the repository

```bash
git clone <your-repo-url>
cd meet-clone
```

### 2. Start MongoDB

```bash
docker-compose up -d
```

### 3. Setup Backend

```bash
cd backend
cp .env.example .env
# Edit .env with your configuration
go mod download
go run cmd/api/main.go
```

Backend will run on http://localhost:8080

### 4. Setup Frontend

```bash
cd frontend
npm install
cp .env.example .env.local
# Edit .env.local with your configuration
npm run dev
```

Frontend will run on http://localhost:3000

## 🔧 Environment Variables

### Backend (.env)
```env
MONGODB_URI=mongodb://localhost:27017/meet-clone
JWT_SECRET=your-super-secret-key-change-this
JWT_EXPIRY=24h
PORT=8080
ENV=development
CORS_ORIGIN=http://localhost:3000
```

### Frontend (.env.local)
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

## 📚 API Documentation

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `GET /api/v1/auth/me` - Get current user (requires auth)

### Rooms
- `POST /api/v1/rooms` - Create new room (requires auth)
- `GET /api/v1/rooms/:id` - Get room details
- `POST /api/v1/rooms/:id/join` - Join room (requires auth)
- `POST /api/v1/rooms/:id/leave` - Leave room (requires auth)
- `GET /api/v1/rooms/:id/participants` - Get participants

### Chat
- `GET /api/v1/rooms/:id/messages` - Get messages (pagination)
- `WS /api/v1/ws/room/:id` - WebSocket connection for real-time events

### WebSocket Events
- `participant_joined` - New participant joined
- `participant_left` - Participant left
- `chat_message` - New chat message
- `room_ended` - Room ended

## 🏗️ Architecture

The backend follows **Hexagonal Architecture** (Ports & Adapters):

- **Core Layer**: Business logic and domain entities
- **Ports**: Interfaces for services and repositories
- **Adapters**: HTTP handlers, WebSocket, MongoDB implementations

## 🧪 Testing

### Backend
```bash
cd backend
go test ./...
```

### Frontend
```bash
cd frontend
npm test
```

## 🐳 Docker Commands

```bash
# Start MongoDB
docker-compose up -d

# Stop MongoDB
docker-compose down

# View logs
docker-compose logs -f
```

## 📦 Building for Production

### Backend
```bash
cd backend
go build -o bin/api cmd/api/main.go
./bin/api
```

### Frontend
```bash
cd frontend
npm run build
npm start
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

MIT License - see LICENSE file for details

## 🙋 Support

For questions or issues, please open a GitHub issue.
