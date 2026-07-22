@echo off
rem Start the extractor service if it is not listening on 8080, then open the page
netstat -ano | findstr /C:":8080 " | findstr /C:"LISTENING" >nul
if errorlevel 1 (
    start "" /min "%~dp0N0vaDesktop-Extractor-arm64.exe"
    ping 127.0.0.1 -n 3 >nul
)
start "" http://127.0.0.1:8080
