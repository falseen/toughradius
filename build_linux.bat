@echo off
rem -------------------------------------------------------------
rem  build_linux.bat  �� Cross-compile ToughRADIUS for Linux/amd64
rem -------------------------------------------------------------
rem  ʹ�÷�����˫������ CMD ��ִ�м����� Windows ���������� Linux ��ִ���ļ���
rem  ���ɵĶ�����λ��  ./dist/toughradius-linux-amd64  ������չ������
rem -------------------------------------------------------------

setlocal

rem ȷ�� Go �� PATH ��
where go >nul 2>nul || (
    echo [ERROR] Go toolchain not found in PATH.
    goto :eof
)

rem �������Ŀ¼
if not exist release mkdir release

rem Cross-compile target
set GOOS=linux
set GOARCH=amd64

echo Building ToughRADIUS for %GOOS%/%GOARCH% (default flags) ...

rem �����Ҫ��̬ or strip����������� -ldflags �Ȳ���
go build -o release/toughradius .
if %errorlevel% neq 0 (
    echo [ERROR] Build failed.
    exit /b %errorlevel%
)

echo.
echo Done. Output: release\toughradius

endlocal 