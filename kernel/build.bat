@echo off
echo 'Building Kernel'
@REM the C compiler "gcc" is necessary https://sourceforge.net/projects/mingw-w64/files/mingw-w64/
go version
set GO111MODULE=on
set GOPROXY=https://mirrors.aliyun.com/goproxy/
set CGO_ENABLED=1
set PATH=C:\_devel\QMK_MSYS\mingw64\bin;P:\__go\go1-23-12\go1.23.12\bin;C:\_devel\cuda-12.1\bin;C:\_devel\cuda-12.1\libnvvp;c:\_devel\python\311_64\;c:\_devel\python\311_64\Scripts\;C:\Program Files\ImageMagick-7.1.0-Q16-HDRI;C:\_devel\cuda-11.8\bin;C:\_devel\cuda-11.8\libnvvp;C:\Windows\system32;C:\Windows;C:\Windows\System32\Wbem;C:\Windows\System32\WindowsPowerShell\v1.0\;C:\Windows\System32\OpenSSH\;C:\Program Files\dotnet\;C:\_devel\nodejs-18\;C:\_devel\CMake\bin;c:\path\;C:\Program Files\Common Files\Breuckmann\Optocat\Lib;C:\_devel\vs2022ent\VC\Auxiliary\Build\;c:\Users\user\.nimble\bin\;C:\Program Files\Microsoft VS Code\bin;c:\_devel\sapling;c:\_devel\github-cli;C:\Program Files\TortoiseHg\;p:\__py\Python310_64\Scripts;c:\_devel\sapling\watchman-v2023.10.09.00-windows\bin\;c:\_devel\git\cmd;C:\_devel\git\bin\;C:\_devel\git\usr\bin\;C:\work\Python27\lib\site-packages\gtk-2.0\runtime\bin;C:\Program Files\NVIDIA\CUDNN\v9.0\bin;C:\Program Files\Microsoft MPI\Bin\;C:\_devel\cuda-12.1\bin;C:\_devel\cuda-12.1\libnvvp;c:\_devel\python\311_64\;c:\_devel\python\311_64\Scripts\;C:\Program Files\ImageMagick-7.1.0-Q16-HDRI;C:\_devel\cuda-11.8\bin;C:\_devel\cuda-11.8\libnvvp;C:\Windows\system32;C:\Windows;C:\Windows\System32\Wbem;C:\Windows\System32\WindowsPowerShell\v1.0\;C:\Windows\System32\OpenSSH\;C:\Program Files\Microsoft SQL Server\150\Tools\Binn\;C:\Program Files\dotnet\;C:\_devel\nodejs-18\;C:\_devel\CMake\bin;C:\Program Files\Microsoft SQL Server\130\Tools\Binn\;c:\path\;C:\Program Files\Common Files\Breuckmann\Optocat\Lib;C:\_devel\vs2022ent\VC\Auxiliary\Build\;c:\Users\user\.nimble\bin\;C:\Program Files\Microsoft VS Code\bin;c:\_devel\sapling;c:\_devel\github-cli;C:\Program Files\TortoiseHg\;p:\__py\Python310_64\Scripts;c:\_devel\sapling\watchman-v2023.10.09.00-windows\bin\;c:\_devel\git\cmd;C:\_devel\git\bin\;C:\_devel\git\usr\bin\;

@REM you can use `go mod tidy` to update kernel dependency before build
@REM you can use `go generate` instead (need add something in main.go)
goversioninfo -platform-specific=true -icon=resource/icon.ico -manifest=resource/goversioninfo.exe.manifest

echo 'Building Kernel amd64'
set GOOS=windows
set GOARCH=amd64
go build --tags fts5 -v -o "../app/kernel/SiYuan-Kernel.exe" -ldflags "-s -w" .
@REM go build --tags fts5 -v -o "../app/kernel/SiYuan-Kernel.exe" -ldflags "-s -w -H=windowsgui" .
@REM sqlite fulltext search не нужен, поиск через делается через аддон... а вот хер, без него падает
@REM go build -v -o "../app/kernel/SiYuan-Kernel.exe" -ldflags "-s -w -H=windowsgui" .
if errorlevel 1 (
    echo 'Building Kernel amd64 - FAILED'
    exit /b %errorlevel%
)
echo 'Building Kernel amd64 - OK'
