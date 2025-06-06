# Сводка по созданию релизов Stock Photo App

## Что было создано

### ✅ Готовые релизы для macOS

1. **macOS Intel (x64)**
   - Файл: `releases/stock-photo-app-macos-intel.zip` (11 MB)
   - Содержит: приложение + скрипт установки + README
   - Совместимость: macOS 10.15+ на Intel процессорах

2. **macOS Apple Silicon (ARM64)**
   - Файл: `releases/stock-photo-app-macos-arm64.zip` (11 MB)
   - Содержит: приложение + скрипт установки + README
   - Совместимость: macOS 11.0+ на Apple Silicon (M1/M2/M3)

### ✅ Готовый релиз для Windows

3. **Windows x64**
   - Файл: `releases/stock-photo-app-windows.zip` (~9 MB)
   - Содержит: готовое приложение + скрипты установки + README
   - Собрано с помощью pure Go SQLite драйвера (modernc.org/sqlite)

## Созданные файлы

### Скрипты сборки
- `build-macos.sh` - автоматическая сборка для обеих архитектур macOS
- `build-windows.bat` - скрипт сборки для Windows (запускать на Windows)

### Скрипты установки зависимостей
- `releases/macos-*/install-dependencies.sh` - установка Homebrew + ExifTool
- `releases/windows/install-dependencies.bat` - установка Chocolatey + ExifTool

### Документация
- `releases/README.md` - общий обзор всех релизов
- `releases/macos-*/README.md` - инструкции для macOS
- `releases/windows/README.md` - инструкции для Windows
- `BUILD_INSTRUCTIONS.md` - подробные инструкции по сборке

## Процесс сборки

### macOS (выполнено)
```bash
# Автоматическая сборка
./build-macos.sh

# Или ручная сборка
wails build -platform darwin/amd64 -clean  # Intel
wails build -platform darwin/arm64         # Apple Silicon
```

### Windows (требует Windows машину)
```cmd
# На Windows машине
build-windows.bat
```

## Установка для пользователей

### macOS
1. Скачать подходящий ZIP архив
2. Распаковать
3. Запустить `./install-dependencies.sh`
4. Открыть `stock-photo-app.app`

### Windows
1. Скачать `stock-photo-app-windows.zip`
2. Распаковать архив
3. Запустить `install-dependencies.bat` (от администратора)
4. Запустить `stock-photo-app.exe`

## Размеры релизов

- **macOS Intel**: 11 MB (сжато)
- **macOS ARM64**: 11 MB (сжато)
- **Windows**: ~9 MB (сжато, с exe файлом 27 MB)

## Зависимости

Все версии требуют **ExifTool** для работы с EXIF метаданными:
- **macOS**: Устанавливается через Homebrew
- **Windows**: Устанавливается через Chocolatey

## Следующие шаги

### Для распространения
1. Загрузить ZIP архивы на файлообменник или GitHub Releases
2. Создать инструкции по скачиванию
3. Рассмотреть подписание кода для macOS (для избежания предупреждений безопасности)

## Техническая информация

- **Фреймворк**: Wails v2.10.1
- **Go версия**: 1.21+
- **Node.js**: 16+
- **Архитектуры**: Intel x64, Apple Silicon ARM64, Windows x64
- **Размер приложения**: ~32 MB (несжатый), ~11 MB (сжатый)

## Проблемы и решения

### Cross-компиляция Windows с macOS
❌ **Проблема**: CGO зависимости (SQLite) не позволяют cross-компиляцию  
✅ **Решение**: Замена на pure Go SQLite драйвер (modernc.org/sqlite) для Windows билда

### Зависимости ExifTool
❌ **Проблема**: ExifTool не входит в состав приложения  
✅ **Решение**: Автоматические скрипты установки для каждой платформы

### Безопасность macOS
❌ **Проблема**: Предупреждения о неподписанном приложении  
✅ **Решение**: Инструкции в README + возможность подписания кода в будущем 