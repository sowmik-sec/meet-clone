#!/bin/bash

# Meet Clone - Installation Verification Script

echo "🔍 Meet Clone - Installation Verification"
echo "=========================================="
echo ""

# Function to check command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo "✅ $2"
    else
        echo "❌ $2"
    fi
}

# Check Go
echo "Checking Prerequisites..."
echo "-------------------------"

if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}')
    print_status 0 "Go installed: $GO_VERSION"
else
    print_status 1 "Go not found"
fi

# Check Node
if command_exists node; then
    NODE_VERSION=$(node --version)
    print_status 0 "Node.js installed: $NODE_VERSION"
else
    print_status 1 "Node.js not found"
fi

# Check npm
if command_exists npm; then
    NPM_VERSION=$(npm --version)
    print_status 0 "npm installed: $NPM_VERSION"
else
    print_status 1 "npm not found"
fi

# Check Docker
if command_exists docker; then
    if docker info > /dev/null 2>&1; then
        print_status 0 "Docker installed and running"
    else
        print_status 1 "Docker installed but not running"
    fi
else
    print_status 1 "Docker not found"
fi

echo ""
echo "Checking Project Files..."
echo "-------------------------"

# Check backend files
if [ -f "backend/go.mod" ]; then
    print_status 0 "Backend Go module found"
else
    print_status 1 "Backend Go module not found"
fi

if [ -f "backend/cmd/api/main.go" ]; then
    print_status 0 "Backend main.go found"
else
    print_status 1 "Backend main.go not found"
fi

if [ -f "backend/.env.example" ]; then
    print_status 0 "Backend .env.example found"
else
    print_status 1 "Backend .env.example not found"
fi

# Check frontend files
if [ -f "frontend/package.json" ]; then
    print_status 0 "Frontend package.json found"
else
    print_status 1 "Frontend package.json not found"
fi

if [ -d "frontend/src" ]; then
    print_status 0 "Frontend src directory found"
else
    print_status 1 "Frontend src directory not found"
fi

if [ -f "frontend/.env.example" ]; then
    print_status 0 "Frontend .env.example found"
else
    print_status 1 "Frontend .env.example not found"
fi

# Check Docker Compose
if [ -f "docker-compose.yml" ]; then
    print_status 0 "docker-compose.yml found"
else
    print_status 1 "docker-compose.yml not found"
fi

echo ""
echo "Checking Dependencies..."
echo "------------------------"

# Check backend dependencies
if [ -f "backend/go.sum" ]; then
    print_status 0 "Backend dependencies downloaded"
else
    echo "⚠️  Backend dependencies not installed yet"
    echo "   Run: cd backend && go mod download"
fi

# Check frontend dependencies
if [ -d "frontend/node_modules" ]; then
    print_status 0 "Frontend dependencies installed"
else
    echo "⚠️  Frontend dependencies not installed yet"
    echo "   Run: cd frontend && npm install"
fi

echo ""
echo "Checking Configuration..."
echo "-------------------------"

# Check backend .env
if [ -f "backend/.env" ]; then
    print_status 0 "Backend .env file exists"
else
    echo "⚠️  Backend .env not found"
    echo "   Run: cp backend/.env.example backend/.env"
fi

# Check frontend .env.local
if [ -f "frontend/.env.local" ]; then
    print_status 0 "Frontend .env.local file exists"
else
    echo "⚠️  Frontend .env.local not found"
    echo "   Run: cp frontend/.env.example frontend/.env.local"
fi

echo ""
echo "=========================================="
echo ""

# Summary
if command_exists go && command_exists node && command_exists docker; then
    echo "✅ All prerequisites are installed!"
    echo ""
    echo "Next steps:"
    echo "1. Run ./start.sh to setup and start MongoDB"
    echo "2. In terminal 1: cd backend && go run cmd/api/main.go"
    echo "3. In terminal 2: cd frontend && npm run dev"
    echo "4. Open http://localhost:3000"
else
    echo "❌ Some prerequisites are missing"
    echo ""
    echo "Please install:"
    if ! command_exists go; then
        echo "  - Go 1.21+: https://go.dev/dl/"
    fi
    if ! command_exists node; then
        echo "  - Node.js 18+: https://nodejs.org/"
    fi
    if ! command_exists docker; then
        echo "  - Docker: https://www.docker.com/get-started"
    fi
fi

echo ""
