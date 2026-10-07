@echo off
setlocal enabledelayedexpansion

echo ==========================================
echo   星火数据统计系统 - Windows 构建脚本
echo ==========================================
echo.

:: 设置Go编译参数
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
set GO111MODULE=on

:: 获取版本信息
for /f "tokens=*" %%i in ('git describe --tags --dirty --always 2^>nul') do set GIT_VERSION=%%i
if "%GIT_VERSION%"=="" set GIT_VERSION=unknown

for /f "tokens=*" %%i in ('git rev-parse HEAD 2^>nul') do set GIT_COMMIT=%%i
if "%GIT_COMMIT%"=="" set GIT_COMMIT=unknown

for /f "tokens=*" %%i in ('powershell -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ss'"') do set DATE=%%i

echo Version:  %GIT_VERSION%
echo Commit:   %GIT_COMMIT%
echo BuildDate:%DATE%
echo.

:: 创建输出目录
if not exist ".\output" mkdir ".\output"

:: 编译 - 关闭CGO实现完全静态编译，可跨机器运行
set CGO_ENABLED=0
go build -trimpath -ldflags "-s -w -X main.Version=%GIT_VERSION% -X main.BuildDate=%DATE%" -o ./output/xing-huo.exe ./main.go

if errorlevel 1 (
    echo Error: Go build failed!
    exit /b 1
)

:: 复制excel目录（如果存在）
if exist ".\excel" (
    echo 复制 excel 目录到 output...
    xcopy /s /i /y ".\excel" ".\output\excel"
)

echo.
echo ==========================================
echo   构建成功!
echo   输出: .\output\xing-huo.exe
echo ==========================================
echo.
echo 使用方法:
echo   1. 将 output 文件夹整个复制到目标电脑
echo   2. 在 output\excel\ 目录下放入 .xlsx 文件
echo   3. 双击运行 xing-huo.exe
echo   4. 浏览器访问 http://localhost/data/index
echo ==========================================

dir .\output