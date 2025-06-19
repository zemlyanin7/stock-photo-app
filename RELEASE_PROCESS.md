# Процесс создания релизов для GitHub

## 🚀 Создание нового релиза

### 1. Подготовка

1. **Обновите CHANGELOG.md**:
   - Переместите изменения из `[Unreleased]` в новую версию
   - Добавьте дату релиза
   - Создайте новую секцию `[Unreleased]`

2. **Соберите приложения**:
   ```bash
   # macOS Intel
   wails build -platform darwin/amd64 -clean
   cd build/bin && zip -r ../../stock-photo-app-macos-intel.zip stock-photo-app.app && cd ../..
   
   # macOS Apple Silicon  
   wails build -platform darwin/arm64
   cd build/bin && zip -r ../../stock-photo-app-macos-arm64.zip stock-photo-app.app && cd ../..
   
   # Windows (требует Windows машину или cross-compilation)
   wails build -platform windows/amd64
   cd build/bin && zip -r ../../stock-photo-app-windows.zip stock-photo-app.exe && cd ../..
   ```

### 2. Создание релиза на GitHub

1. **Перейдите в раздел Releases**: https://github.com/[username]/[repo]/releases

2. **Нажмите "Create a new release"**

3. **Заполните форму**:
   - **Tag version**: `v1.0.0` (следуйте Semantic Versioning)
   - **Release title**: `Stock Photo App v1.0.0`
   - **Description**: Скопируйте из CHANGELOG.md для этой версии

4. **Прикрепите файлы**:
   - `stock-photo-app-macos-intel.zip`
   - `stock-photo-app-macos-arm64.zip`
   - `stock-photo-app-windows.zip`

5. **Опубликуйте релиз**

### 3. Пример описания релиза

```markdown
# Stock Photo Automation App v1.0.0

Первый официальный релиз приложения для автоматизации работы со стоковыми фотографиями.

## 🚀 Основные возможности

- **AI анализ фотографий** с помощью OpenAI GPT-4 Vision
- **Batch обработка** папок с фотографиями
- **Editorial и Commercial** типы контента с разными промптами
- **EXIF запись** метаданных в файлы изображений
- **Модульная система загрузки** на стоковые площадки
- **Детальное логирование** всех операций

## 📥 Установка

### macOS
1. Скачайте подходящую версию:
   - **Intel Mac**: `stock-photo-app-macos-intel.zip`
   - **Apple Silicon (M1/M2/M3)**: `stock-photo-app-macos-arm64.zip`
2. Распакуйте и запустите `stock-photo-app.app`
3. Установите ExifTool: `brew install exiftool`

### Windows
1. Скачайте `stock-photo-app-windows.zip`
2. Распакуйте и запустите `stock-photo-app.exe`
3. Установите ExifTool с официального сайта

## 📚 Документация

- [Полная документация](DOCUMENTATION.md)
- [История изменений](CHANGELOG.md)
- [Инструкции по установке](releases/README.md)

## ⚠️ Системные требования

- **macOS**: 10.15+ (Intel) или 11.0+ (Apple Silicon)
- **Windows**: 10+ (64-bit)
- **ExifTool**: для записи метаданных

## 🐛 Сообщить о проблеме

Если вы нашли баг или у вас есть предложения по улучшению, создайте [issue](../../issues/new).
```

## 📋 Чеклист для релиза

### Перед созданием релиза:
- [ ] Все функции протестированы
- [ ] Обновлен CHANGELOG.md
- [ ] Версия указана в wails.json
- [ ] Собраны все платформы
- [ ] Проверены размеры архивов
- [ ] Готово описание релиза

### После создания релиза:
- [ ] Ссылки в README работают
- [ ] Все файлы скачиваются корректно
- [ ] Проверена установка на разных платформах
- [ ] Обновлена документация при необходимости

## 🔄 Semantic Versioning

Мы используем [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0): Несовместимые изменения API
- **MINOR** (1.1.0): Новая функциональность (обратно совместимая)
- **PATCH** (1.0.1): Исправления багов

## 📦 Что НЕ хранить в Git

- ✅ **Хранить**: Документацию, инструкции, скрипты сборки
- ❌ **НЕ хранить**: ZIP файлы, исполняемые файлы, собранные приложения
- ❌ **НЕ хранить**: Файлы релизов (они должны быть только в GitHub Releases)

## 🛠️ Автоматизация (будущее)

Можно настроить GitHub Actions для автоматической сборки релизов:

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags:
      - 'v*'
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v3
      - name: Build releases
        run: |
          # сборка для всех платформ
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            stock-photo-app-*.zip
``` 