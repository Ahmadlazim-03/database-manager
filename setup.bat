@echo off
setlocal EnableDelayedExpansion

REM Database Manager Setup Script with Password Protection
REM This script protects the project from unauthorized access

echo ==========================================
echo   Database Manager Project Setup
echo ==========================================
echo.

REM Password protection
set "CORRECT_PASSWORD=admin123secure"
set "MAX_ATTEMPTS=3"
set "attempts=0"

:password_loop
if !attempts! geq !MAX_ATTEMPTS! (
    echo ❌ Maximum attempts exceeded. Access denied!
    echo 🔒 Project setup terminated for security reasons.
    pause
    exit /b 1
)

set /p "password=Enter setup password: "
if "!password!"=="!CORRECT_PASSWORD!" (
    echo ✅ Password correct! Proceeding with setup...
    echo.
    goto setup_start
) else (
    set /a attempts+=1
    set /a remaining=!MAX_ATTEMPTS!-!attempts!
    if !remaining! gtr 0 (
        echo ❌ Incorrect password. !remaining! attempts remaining.
        echo.
        goto password_loop
    ) else (
        echo ❌ Maximum attempts exceeded. Access denied!
        echo 🔒 Project setup terminated for security reasons.
        pause
        exit /b 1
    )
)

:setup_start
echo 🚀 Starting Database Manager setup...
echo.

REM Check if Node.js is installed
node --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Node.js is not installed. Please install Node.js first.
    echo    Download from: https://nodejs.org/
    pause
    exit /b 1
)

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ Go is not installed. Please install Go first.
    echo    Download from: https://golang.org/dl/
    pause
    exit /b 1
)

for /f "tokens=*" %%i in ('node --version') do set NODE_VERSION=%%i
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i

echo ✅ Node.js version: !NODE_VERSION!
echo ✅ Go version: !GO_VERSION!
echo.

REM Setup backend
echo 📦 Setting up backend dependencies...
cd backend

if not exist ".env" (
    echo 📝 Creating backend .env file...
    (
        echo PORT=8080
        echo JWT_SECRET=your_jwt_secret_here_change_this
        echo DB_HOST=localhost
        echo DB_PORT=5432
        echo DB_NAME=postgres
        echo DB_USER=postgres
        echo DB_PASSWORD=
        echo.
    ) > .env
    echo ✅ Backend .env file created
) else (
    echo ✅ Backend .env file already exists
)

go mod init db-manager-backend 2>nul
echo 📥 Installing Go dependencies...
go mod tidy
if errorlevel 1 (
    echo ❌ Failed to install backend dependencies
    pause
    exit /b 1
)
echo ✅ Backend dependencies installed successfully

REM Build backend
echo 🔨 Building backend...
go build -o db-manager-backend.exe main.go
if errorlevel 1 (
    echo ❌ Failed to build backend
    pause
    exit /b 1
)
echo ✅ Backend built successfully

cd ..

REM Setup frontend
echo.
echo 📦 Setting up frontend dependencies...
cd frontend

if not exist ".env" (
    echo 📝 Creating frontend .env file...
    (
        echo # Backend Configuration
        echo VITE_BACKEND_PORT=8080
        echo VITE_API_BASE_URL=http://localhost:8080/api
        echo.
        echo # Frontend Configuration ^(automatically detected by Vite^)
        echo VITE_FRONTEND_PORT=3000
    ) > .env
    echo ✅ Frontend .env file created
) else (
    echo ✅ Frontend .env file already exists
)

echo 📥 Installing Node.js dependencies...
npm install
if errorlevel 1 (
    echo ❌ Failed to install frontend dependencies
    pause
    exit /b 1
)
echo ✅ Frontend dependencies installed successfully

cd ..

REM Create startup scripts
echo.
echo 📜 Creating startup scripts...

REM Create development start script
(
    echo @echo off
    echo echo ========================================
    echo echo   Database Manager - Development Mode
    echo echo ========================================
    echo echo.
    echo echo Starting both backend and frontend...
    echo echo.
    echo echo Backend will run on: http://localhost:8080
    echo echo Frontend will run on: http://localhost:3000
    echo echo.
    echo start "Backend" cmd /k "cd backend && db-manager-backend.exe"
    echo timeout /t 3 /nobreak ^>nul
    echo start "Frontend" cmd /k "cd frontend && npm run dev"
    echo echo.
    echo echo ✅ Both services started!
    echo echo.
    echo echo 📖 Usage:
    echo echo    - Open browser to http://localhost:3000
    echo echo    - Backend API available at http://localhost:8080/api
    echo echo.
    echo pause
) > start-dev.bat

echo ✅ Development startup script created

echo.
echo 🎉 Setup completed successfully!
echo.
echo 📋 Next steps:
echo    1. Configure your database settings in backend\.env
echo    2. Update JWT_SECRET in backend\.env
echo    3. Run the application:
echo.
echo    Windows:
echo      - start-dev.bat ^(starts both services^)
echo.
echo 🌐 Access the application at: http://localhost:3000
echo.
pause
