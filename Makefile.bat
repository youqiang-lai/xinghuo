@echo off
setlocal enabledelayedexpansion

echo ==============================================
echo   星火数据统计系统 - Windows 构建脚本
echo ==============================================
echo.

:: 设置Go编译参数
set CGO_ENABLED=0
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

set LDFLAGS=-s -w -X main.Version=%GIT_VERSION% -X main.BuildDate=%DATE%

:: ========== Windows ==========
echo [1/3] 编译 Windows amd64...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -trimpath -ldflags "%LDFLAGS%" -o ./output/xing-huo.exe ./main.go
if errorlevel 1 (
    echo Error: Windows build failed!
    exit /b 1
)
echo   OK: xing-huo.exe

:: ========== macOS Intel ==========
echo [2/3] 编译 macOS Intel...
set GOOS=darwin
set GOARCH=amd64
set CGO_ENABLED=0
go build -trimpath -ldflags "%LDFLAGS%" -o ./output/xing-huo-darwin-amd64 ./main.go
if errorlevel 1 (
    echo Error: macOS Intel build failed!
    exit /b 1
)
echo   OK: xing-huo-darwin-amd64
:: 打包成 tar.gz 以保留文件权限
tar --mode=755 -czf ./output/xing-huo-darwin-amd64.tar.gz -C ./output xing-huo-darwin-amd64
if errorlevel 1 (
    echo Warning: tar 打包 macOS Intel 失败（可能未安装 tar），跳过
)
del /q .\output\xing-huo-darwin-amd64

:: ========== macOS Apple Silicon ==========
echo [3/3] 编译 macOS Apple Silicon...
set GOOS=darwin
set GOARCH=arm64
set CGO_ENABLED=0
go build -trimpath -ldflags "%LDFLAGS%" -o ./output/xing-huo-darwin-arm64 ./main.go
if errorlevel 1 (
    echo Error: macOS ARM build failed!
    exit /b 1
)
echo   OK: xing-huo-darwin-arm64
tar --mode=755 -czf ./output/xing-huo-darwin-arm64.tar.gz -C ./output xing-huo-darwin-arm64
if errorlevel 1 (
    echo Warning: tar 打包 macOS ARM 失败（可能未安装 tar），跳过
)
del /q .\output\xing-huo-darwin-arm64

:: 复制excel目录（如果存在）
if exist ".\excel" (
    echo.
    echo 复制 excel 目录到 output...
    xcopy /s /i /y ".\excel" ".\output\excel"
)

echo.
echo ==============================================
echo   构建成功!
echo ==============================================
echo.
echo Windows 用户:
echo   使用 output\xing-huo.exe
echo.
echo macOS 用户:
echo   将 output\xing-huo-darwin-amd64.tar.gz 或
echo   output\xing-huo-darwin-arm64.tar.gz 拷贝到 Mac
echo   解压后终端执行:
echo     tar -xzf xing-huo-darwin-arm64.tar.gz
echo     chmod +x xing-huo-darwin-arm64
echo     ./xing-huo-darwin-arm64
echo.
echo 使用方法:
echo   1. 将 output 文件夹整个复制到目标电脑
echo   2. 在 output\excel\ 目录下放入 .xlsx 文件
echo   3. 运行对应平台的可执行文件
echo   4. 浏览器访问 http://localhost/data/index
echo ==============================================

dir .\output