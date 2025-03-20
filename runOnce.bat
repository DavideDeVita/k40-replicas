@echo off
setlocal enabledelayedexpansion

set "fileList="  REM Initialize an empty string

REM Loop through all .go files in the cyrr directory
for %%f in (.\*.go) do (
    set "fileList=!fileList! %%~nxf"  REM Append the file name to the string
)

REM Remove leading space (optional)
set "fileList=!fileList:~1!"

echo %fileList%

go run %fileList% 0 0 
