# Настройка GitHub репозитория

## Создание репозитория

### 1. Подготовка проекта
```bash
# Инициализируем Git (если еще не сделано)
git init

# Добавляем все файлы (кроме исключенных в .gitignore)
git add .

# Создаем первый коммит
git commit -m "Initial commit: Stock Photo Automation App

- Complete Wails v2.10.1 application
- AI-powered photo analysis with GPT-4 Vision & Claude
- Batch processing with drag & drop
- EXIF metadata writing
- Multi-platform releases (macOS Intel/ARM64, Windows)
- FTP/SFTP upload support
- Editorial/Commercial category support
- Comprehensive build scripts and documentation"
```

### 2. Создание GitHub репозитория

**Вариант A: Через GitHub CLI**
```bash
# Установите GitHub CLI если нет: brew install gh
gh auth login
gh repo create stock-photo-app --public --description "AI-powered stock photo automation tool"
git remote add origin https://github.com/ВАШ_USERNAME/stock-photo-app.git
git push -u origin main
```

**Вариант B: Через веб-интерфейс**
1. Перейдите на https://github.com/new
2. Название: `stock-photo-app`
3. Описание: `AI-powered stock photo automation tool`
4. Выберите Public или Private
5. НЕ создавайте README, .gitignore или license (у нас уже есть)
6. Нажмите "Create repository"

```bash
git remote add origin https://github.com/ВАШ_USERNAME/stock-photo-app.git
git branch -M main
git push -u origin main
```

## Настройка релизов

### GitHub Releases

1. Перейдите в ваш репозиторий на GitHub
2. Нажмите "Releases" → "Create a new release"
3. Tag: `v1.0.0`
4. Title: `v1.0.0 - Initial Release`
5. Описание:
```markdown
# Stock Photo Automation App v1.0.0

🎉 Первый релиз приложения для автоматизации работы со стоковыми площадками!

## 📦 Загрузки

- **macOS Intel**: [stock-photo-app-macos-intel.zip](releases/stock-photo-app-macos-intel.zip)
- **macOS Apple Silicon**: [stock-photo-app-macos-arm64.zip](releases/stock-photo-app-macos-arm64.zip)
- **Windows 64-bit**: [stock-photo-app-windows.zip](releases/stock-photo-app-windows.zip)

## ✨ Основные возможности

- 🤖 **AI анализ** с GPT-4 Vision и Claude
- 📸 **Batch обработка** папок с фотографиями
- 🏷️ **Автоматические метаданные** (48-55 ключевых слов)
- 📁 **Editorial/Commercial** категории
- 🔄 **EXIF запись** метаданных
- 📤 **FTP/SFTP загрузка** на стоки
- 🚀 **Bulk операции** (approve/reject/regenerate all)

## 📋 Системные требования

- **macOS**: 10.15+ (Intel) или 11.0+ (Apple Silicon)
- **Windows**: 10+ (64-bit)
- **ExifTool** (устанавливается автоматически)

## 🚀 Быстрый старт

1. Скачайте подходящий архив
2. Распакуйте
3. Запустите скрипт установки зависимостей
4. Откройте приложение
5. Настройте AI API ключи

📖 Подробные инструкции в [releases/README.md](releases/README.md)
```

6. Загрузите файлы релизов:
   - `releases/stock-photo-app-macos-intel.zip`
   - `releases/stock-photo-app-macos-arm64.zip`
   - `releases/stock-photo-app-windows.zip`

## Настройка автоматических билдов (CI/CD)

### GitHub Actions

Создайте `.github/workflows/build.yml`:

```yaml
name: Build Releases

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:

jobs:
  build-macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
        
      - name: Build macOS releases
        run: ./build-macos.sh
        
      - name: Upload macOS artifacts
        uses: actions/upload-artifact@v4
        with:
          name: macos-builds
          path: releases/stock-photo-app-macos-*.zip

  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
        
      - name: Build Windows
        run: build-windows.bat
        
      - name: Upload Windows artifacts
        uses: actions/upload-artifact@v4
        with:
          name: windows-build
          path: releases/stock-photo-app-windows.zip
```

## Безопасность

### ✅ Что ВКЛЮЧЕНО в репозиторий:
- Исходный код приложения
- Документация и инструкции
- Скрипты сборки
- Конфигурационные файлы
- Пример API интеграций

### ❌ Что ИСКЛЮЧЕНО из репозитория:
- `app.db` - база данных с API ключами
- `temp/` - временные файлы и миниатюры
- `releases/*.zip` - готовые сборки
- `build/bin/` - скомпилированные бинарники
- `.DS_Store` и другие системные файлы

### 🔒 Проверка безопасности

Перед публикацией убедитесь:

```bash
# Проверяем что чувствительные файлы игнорируются
git check-ignore app.db temp/ releases/*.zip

# Проверяем что будет загружено
git ls-files | grep -E "(\.db|api|key|secret|password)" || echo "✅ Безопасно"

# Проверяем размер репозитория
du -sh .git
```

## Поддержка пользователей

### Issues
- Настройте шаблоны для bug reports и feature requests
- Добавьте labels: `bug`, `enhancement`, `question`, `help wanted`

### Discussions
- Включите GitHub Discussions для сообщества
- Разделы: General, Q&A, Show and tell, Feature requests

### Wiki
- Создайте Wiki с детальной документацией
- Страницы: Installation, Configuration, API Setup, Troubleshooting

## Лицензия

Добавьте подходящую лицензию в файл `LICENSE`:
- MIT - для открытого использования
- Apache 2.0 - для коммерческого использования
- GPL v3 - для копилефт проектов 