@echo off
setlocal enabledelayedexpansion

echo ============================================
echo   WebGain Installer - Configurazione
echo   Ambiente di Sviluppo
echo ============================================
echo.
echo   Questo script installa tutti gli strumenti
echo   necessari per compilare WebGain Installer.
echo.
echo   Requisiti:
echo     - Windows 10/11 o Server 2019+
echo     - Connessione internet
echo     - Diritti amministrativi
echo ============================================
echo.

net session >nul 2>&1
if !errorlevel! neq 0 (
    echo ERRORE: Questo script richiede diritti amministrativi.
    echo Eseguire come Amministratore.
    pause
    exit /b 1
)

where winget >nul 2>&1
if !errorlevel! neq 0 (
    echo ERRORE: winget non trovato. Installare App Installer dal Microsoft Store.
    echo In alternativa, installare manualmente i componenti elencati sotto.
    echo.
    echo Componenti necessari:
    echo   - Go ^>= 1.22      https://go.dev/dl/
    echo   - Node.js ^>= 20   https://nodejs.org/
    echo   - Wails CLI v2     go install github.com/wailsapp/wails/v2/cmd/wails@latest
    echo   - Garble           go install mvdan.cc/garble@latest
    echo   - UPX              https://github.com/upx/upx/releases
    echo   - Git              https://git-scm.com/
    echo.
    pause
    exit /b 1
)

echo [1/7] Installazione Go...
where go >nul 2>&1
if !errorlevel! equ 0 (
    echo       Go gia' presente:
    go version
) else (
    winget install -e --id GoLang.Go --accept-source-agreements --accept-package-agreements
    if !errorlevel! neq 0 (
        echo ATTENZIONE: installazione Go fallita, provare manualmente.
    ) else (
        echo       Installato.
    )
)
echo.

echo [2/7] Installazione Node.js (LTS)...
where node >nul 2>&1
if !errorlevel! equ 0 (
    echo       Node.js gia' presente:
    node --version
) else (
    winget install -e --id OpenJS.NodeJS.LTS --accept-source-agreements --accept-package-agreements
    if !errorlevel! neq 0 (
        echo ATTENZIONE: installazione Node.js fallita, provare manualmente.
    ) else (
        echo       Installato.
    )
)
echo.

echo [3/7] Installazione Git...
where git >nul 2>&1
if !errorlevel! equ 0 (
    echo       Git gia' presente:
    git --version
) else (
    winget install -e --id Git.Git --accept-source-agreements --accept-package-agreements
    if !errorlevel! neq 0 (
        echo ATTENZIONE: installazione Git fallita, provare manualmente.
    ) else (
        echo       Installato.
    )
)
echo.

echo [4/7] Installazione UPX (compressione eseguibili)...
where upx >nul 2>&1
if !errorlevel! equ 0 (
    echo       UPX gia' presente:
    upx --version 2>nul | findstr /r "^upx"
) else (
    winget install -e --id UPX.UPX --accept-source-agreements --accept-package-agreements
    if !errorlevel! neq 0 (
        echo ATTENZIONE: installazione UPX fallita.
        echo Installare manualmente da: https://github.com/upx/upx/releases
    ) else (
        echo       Installato.
    )
)
echo.

call :RefreshPath

echo [5/7] Installazione Wails CLI...
where wails >nul 2>&1
if !errorlevel! equ 0 (
    echo       Wails gia' presente:
    wails version
) else (
    call go install github.com/wailsapp/wails/v2/cmd/wails@latest
    if !errorlevel! neq 0 (
        echo ATTENZIONE: installazione Wails fallita.
        echo Assicurarsi che Go sia installato e nel PATH, poi eseguire:
        echo   go install github.com/wailsapp/wails/v2/cmd/wails@latest
    ) else (
        echo       Installato.
    )
)
echo.

echo [6/7] Installazione Garble (offuscamento Go)...
where garble >nul 2>&1
if !errorlevel! equ 0 (
    echo       Garble gia' presente:
    garble version
) else (
    call go install mvdan.cc/garble@latest
    if !errorlevel! neq 0 (
        echo ATTENZIONE: installazione Garble fallita.
        echo Assicurarsi che Go sia installato e nel PATH, poi eseguire:
        echo   go install mvdan.cc/garble@latest
    ) else (
        echo       Installato.
    )
)
echo.

echo [7/7] Verifica finale...
call :RefreshPath

set "allOk=1"

where go >nul 2>&1
if !errorlevel! equ 0 (
    echo   [OK] Go
) else (
    echo   [!!] Go - NON TROVATO
    set "allOk=0"
)

where node >nul 2>&1
if !errorlevel! equ 0 (
    echo   [OK] Node.js
) else (
    echo   [!!] Node.js - NON TROVATO
    set "allOk=0"
)

where npm >nul 2>&1
if !errorlevel! equ 0 (
    echo   [OK] npm
) else (
    echo   [!!] npm - NON TROVATO
    set "allOk=0"
)

where git >nul 2>&1
if !errorlevel! equ 0 (
    echo   [OK] Git
) else (
    echo   [!!] Git - NON TROVATO
    set "allOk=0"
)

where wails >nul 2>&1
if !errorlevel! equ 0 (
    echo   [OK] Wails CLI
) else (
    echo   [!!] Wails CLI - NON TROVATO
    set "allOk=0"
)

where garble >nul 2>&1
if !errorlevel! equ 0 (
    echo   [OK] Garble
) else (
    echo   [!!] Garble - NON TROVATO
    set "allOk=0"
)

where upx >nul 2>&1
if !errorlevel! equ 0 (
    echo   [OK] UPX
) else (
    echo   [!!] UPX - NON TROVATO
    set "allOk=0"
)

echo.
if "!allOk!"=="1" (
    echo   Ambiente configurato correttamente!
    echo   Puoi compilare con: release.bat
) else (
    echo   ATTENZIONE: alcuni componenti mancano.
    echo   Chiudi il terminale, riaprilo e rilancia questo script.
    echo   Se il problema persiste, installare manualmente.
)
echo.
echo ============================================

pause
endlocal
goto :eof

:RefreshPath
echo       Aggiornamento PATH sessione corrente...
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\refresh_path.ps1"
if exist "%TEMP%\path_refresh.txt" (
    for /f "usebackq delims=" %%P in ("%TEMP%\path_refresh.txt") do set "PATH=%%P"
    del /q "%TEMP%\path_refresh.txt" 2>nul
    echo       PATH aggiornato con successo.
) else (
    echo       ATTENZIONE: impossibile aggiornare PATH automaticamente.
)
echo.
goto :eof
