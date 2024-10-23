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

for /L %%i in (1,1,6) do (
    for /L %%j in (1,1,10) do (
        go run %fileList% %%i %%j 

        echo Executed Test: %%i : %%j
    )
)