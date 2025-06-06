@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ============================================
echo Stock Photo App - Windows Dependencies Setup
echo ============================================
echo.

REM Проверяем права администратора
net session >nul 2>&1
if %errorLevel% == 0 (
    echo ✅ Права администратора подтверждены
) else (
    echo ❌ ОШИБКА: Запустите этот скрипт как администратор!
    echo Правый клик → "Запуск от имени администратора"
    pause
    exit /b 1
)

REM Проверяем наличие ExifTool
where exiftool >nul 2>&1
if %errorLevel% == 0 (
    echo ✅ ExifTool уже установлен
    for /f "delims=" %%i in ('exiftool -ver') do set EXIFTOOL_VERSION=%%i
    echo Версия: !EXIFTOOL_VERSION!
    goto :check_scoop
)

echo.
echo Устанавливаем ExifTool...

REM Проверяем наличие Scoop (пакетный менеджер для Windows)
where scoop >nul 2>&1
if %errorLevel% == 0 (
    echo ✅ Scoop найден
    scoop install exiftool
    goto :verify_install
)

REM Проверяем наличие Chocolatey
where choco >nul 2>&1
if %errorLevel% == 0 (
    echo ✅ Chocolatey найден
    choco install exiftool -y
    goto :verify_install
)

echo.
echo Устанавливаем Chocolatey...
powershell -Command "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))"

REM Обновляем PATH
call refreshenv

echo.
echo Устанавливаем ExifTool через Chocolatey...
choco install exiftool -y

:verify_install
echo.
echo Проверяем установку...
where exiftool >nul 2>&1
if %errorLevel% == 0 (
    for /f "delims=" %%i in ('exiftool -ver') do set EXIFTOOL_VERSION=%%i
    echo ✅ ExifTool версия: !EXIFTOOL_VERSION!
) else (
    echo ❌ ExifTool не найден в PATH
    echo.
    echo Возможные решения:
    echo 1. Перезагрузите компьютер и попробуйте снова
    echo 2. Вручную скачайте ExifTool с https://exiftool.org/
    echo 3. Добавьте ExifTool в переменную PATH
    pause
    exit /b 1
)

:check_scoop
echo.
echo ============================================
echo ✅ Все зависимости установлены успешно!
echo ============================================
echo.
echo Теперь вы можете запустить Stock Photo App:
echo 1. Дважды кликните на stock-photo-app.exe
echo 2. Или запустите из командной строки: stock-photo-app.exe
echo.
echo Примечание: Windows Defender может показать предупреждение
echo о неизвестном приложении. Нажмите "Подробнее" → "Выполнить в любом случае"
echo.
pause 