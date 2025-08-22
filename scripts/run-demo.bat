@echo off
REM Windows batch file to run the complete demo scenario
REM Requirements: 7.4, 7.5

echo Starting Monitoring System Demo...
echo.

REM Check if bash is available (Git Bash, WSL, etc.)
where bash >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Error: bash is not available in PATH
    echo Please install Git Bash, WSL, or another bash environment
    echo Alternatively, run the scripts manually in a bash shell
    pause
    exit /b 1
)

REM Run the demo scenario
bash scripts/demo-scenario.sh

pause