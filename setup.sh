#!/bin/bash

# Database Manager Setup Script with Password Protection
# This script protects the project from unauthorized access

echo "=========================================="
echo "  Database Manager Project Setup"
echo "=========================================="
echo ""

# Password protection
CORRECT_PASSWORD="admin123secure"  # Change this to your desired password
MAX_ATTEMPTS=3
attempts=0

while [ $attempts -lt $MAX_ATTEMPTS ]; do
    echo -n "Enter setup password: "
    read -s password
    echo ""
    
    if [ "$password" = "$CORRECT_PASSWORD" ]; then
        echo "✅ Password correct! Proceeding with setup..."
        echo ""
        break
    else
        attempts=$((attempts + 1))
        remaining=$((MAX_ATTEMPTS - attempts))
        
        if [ $remaining -gt 0 ]; then
            echo "❌ Incorrect password. $remaining attempts remaining."
            echo ""
        else
            echo "❌ Maximum attempts exceeded. Access denied!"
            echo "🔒 Project setup terminated for security reasons."
            exit 1
        fi
    fi
done

echo "🚀 Starting Database Manager setup..."
echo ""

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js first."
    echo "   Download from: https://nodejs.org/"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    echo "   Download from: https://golang.org/dl/"
    exit 1
fi

echo "✅ Node.js version: $(node --version)"
echo "✅ Go version: $(go version)"
echo ""

# Setup backend
echo "📦 Setting up backend dependencies..."
cd backend
if [ ! -f ".env" ]; then
    echo "📝 Creating backend .env file..."
    cat > .env << EOL
PORT=8080
JWT_SECRET=your_jwt_secret_here_change_this
DB_HOST=localhost
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=

EOL
    echo "✅ Backend .env file created"
else
    echo "✅ Backend .env file already exists"
fi

echo "📥 Installing Go dependencies..."
go mod download
if [ $? -eq 0 ]; then
    echo "✅ Backend dependencies installed successfully"
else
    echo "❌ Failed to install backend dependencies"
    exit 1
fi

# Build backend
echo "🔨 Building backend..."
go build -o db-manager-backend main.go
if [ $? -eq 0 ]; then
    echo "✅ Backend built successfully"
else
    echo "❌ Failed to build backend"
    exit 1
fi

cd ..

# Setup frontend
echo ""
echo "📦 Setting up frontend dependencies..."
cd frontend

if [ ! -f ".env" ]; then
    echo "📝 Creating frontend .env file..."
    cat > .env << EOL
# Backend Configuration
VITE_BACKEND_PORT=8080
VITE_API_BASE_URL=http://localhost:8080/api

# Frontend Configuration (automatically detected by Vite)
VITE_FRONTEND_PORT=3000
EOL
    echo "✅ Frontend .env file created"
else
    echo "✅ Frontend .env file already exists"
fi

echo "📥 Installing Node.js dependencies..."
npm install
if [ $? -eq 0 ]; then
    echo "✅ Frontend dependencies installed successfully"
else
    echo "❌ Failed to install frontend dependencies"
    exit 1
fi

cd ..

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "📋 Next steps:"
echo "   1. Configure your database settings in backend/.env"
echo "   2. Update JWT_SECRET in backend/.env"
echo "   3. Run the application using start-dev.sh or individual scripts"
echo ""
echo "🌐 Access the application at: http://localhost:3000"
echo ""