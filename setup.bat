@echo off
echo Setting up Database Manager Application...

echo.
echo Installing backend dependencies...
cd backend
go mod init db-manager-backend 2>nul
go mod tidy
if %errorlevel% neq 0 (
    echo Failed to install backend dependencies
    pause
    exit /b 1
)

echo.
echo Installing frontend dependencies...
cd ..\frontend
call npm install
if %errorlevel% neq 0 (
    echo Failed to install frontend dependencies
    pause
    exit /b 1
)

cd ..

echo.
echo Setup completed successfully!
echo.
echo To start the application:
echo   npm run dev
echo.
echo The application will be available at:
echo   Frontend: http://localhost:3000
echo   Backend API: http://localhost:8080
echo.
pause
