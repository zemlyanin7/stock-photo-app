#!/bin/bash

echo "============================================"
echo "Building Stock Photo App for macOS"
echo "============================================"

# Проверяем зависимости
if ! command -v wails &> /dev/null; then
    echo "❌ Wails не найден! Установите:"
    echo "go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "❌ Go не найден! Скачайте с https://golang.org/dl/"
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "❌ Node.js не найден! Скачайте с https://nodejs.org/"
    exit 1
fi

echo "✅ Все зависимости найдены"
echo ""

# Создаем папки релизов
mkdir -p releases/macos-intel releases/macos-arm64

# Сборка для Intel
echo "🔨 Сборка для Intel (x64)..."
wails build -platform darwin/amd64 -clean

if [ $? -eq 0 ]; then
    cp -r build/bin/stock-photo-app.app releases/macos-intel/
    echo "✅ Intel версия готова"
else
    echo "❌ Ошибка сборки для Intel"
    exit 1
fi

# Сборка для Apple Silicon
echo ""
echo "🔨 Сборка для Apple Silicon (ARM64)..."
wails build -platform darwin/arm64

if [ $? -eq 0 ]; then
    cp -r build/bin/stock-photo-app.app releases/macos-arm64/
    echo "✅ ARM64 версия готова"
else
    echo "❌ Ошибка сборки для ARM64"
    exit 1
fi

echo ""
echo "============================================"
echo "✅ Все сборки завершены успешно!"
echo "============================================"
echo ""
echo "Релизы готовы:"
echo "- macOS Intel: releases/macos-intel/"
echo "- macOS ARM64: releases/macos-arm64/"
echo ""
echo "Не забудьте запустить install-dependencies.sh"
echo "в каждой папке перед первым использованием!" 