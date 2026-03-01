@echo off
echo Starting Fashion Look Generator servers...

echo [1/2] Starting Backend (Go) on port 8085...
start "Backend Server" cmd /k "go run cmd/server/main.go"

echo Waiting 3 seconds for backend to initialize...
timeout /t 3 /nobreak >nul

echo [2/2] Starting Frontend (React) on port 3005...
cd frontend
start "Frontend Server" cmd /k "npm run dev"

echo.
echo ==========================================
echo Servers starting...
echo.
echo Backend:  http://localhost:8085
echo Frontend: http://localhost:3005
echo.
echo Browser will open automatically...
echo ==========================================
timeout /t 2 /nobreak >nul
start http://localhost:3005/
