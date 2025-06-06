# Настройка Shutterstock FTP

## 📋 Обзор

Shutterstock предоставляет FTP доступ для загрузки больших объемов изображений. Этот документ описывает настройку FTP подключения к Shutterstock.

## 🔧 Настройки подключения

### Основные параметры

| Параметр | Значение |
|----------|----------|
| **Протокол** | FTP (с TLS шифрованием) |
| **Хост** | `ftps.shutterstock.com` |
| **Порт** | `21` |
| **Шифрование** | `explicit` (Явный FTPS через TLS) |
| **Пассивный режим** | `true` |
| **Таймаут** | `120` секунд |
| **Проверка сертификатов** | `true` |

### Учетные данные

- **Пользователь**: Ваш email от Shutterstock аккаунта
- **Пароль**: Ваш пароль от Shutterstock аккаунта

## 🛠️ Настройка в приложении

### 1. Создание конфигурации

В разделе "Настройки стоков" создайте новую конфигурацию:

```json
{
  "name": "Shutterstock FTP",
  "type": "ftp",
  "supportedTypes": ["commercial", "editorial"],
  "connection": {
    "host": "ftps.shutterstock.com",
    "port": 21,
    "username": "ваш-email@example.com",
    "password": "ваш-пароль",
    "path": "/",
    "timeout": 120,
    "passive": true,
    "encryption": "explicit",
    "verifyCert": true
  }
}
```

### 2. Проверка настроек

После ввода данных используйте кнопку "Проверить подключение" для валидации настроек.

## 🔍 Решение проблем

### Ошибка "i/o timeout"

**Причина**: Неправильные настройки хоста/порта или проблемы с сетью

**Решение**:
1. Проверьте что используется `ftps.shutterstock.com` (НЕ IP адрес)
2. Убедитесь что порт `21`
3. Увеличьте таймаут до 180 секунд если проблемы остаются

### Ошибка авторизации

**Причина**: Неверные учетные данные

**Решение**:
1. Проверьте email и пароль от Shutterstock
2. Убедитесь что аккаунт активен и имеет права на FTP
3. Попробуйте войти через веб-интерфейс Shutterstock

### Проблемы с TLS

**Причина**: Проблемы с SSL/TLS сертификатами

**Решение**:
1. Убедитесь что `encryption: "explicit"`
2. Если проблемы остаются, попробуйте `verifyCert: false`

## 📁 Структура файлов

Shutterstock обычно ожидает следующую структуру:

```
/
├── editorial/          # Редакционные фото
│   ├── photo1.jpg
│   └── photo2.jpg
└── commercial/         # Коммерческие фото
    ├── photo3.jpg
    └── photo4.jpg
```

## ⚙️ Дополнительные настройки

### Промпты для AI

**Commercial**:
```
Create an engaging commercial title and keywords for this stock photo that would appeal to businesses and advertisers. Focus on marketing potential and commercial applications.
```

**Editorial**:
```
Create a descriptive editorial title and keywords for this news/documentary photo focusing on factual content. Include relevant news context and journalistic value.
```

### Параметры качества

| Параметр | Рекомендация |
|----------|--------------|
| **Максимум ключевых слов** | 50 |
| **Минимум ключевых слов** | 5 |
| **Автокатегоризация** | включена |
| **Проверка качества** | включена |

## 🚨 Важные моменты

1. **Безопасность**: Всегда используйте FTPS (шифрование), не обычный FTP
2. **Таймауты**: Shutterstock может быть медленным, устанавливайте достаточный таймаут
3. **Retry логика**: Приложение автоматически делает 3 попытки при сбоях
4. **Лимиты**: Следите за лимитами загрузки Shutterstock

## 📞 Поддержка

При проблемах с FTP доступом обращайтесь в поддержку Shutterstock:
- Email: support@shutterstock.com
- Документация: https://developers.shutterstock.com/ 