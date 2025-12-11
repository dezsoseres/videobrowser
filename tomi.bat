del *.rar
del *.exe
for /f "tokens=1-3 delims=. " %%a in ('date /t') do (
    set yyyy=%%a
    set mm=%%b
    set dd=%%c
)
set DATE_YYYYMMDD=%yyyy%%mm%%dd%
echo %DATE_YYYYMMDD%
c:\"program files"\seresd\winrar\rar a -r -s videobrowser%DATE_YYYYMMDD%







