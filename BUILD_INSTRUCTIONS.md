# Инструкции по сборке Stock Photo App

## Требования для сборки

### Общие требования
- Go 1.21 или новее
- Node.js 16 или новее
- Wails CLI v2.10.1

### Установка Wails CLI
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Сборка для macOS

### Автоматическая сборка (рекомендуется)
```bash
chmod +x build-macos.sh
./build-macos.sh
```

Этот скрипт создаст оба варианта:
- `releases/macos-intel/` - для Intel процессоров
- `releases/macos-arm64/` - для Apple Silicon (M1/M2/M3)

### Ручная сборка

#### Intel (x64)
```bash
wails build -platform darwin/amd64 -clean
mkdir -p releases/macos-intel
cp -r build/bin/stock-photo-app.app releases/macos-intel/
```

#### Apple Silicon (ARM64)
```bash
wails build -platform darwin/arm64
mkdir -p releases/macos-arm64  
cp -r build/bin/stock-photo-app.app releases/macos-arm64/
```

## Сборка для Windows

⚠️ **Важно**: Сборка для Windows должна происходить **на Windows машине**

### На Windows машине

1. Установите зависимости:
   - Go: https://golang.org/dl/
   - Node.js: https://nodejs.org/
   - Git: https://git-scm.com/

2. Установите Wails:
   ```cmd
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

3. Клонируйте репозиторий:
   ```cmd
   git clone <repository-url>
   cd stock-photo-app
   ```

4. Запустите сборку:
   ```cmd
   build-windows.bat
   ```

### Cross-компиляция с macOS (не рекомендуется)

Cross-компиляция Go приложений с CGO зависимостями (SQLite) с macOS на Windows может быть проблематичной. Рекомендуется использовать Windows машину или виртуальную машину.

## Структура релизов

После сборки структура папок будет:

```
releases/
├── macos-intel/
│   ├── stock-photo-app.app
│   ├── install-dependencies.sh
│   └── README.md
├── macos-arm64/
│   ├── stock-photo-app.app
│   ├── install-dependencies.sh
│   └── README.md
└── windows/
    ├── stock-photo-app.exe
    ├── install-dependencies.bat
    └── README.md
```

## Зависимости

### macOS
- ExifTool (устанавливается через Homebrew)

### Windows  
- ExifTool (устанавливается через Chocolatey)

### Автоматическая установка зависимостей

В каждой папке релиза есть скрипты для автоматической установки:

- **macOS**: `install-dependencies.sh`
- **Windows**: `install-dependencies.bat` (требует права администратора)

## Troubleshooting

### macOS

#### Ошибка: "cannot execute binary file"
Проверьте архитектуру:
```bash
file build/bin/stock-photo-app.app/Contents/MacOS/stock-photo-app
lipo -info build/bin/stock-photo-app.app/Contents/MacOS/stock-photo-app
```

#### Ошибка подписи кода
```bash
codesign --force --deep --sign - build/bin/stock-photo-app.app
```

### Windows

#### CGO ошибки при cross-компиляции
Используйте Windows машину для сборки или Docker с Windows контейнером:

```dockerfile
FROM mcr.microsoft.com/windows/servercore:ltsc2022
# Установка Go, Node.js, Git
# Клонирование и сборка проекта
```

#### Wails ошибки на Windows
Убедитесь что установлены:
- Microsoft Build Tools
- Windows SDK
- Git

```cmd
# Проверка установки
wails doctor
```

## Подписание и нотаризация (macOS)

Для распространения вне App Store:

```bash
# Подписание
codesign --force --options runtime --sign "Developer ID Application: Your Name" build/bin/stock-photo-app.app

# Создание DMG
hdiutil create -volname "Stock Photo App" -srcfolder build/bin/stock-photo-app.app -ov -format UDZO stock-photo-app.dmg

# Нотаризация (требует Apple Developer аккаунт)
xcrun notarytool submit stock-photo-app.dmg --keychain-profile "notarytool-profile" --wait
```

## CI/CD

### GitHub Actions пример

```yaml
name: Build Releases
on: [push, pull_request]

jobs:
  build-macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
      - name: Build
        run: ./build-macos.sh
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: macos-builds
          path: releases/

  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
      - name: Build
        run: build-windows.bat
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: windows-build
          path: releases/windows/
``` 