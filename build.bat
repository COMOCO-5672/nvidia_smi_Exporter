@echo off
rem 获取git上提交号，该提交号为commander提交版本号
for /F %%i in ('git rev-parse --short HEAD') do ( set commitid=%%i)
echo commitid=%commitid%
rem 在文件version中获取版本号信息（官网上这样写的）
for /f "tokens=*" %%a in (./src/VERSION) do ( set version=%%a)
echo version=%version%
rem 获取git分支情况
for /F %%b in ('git symbolic-ref --short -q HEAD') do ( set Branch=%%b )
echo Branch=%Branch%
rem 获取创建时间
set "creattime=%date:~0,4%%date:~5,2%%date:~8,2%%time:~0,2%%time:~3,2%%time:~6,2%"
echo %creattime%
rem 开始build
cd ./src
echo %chdir%
go build -ldflags "-X github.com/prometheus/common/version.BuildDate=%creattime% -X github.com/prometheus/common/version.Version=%version% -X github.com/prometheus/common/version.BuildUser=TR -X github.com/prometheus/common/version.Revision=%commitid%  -X github.com/prometheus/common/version.Branch=%Branch%" -o ../

pause