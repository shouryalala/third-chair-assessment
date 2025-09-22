#!/bin/bash

set -e  # Exit on any error

echo "🚀 Instagram User Processor - Setup Script"
echo "=========================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.20+ first."
    echo "   Visit: https://golang.org/doc/install"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go version: $GO_VERSION"

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo "❌ PostgreSQL is not installed. Please install PostgreSQL first."
    echo "   - macOS: brew install postgresql"
    echo "   - Ubuntu/Debian: sudo apt-get install postgresql"
    echo "   - Windows: Download from https://www.postgresql.org/download/"
    exit 1
fi

echo "✅ PostgreSQL is available"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "⚠️  Please edit .env file and add your ROCKETAPI_KEY"
    echo "⚠️  Update DATABASE_URL with your PostgreSQL credentials"
else
    echo "✅ .env file already exists"
fi

# Download Go dependencies
echo "📦 Installing Go dependencies..."
go mod download
go mod tidy

# Create database (prompt for credentials)
echo ""
echo "🐘 Database Setup"
echo "================"
echo "Please enter your PostgreSQL credentials:"
read -p "PostgreSQL user (default: postgres): " DB_USER
DB_USER=${DB_USER:-postgres}

read -p "PostgreSQL host (default: localhost): " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "PostgreSQL port (default: 5432): " DB_PORT
DB_PORT=${DB_PORT:-5432}

echo ""
echo "Creating database 'instagram_processor'..."
echo "(You may be prompted for your PostgreSQL password)"

# Create database if it doesn't exist
createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" instagram_processor 2>/dev/null || {
    echo "Database might already exist or creation failed."
    echo "If the database already exists, you can continue."
    read -p "Continue anyway? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
}

# Run database initialization
echo "📊 Initializing database schema..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d instagram_processor -f init.sql || {
    echo "⚠️  Database initialization might have failed."
    echo "This could be normal if tables already exist."
}

# Run test data (optional)
read -p "Load sample test data? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "📝 Loading test data..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d instagram_processor -f scripts/test_data.sql
fi

# Build the application
echo "🔨 Building the application..."
go build -o bin/server cmd/server/main.go

# Update .env with database connection
echo ""
echo "📝 Update your .env file with this DATABASE_URL:"
echo "DATABASE_URL=postgres://$DB_USER:YOUR_PASSWORD@$DB_HOST:$DB_PORT/instagram_processor?sslmode=disable"

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "Next steps:"
echo "1. Edit the .env file:"
echo "   - Add your RocketAPI key (get from https://rocketapi.io)"
echo "   - Update DATABASE_URL with your PostgreSQL password"
echo "2. Start the server: ./bin/server"
echo "3. Or run in development mode: go run cmd/server/main.go"
echo ""
echo "🌐 API will be available at: http://localhost:8080"
echo "📚 Health check: http://localhost:8080/health"
echo ""
echo "💡 Test the single user endpoint:"
echo "   curl http://localhost:8080/api/v1/instagram/user/musiclover2024"
echo ""
echo "🚀 Ready to implement batch processing!"