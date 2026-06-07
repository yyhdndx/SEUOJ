@echo off
REM 双击或在 cmd 中运行：start-dev.bat
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0dev.ps1" %*
pause
