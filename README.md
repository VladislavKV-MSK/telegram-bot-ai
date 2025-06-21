# Telegram-бот с интеграцией DeepSeek AI

<img src="https://img.shields.io/badge/Docker-✓-blue?logo=docker" alt="Docker"> <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go" alt="Go"> <img src="https://img.shields.io/badge/OpenRouter-API-7F5AB6" alt="OpenRouter">

Умный Telegram-бот с искусственным интеллектом на базе DeepSeek v3 через OpenRouter API, поддерживающий контекстные диалоги для разных пользовталей в одном групповом чате.

## 🌟 Основные возможности

- **Ответы на основе ИИ** - Использует модель DeepSeek v3 через OpenRouter
- **Контекст диалога** - Помнит последние 10 сообщений каждого пользователя
- **Управление для админов** - Только администраторы могут активировать/деактивировать бота
- **Работа в чатах** - Отвечает на упоминания (@username_bot) в группах
- **Простое развертывание** - Готовые Docker-образы

## 🚀 Быстрый старт

### Необходимые компоненты
- Установленный Docker
- Токен бота от [@BotFather](https://t.me/BotFather)
- API-ключ от [openrouter.ai](https://openrouter.ai)

### 1. Клонируем репозиторий
```bash
git clone https://github.com/yourusername/orchestrator-ai-bot.git
cd orchestrator-ai-bot
```
### 2. Настраиваем окружение
Создаем файл .env:
```bash
cp .env.example .env
```
Редактируем .env:
```bash
TELEGRAM_BOT_TOKEN=ваш_токен_бота
OPENROUTER_API_KEY=ваш_ключ_openrouter
ADMIN_IDS=123456789,987654321  # ID администраторов через запятую
```
### 3. Запускаем бота
```bash
docker build -t orchestrator-bot .
docker run --env-file .env -it orchestrator-bot
```
## ⚙️ Команды бота
**/start** - Активация бота (только для админов)

**/stop** - Деактивация бота (только для админов)

В группах бот отвечает только на сообщения с упоминанием (@username_bot).
## 🔧 Технические детали
- Язык: Go 1.24+
- Хранение контекста: В памяти (до 10 сообщений на пользователя)
- Логирование: В консоль (можно перенаправить в файл)
## 🤝 Участие в разработке
1. Форкните репозиторий
2. Создайте ветку для своей фичи (git checkout -b feature/amazing-feature)
3. Зафиксируйте изменения (git commit -m 'Add some amazing feature')
4. Запушьте в форк (git push origin feature/amazing-feature)
5. Откройте Pull Request

## 📝 Лицензия
MIT License. Подробнее в файле LICENSE.
