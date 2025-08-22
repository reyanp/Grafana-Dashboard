@echo off
echo Starting server...
start /B api.exe

echo Waiting for server to start...
timeout /t 2 /nobreak > nul

echo Making work request in background...
start /B curl "http://localhost:8080/api/v1/work?ms=3000"

echo Waiting a moment for request to start...
timeout /t 1 /nobreak > nul

echo Sending SIGTERM to server (Ctrl+C)...
echo Server should wait for work to complete before shutting down
echo Check the logs to verify graceful shutdown behavior

taskkill /F /IM api.exe > nul 2>&1