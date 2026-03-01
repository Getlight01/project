@echo off
echo Stopping Fashion Look Generator servers...

echo [1/2] Stopping Frontend (Node.js)...
taskkill /F /IM node.exe 2>nul
if %errorlevel% == 0 (
    echo Frontend stopped successfully.
) else (
    echo Frontend was not running.
)

echo [2/2] Stopping Backend (Go server)...
taskkill /F /IM server.exe 2>nul
taskkill /F /IM go.exe 2>nul
if %errorlevel% == 0 (
    echo Backend stopped successfully.
) else (
    echo Backend was not running.
)

echo.
echo ==========================================
echo All servers stopped!
echo ==========================================
pause
