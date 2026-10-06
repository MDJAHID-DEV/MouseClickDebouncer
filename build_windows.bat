@echo off
setlocal

echo [1/3] Checking Go...
go version || exit /b 1

echo [2/3] Embedding Windows manifest + version information...
where go-winres >nul 2>nul
if errorlevel 1 (
  echo go-winres is not installed.
  echo Install it once with:
  echo   go install github.com/tc-hib/go-winres@v0.3.3
  exit /b 1
)
go-winres make || exit /b 1

echo [3/3] Building MouseClickDebouncer.exe...
go build -ldflags="-H=windowsgui" -o MouseClickDebouncer.exe . || exit /b 1

echo.
echo Build complete: MouseClickDebouncer.exe
endlocal
