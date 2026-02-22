@echo off
setlocal

echo ============================================
echo   WebGain Installer - Clean
echo ============================================
echo.

cd /d "%~dp0"

echo [1/5] Rimozione cartella build...
if exist "build" (
    rmdir /s /q "build"
    echo       Rimossa.
) else (
    echo       Non presente, salto.
)
echo.

echo [2/5] Rimozione frontend\dist...
if exist "frontend\dist" (
    rmdir /s /q "frontend\dist"
    echo       Rimossa.
) else (
    echo       Non presente, salto.
)
echo.

echo [3/5] Rimozione frontend\node_modules...
if exist "frontend\node_modules" (
    rmdir /s /q "frontend\node_modules"
    echo       Rimossa.
) else (
    echo       Non presente, salto.
)
echo.

echo [4/5] Rimozione frontend\package-lock.json...
if exist "frontend\package-lock.json" (
    del /f /q "frontend\package-lock.json"
    echo       Rimosso.
) else (
    echo       Non presente, salto.
)
echo.

echo [5/5] Rimozione release.txt...
if exist "release.txt" (
    del /f /q "release.txt"
    echo       Rimosso.
) else (
    echo       Non presente, salto.
)
echo.

echo ============================================
echo   Pulizia completata!
echo ============================================

endlocal
