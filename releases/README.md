# Stock Photo App - Релизы

Автоматизация работы фотографа со стоковыми площадками.

## Доступные версии

### macOS

#### Intel процессоры (x64)
- **Файл**: `stock-photo-app-macos-intel.zip`
- **Размер**: ~11 MB
- **Совместимость**: macOS 10.15+ на Intel процессорах
- **Содержимое**: 
  - `stock-photo-app.app` - основное приложение
  - `install-dependencies.sh` - скрипт установки зависимостей
  - `README.md` - инструкции по установке

#### Apple Silicon (ARM64)
- **Файл**: `stock-photo-app-macos-arm64.zip`
- **Размер**: ~11 MB
- **Совместимость**: macOS 11.0+ на Apple Silicon (M1/M2/M3)
- **Содержимое**: 
  - `stock-photo-app.app` - основное приложение
  - `install-dependencies.sh` - скрипт установки зависимостей
  - `README.md` - инструкции по установке

### Windows

#### x64 (64-bit)
- **Файл**: `stock-photo-app-windows.zip`
- **Размер**: ~9 MB (полный релиз с приложением)
- **Совместимость**: Windows 10+ (64-bit)
- **Содержимое**: 
  - `stock-photo-app.exe` - готовое приложение (27 MB)
  - `install-dependencies.bat` - скрипт установки зависимостей
  - `README.md` - инструкции по установке

## Быстрый старт

### macOS

1. Скачайте подходящую версию:
   - Intel Mac: `stock-photo-app-macos-intel.zip`
   - Apple Silicon: `stock-photo-app-macos-arm64.zip`

2. Распакуйте архив

3. Установите зависимости:
   ```bash
   chmod +x install-dependencies.sh
   ./install-dependencies.sh
   ```

4. Запустите приложение:
   ```bash
   open stock-photo-app.app
   ```

### Windows

1. Скачайте `stock-photo-app-windows.zip`

2. Распакуйте архив

3. Установите зависимости (от имени администратора):
   ```cmd
   install-dependencies.bat
   ```

4. Запустите приложение:
   ```cmd
   stock-photo-app.exe
   ```

## Системные требования

### macOS
- **Intel**: macOS 10.15 (Catalina) или новее
- **Apple Silicon**: macOS 11.0 (Big Sur) или новее
- 4 GB RAM минимум, 8 GB рекомендуется
- 500 MB свободного места

### Windows
- Windows 10 или новее (64-bit)
- 4 GB RAM минимум, 8 GB рекомендуется
- 500 MB свободного места
- .NET Framework 4.7.2+

## Зависимости

Все версии требуют **ExifTool** для работы с метаданными изображений.

### Автоматическая установка
- **macOS**: Homebrew + ExifTool
- **Windows**: Chocolatey + ExifTool

### Ручная установка
- **ExifTool**: https://exiftool.org/

## Поддержка

### Проблемы с безопасностью

#### macOS
При первом запуске может появиться предупреждение:
1. Системные настройки → Безопасность и конфиденциальность
2. Нажмите "Все равно открыть"

#### Windows
Windows Defender может заблокировать:
1. Нажмите "Подробнее" → "Выполнить в любом случае"
2. Или добавьте в исключения антивируса

### Логи и отладка

#### macOS
```bash
# Просмотр логов
Console.app → Поиск: "stock-photo-app"

# Запуск из терминала для отладки
./stock-photo-app.app/Contents/MacOS/stock-photo-app
```

#### Windows
```cmd
# Просмотр логов
eventvwr.msc → Windows Logs → Application

# Запуск из командной строки для отладки
stock-photo-app.exe
```

## Версии

- **v1.0.0** - Первый релиз
  - AI анализ изображений
  - Batch обработка
  - EXIF метаданные
  - Поддержка Editorial/Commercial категорий
  - Bulk операции (approve/reject/regenerate all)

## Лицензия

Приложение создано для личного использования автором-фотографом.

## Техническая информация

- **Фреймворк**: Wails v2.10.1 (Go + Web)
- **Backend**: Go 1.21+
- **Frontend**: Vanilla JS + Tailwind CSS
- **База данных**: SQLite
- **AI**: OpenAI GPT-4 Vision / Claude Vision 