#!/bin/bash

echo "============================================"
echo "Stock Photo App - macOS Dependencies Setup"
echo "============================================"

# Проверяем есть ли Homebrew
if ! command -v brew &> /dev/null; then
    echo "❌ Homebrew не найден!"
    echo "Устанавливаем Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    
    # Добавляем Homebrew в PATH для Apple Silicon Macs
    if [[ $(uname -m) == 'arm64' ]]; then
        echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zshrc
        eval "$(/opt/homebrew/bin/brew shellenv)"
    fi
else
    echo "✅ Homebrew найден"
fi

# Устанавливаем ExifTool
echo ""
echo "Устанавливаем ExifTool..."
if brew list exiftool &> /dev/null; then
    echo "✅ ExifTool уже установлен"
    brew upgrade exiftool || true
else
    brew install exiftool
    echo "✅ ExifTool установлен"
fi

# Проверяем установку
echo ""
echo "Проверяем установку..."
if command -v exiftool &> /dev/null; then
    EXIFTOOL_VERSION=$(exiftool -ver)
    echo "✅ ExifTool версия: $EXIFTOOL_VERSION"
else
    echo "❌ ExifTool не найден в PATH"
    exit 1
fi

echo ""
echo "============================================"
echo "✅ Все зависимости установлены успешно!"
echo "============================================"
echo ""
echo "Теперь вы можете запустить Stock Photo App:"
echo "1. Откройте приложение stock-photo-app.app"
echo "2. Или запустите из терминала: open stock-photo-app.app"
echo ""
echo "Примечание: При первом запуске macOS может попросить"
echo "разрешение на доступ к файлам и папкам." 