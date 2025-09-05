#!/bin/bash

echo "Setting up Database Manager Application..."

echo ""
echo "Installing backend dependencies..."
cd backend
go mod init db-manager-backend 2>/dev/null
go mod tidy
if [ $? -ne 0 ]; then
    echo "Failed to install backend dependencies"
    exit 1
fi

echo ""
echo "Installing frontend dependencies..."
cd ../frontend
npm install
if [ $? -ne 0 ]; then
    echo "Failed to install frontend dependencies"
    exit 1
fi

cd ..

echo ""
echo "Setup completed successfully!"
echo ""
echo "To start the application:"
echo "  npm run dev"
echo ""
echo "The application will be available at:"
echo "  Frontend: http://localhost:3000"
echo "  Backend API: http://localhost:8080"
echo ""
