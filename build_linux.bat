@echo off
rem -------------------------------------------------------------
rem  build_linux.bat  — Cross-compile ToughRADIUS for Linux/amd64
rem -------------------------------------------------------------
rem  使用方法：双击或在 CMD 中执行即可在 Windows 环境下生成 Linux 可执行文件。
rem  生成的二进制位于  ./dist/toughradius-linux-amd64  （无扩展名）。
rem -------------------------------------------------------------

setlocal

rem 确保 Go 在 PATH 中
where go >nul 2>nul || (
    echo [ERROR] Go toolchain not found in PATH.
    goto :eof
)

rem 创建输出目录
if not exist release mkdir release

rem Cross-compile target
set GOOS=linux
set GOARCH=amd64

echo Building ToughRADIUS for %GOOS%/%GOARCH% (default flags) ...

rem 如果需要静态 or strip，可自行添加 -ldflags 等参数
go build -o release/toughradius .
if %errorlevel% neq 0 (
    echo [ERROR] Build failed.
    exit /b %errorlevel%
)

echo.
echo Done. Output: release\toughradius

endlocal 