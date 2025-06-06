// Internationalization (i18n) manager
class I18nManager {
    constructor() {
        this.currentLanguage = 'en';
        this.translations = {};
        this.observers = [];
    }

    async init() {
        // Загружаем сохраненный язык из localStorage
        const savedLanguage = localStorage.getItem('language') || 'en';
        await this.loadLanguage(savedLanguage);
    }

    async loadLanguage(language) {
        try {
            const response = await fetch(`./src/locales/${language}.json`);
            const translations = await response.json();
            
            this.translations[language] = translations;
            this.currentLanguage = language;
            
            // Сохраняем выбранный язык
            localStorage.setItem('language', language);
            
            // Уведомляем наблюдателей об изменении языка
            this.notifyObservers();
            
            // Обновляем интерфейс
            this.updateInterface();
            
            return translations;
        } catch (error) {
            console.error(`Failed to load language ${language}:`, error);
            
            // Fallback к английскому если не удалось загрузить
            if (language !== 'en') {
                return this.loadLanguage('en');
            }
            throw error;
        }
    }

    t(key, params = {}) {
        const keys = key.split('.');
        let value = this.translations[this.currentLanguage];
        
        for (const k of keys) {
            if (value && typeof value === 'object') {
                value = value[k];
            } else {
                // Fallback к английскому если ключ не найден
                value = this.getFallbackTranslation(key);
                break;
            }
        }
        
        if (typeof value !== 'string') {
            console.warn(`Translation key not found: ${key}`);
            return key;
        }
        
        // Подстановка параметров
        return this.interpolate(value, params);
    }

    getFallbackTranslation(key) {
        const keys = key.split('.');
        let value = this.translations['en'];
        
        for (const k of keys) {
            if (value && typeof value === 'object') {
                value = value[k];
            } else {
                return key; // Возвращаем сам ключ если перевод не найден
            }
        }
        
        return typeof value === 'string' ? value : key;
    }

    interpolate(template, params) {
        return template.replace(/\{\{(\w+)\}\}/g, (match, key) => {
            return params[key] !== undefined ? params[key] : match;
        });
    }

    getCurrentLanguage() {
        return this.currentLanguage;
    }

    getAvailableLanguages() {
        return ['en', 'ru'];
    }

    subscribe(callback) {
        this.observers.push(callback);
    }

    unsubscribe(callback) {
        this.observers = this.observers.filter(obs => obs !== callback);
    }

    notifyObservers() {
        this.observers.forEach(callback => callback(this.currentLanguage));
    }

    updateInterface() {
        // Обновляем заголовок документа
        document.title = this.t('app.title');
        
        // Обновляем все элементы с data-i18n атрибутами
        document.querySelectorAll('[data-i18n]').forEach(element => {
            const key = element.getAttribute('data-i18n');
            const translation = this.t(key);
            
            if (element.tagName === 'INPUT' && (element.type === 'text' || element.type === 'number')) {
                element.placeholder = translation;
            } else if (element.tagName === 'TEXTAREA') {
                element.placeholder = translation;
            } else {
                element.textContent = translation;
            }
        });

        // Обновляем элементы с data-i18n-placeholder атрибутами
        document.querySelectorAll('[data-i18n-placeholder]').forEach(element => {
            const key = element.getAttribute('data-i18n-placeholder');
            element.placeholder = this.t(key);
        });

        // Обновляем элементы с data-i18n-title атрибутами
        document.querySelectorAll('[data-i18n-title]').forEach(element => {
            const key = element.getAttribute('data-i18n-title');
            element.title = this.t(key);
        });
    }

    async switchLanguage(language) {
        if (language !== this.currentLanguage) {
            await this.loadLanguage(language);
        }
    }
}

// Создаем глобальный экземпляр
window.i18n = new I18nManager();

export default window.i18n; 