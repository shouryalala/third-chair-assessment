#!/bin/bash

set -e  # Exit on any error

echo "🚀 Instagram User Processor - Setup Script"
echo "=========================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.23+ first."
    echo "   Visit: https://golang.org/doc/install"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go version: $GO_VERSION"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    echo "   Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

echo "✅ Docker is available"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "⚠️  Please edit .env file and add your ROCKETAPI_KEY"
else
    echo "✅ .env file already exists"
fi

# Download Go dependencies
echo "📦 Installing Go dependencies..."
go mod tidy

# Start database
echo "🐘 Starting PostgreSQL database..."
docker-compose up -d postgres

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U user -d instagram_processor > /dev/null 2>&1; then
        echo "✅ Database is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ Database failed to start after 30 seconds"
        exit 1
    fi
    sleep 1
done

# Build the application
echo "🔨 Building the application..."
go build -o bin/server cmd/server/main.go

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "Next steps:"
echo "1. Edit the .env file and add your RocketAPI key"
echo "2. Start the server: ./bin/server"
echo "3. Or run in development mode: go run cmd/server/main.go"
echo ""
echo "🌐 API will be available at: http://localhost:8080"
echo "🔧 Database admin interface: http://localhost:8081 (adminer)"
echo "📚 Health check: http://localhost:8080/health"
echo ""
echo "💡 Test the single user endpoint:"
echo "   curl http://localhost:8080/api/v1/instagram/user/musiclover2024"
echo ""
echo "🚀 Ready to implement batch processing!"