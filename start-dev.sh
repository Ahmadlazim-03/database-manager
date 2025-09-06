#!/bin/bash

echo "========================================"
echo "  Database Manager - Development Mode"
echo "========================================"
echo ""

echo "Starting both backend and frontend..."
echo ""
echo "Backend will run on: http://localhost:8080"
echo "Frontend will run on: http://localhost:3000"
echo ""

# Start backend in background
cd backend
echo "🚀 Starting backend..."
./db-manager-backend &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 3

# Start frontend in background  
cd ../frontend
echo "🚀 Starting frontend..."
npm run dev &
FRONTEND_PID=$!

echo ""
echo "✅ Both services started!"
echo ""
echo "📖 Usage:"
echo "   - Open browser to http://localhost:3000"
echo "   - Backend API available at http://localhost:8080/api"
echo ""
echo "Press Ctrl+C to stop both services..."

# Function to cleanup processes on exit
cleanup() {
    echo ""
    echo "🛑 Stopping services..."
    kill $BACKEND_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    echo "✅ Services stopped successfully!"
    exit 0
}

# Set trap to cleanup on script exit
trap cleanup SIGINT SIGTERM

# Wait for processes to finish
wait
