@echo off
chcp 65001 >nul
echo =============================================
echo Building Stock Photo App for Windows
echo =============================================
echo.

REM Проверяем наличие Wails
where wails >nul 2>&1
if %errorLevel% neq 0 (
    echo ❌ Wails не найден! Установите Wails:
    echo.
    echo go install github.com/wailsapp/wails/v2/cmd/wails@latest
    echo.
    pause
    exit /b 1
)

REM Проверяем наличие Go
where go >nul 2>&1
if %errorLevel% neq 0 (
    echo ❌ Go не найден! Скачайте с https://golang.org/dl/
    pause
    exit /b 1
)

REM Проверяем наличие Node.js
where npm >nul 2>&1
if %errorLevel% neq 0 (
    echo ❌ Node.js не найден! Скачайте с https://nodejs.org/
    pause
    exit /b 1
)

echo ✅ Все зависимости найдены
echo.

echo Собираем приложение для Windows...
wails build -platform windows/amd64 -clean

if %errorLevel% == 0 (
    echo.
    echo ✅ Сборка завершена успешно!
    
    REM Создаем папку релиза
    if not exist "releases\windows" mkdir "releases\windows"
    
    REM Копируем exe файл
    if exist "build\bin\stock-photo-app.exe" (
        copy "build\bin\stock-photo-app.exe" "releases\windows\"
        echo ✅ Файл скопирован в releases\windows\
    ) else (
        echo ❌ Файл stock-photo-app.exe не найден в build\bin\
    )
    
    echo.
    echo Релиз готов в папке: releases\windows\
    echo Не забудьте запустить install-dependencies.bat перед первым использованием!
) else (
    echo ❌ Ошибка при сборке
    exit /b 1
)

echo.
pause 