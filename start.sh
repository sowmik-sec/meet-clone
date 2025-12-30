#!/bin/bash

# Meet Clone - Quick Start Script
# This script will start all services for local development

echo "🚀 Starting Meet Clone Application..."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Start MongoDB
echo "📦 Starting MongoDB..."
docker-compose up -d

# Wait for MongoDB to be ready
echo "⏳ Waiting for MongoDB to be ready..."
sleep 3

# Check if backend dependencies are installed
if [ ! -f "backend/go.sum" ]; then
    echo "📥 Installing backend dependencies..."
    cd backend && go mod download && cd ..
fi

# Check if frontend dependencies are installed
if [ ! -d "frontend/node_modules" ]; then
    echo "📥 Installing frontend dependencies..."
    cd frontend && npm install && cd ..
fi

# Check if .env files exist
if [ ! -f "backend/.env" ]; then
    echo "⚙️  Creating backend .env file..."
    cp backend/.env.example backend/.env
fi

if [ ! -f "frontend/.env.local" ]; then
    echo "⚙️  Creating frontend .env.local file..."
    cp frontend/.env.example frontend/.env.local
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "📝 Next steps:"
echo "   1. Open a new terminal and run: cd backend && go run cmd/api/main.go"
echo "   2. Open another terminal and run: cd frontend && npm run dev"
echo "   3. Open http://localhost:3000 in your browser"
echo ""
echo "💡 Tips:"
echo "   - Backend will run on http://localhost:8080"
echo "   - Frontend will run on http://localhost:3000"
echo "   - MongoDB will run on mongodb://localhost:27017"
echo ""
echo "🛑 To stop MongoDB: docker-compose down"
echo ""
