@echo off
setlocal enabledelayedexpansion

set "DEV_MODE=0"
if /i "%~1"=="dev" set "DEV_MODE=1"

if "!DEV_MODE!"=="1" (
    echo ============================================
    echo   WebGain Installer - Dev Build
    echo   Target: Windows x64 ^(amd64^)
    echo   No offuscamento / No compressione
    echo ============================================
) else (
    echo ============================================
    echo   WebGain Installer - Release Build
    echo   Target: Windows x64 ^(amd64^)
    echo   Garble + UPX
    echo ============================================
)
echo.

cd /d "%~dp0"

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
set "RELEASE_FILE=%~dp0release.txt"

echo       Aggiornamento PATH sessione corrente...
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\refresh_path.ps1"
if exist "%TEMP%\path_refresh.txt" (
    for /f "usebackq delims=" %%P in ("%TEMP%\path_refresh.txt") do set "PATH=%%P"
    del /q "%TEMP%\path_refresh.txt" 2>nul
)
echo.

echo [1/6] Pulizia e preparazione build...
if exist "build" (
    rmdir /s /q "build"
)
if exist "frontend\dist" (
    rmdir /s /q "frontend\dist"
)
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\init_build.ps1"
echo       Completato.
echo.

echo [2/6] Verifica strumenti di build...
set "toolsOk=1"
where go >nul 2>&1 || (
    echo       [!!] go non trovato
    set "toolsOk=0"
)
where wails >nul 2>&1 || (
    echo       [!!] wails non trovato
    set "toolsOk=0"
)
if "!DEV_MODE!"=="0" (
    where garble >nul 2>&1 || (
        echo       [!!] garble non trovato - installa con: go install mvdan.cc/garble@latest
        set "toolsOk=0"
    )
    where upx >nul 2>&1 || (
        echo       [!!] upx non trovato - installa con: winget install upx.upx
        set "toolsOk=0"
    )
)
if "!toolsOk!"=="0" (
    echo.
    echo ERRORE: strumenti mancanti. Esegui configure.bat o installa manualmente.
    exit /b 1
)
echo       Tutti gli strumenti presenti.
echo.

echo [3/6] Verifica dipendenze Go...
call go mod tidy
if !errorlevel! neq 0 (
    echo ERRORE: go mod tidy fallito.
    exit /b 1
)
echo       Completato.
echo.

echo [4/6] Verifica dipendenze frontend...
set "NEED_NPM=1"

if exist "frontend\package.json" if exist "frontend\package-lock.json" if exist "frontend\node_modules" (
    for /f "usebackq" %%A in (`powershell -NoProfile -Command "(@((Get-Item 'frontend\package.json').LastWriteTime, (Get-Item 'frontend\package-lock.json').LastWriteTime, (Get-Item 'frontend\node_modules').LastWriteTime) | Measure-Object -Maximum).Maximum.ToString('yyyyMMddHHmmss')"`) do set "CURRENT_DATE=%%A"

    if exist "!RELEASE_FILE!" (
        set /p SAVED_DATE=<"!RELEASE_FILE!"
        if "!CURRENT_DATE!" LEQ "!SAVED_DATE!" (
            set "NEED_NPM=0"
        )
    )
)

if "!NEED_NPM!"=="1" (
    echo       Installazione dipendenze frontend...
    cd frontend
    if exist "node_modules" (
        rmdir /s /q "node_modules"
    )
    if exist "package-lock.json" (
        del /f /q "package-lock.json"
    )
    call npm install
    if !errorlevel! neq 0 (
        echo ERRORE: npm install fallito.
        exit /b 1
    )
    cd ..
    for /f "usebackq" %%A in (`powershell -NoProfile -Command "(@((Get-Item 'frontend\package.json').LastWriteTime, (Get-Item 'frontend\package-lock.json').LastWriteTime, (Get-Item 'frontend\node_modules').LastWriteTime) | Measure-Object -Maximum).Maximum.ToString('yyyyMMddHHmmss')"`) do >"!RELEASE_FILE!" echo %%A
    echo       Completato.
) else (
    echo       Dipendenze frontend invariate, salto reinstallazione.
)
echo.

if "!DEV_MODE!"=="1" (
    echo [5/6] Compilazione sviluppo...
    call wails build -platform windows/amd64 -ldflags "-s -w" -trimpath -clean
) else (
    echo [5/6] Compilazione produzione con offuscamento e compressione...
    call wails build -platform windows/amd64 -ldflags "-s -w" -trimpath -clean -obfuscated -upx -upxflags "--best"
)
if !errorlevel! neq 0 (
    echo ERRORE: compilazione fallita.
    exit /b 1
)
echo       Completato.
echo.

echo [6/6] Verifica eseguibile...
if not exist "build\bin\WebGainInstaller.exe" (
    echo ERRORE: eseguibile non trovato!
    exit /b 1
)
for %%A in ("build\bin\WebGainInstaller.exe") do (
    set "fileSize=%%~zA"
)
set /a sizeMB=!fileSize! / 1048576
echo.
echo ============================================
echo   Build completata con successo!
echo   Eseguibile: build\bin\WebGainInstaller.exe
echo   Dimensione: circa !sizeMB! MB
if "!DEV_MODE!"=="1" (
    echo   Modalita': sviluppo
) else (
    echo   Offuscamento: garble
    echo   Compressione: UPX --best
)
echo ============================================

endlocal
