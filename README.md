# Tinkoff Invest Telegram Bot

Телеграм-бот для мониторинга сделок на Тинькофф Инвестиции. Бот отслеживает все новые покупки и продажи акций и отправляет уведомления в Telegram-группу.

## Возможности

- ✅ Мониторинг сделок в реальном времени (каждые 30 секунд)
- ✅ Отправка уведомлений о новых сделках в Telegram-группу
- ✅ Уведомления по сделкам за текущий день при запуске
- ✅ Фильтрация инвесткопилки (автоматически пропускается)
- ✅ Поддержка sandbox и production режимов
- ✅ Сохранение операций в SQLite базу данных
- ✅ Время в сообщениях по Кемеровскому часовому поясу (UTC+7)

## Требования

- Go 1.21+
- Docker и Docker Compose
- Токен Telegram-бота
- Токен Tinkoff Invest API

## Установка

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd tinvest
```

### 2. Настройка переменных окружения

Скопируйте файл `.env.example` в `.env` и заполните:

```bash
cp .env.example .env
```

Отредактируйте `.env`:

```env
# Токен Telegram бота
TELEGRAM_BOT_TOKEN=your_telegram_bot_token

# ID группы для уведомлений (отрицательное число для групп)
TELEGRAM_GROUP_ID=-123456789

# Токен Tinkoff Invest API
TINKOFF_TOKEN=your_tinkoff_token

# ID счета (опционально, если пусто - используется первый доступный)
TINKOFF_ACCOUNT_ID=

# Путь к базе данных
DB_PATH=/app/data/tinvest.db

# Интервал проверки новых сделок
POLL_INTERVAL=30s

# Режим sandbox (true/false)
USE_SANDBOX=false
```

### 3. Запуск через Docker Compose

```bash
docker compose up -d --build
```

## Структура проекта

```
tinvest/
├── bot/                 # Telegram Bot клиент
├── config/              # Конфигурация приложения
├── notifier/            # Формирование и отправка уведомлений
├── storage/             # Работа с SQLite базой данных
├── tinkoff/             # Tinkoff Invest API клиент
├── watcher/             # Мониторинг новых сделок
├── data/                # Папка для базы данных
├── main.go              # Точка входа
├── docker-compose.yml   # Docker Compose конфигурация
├── Dockerfile           # Docker образ
└── .env                 # Переменные окружения
```

## База данных

База данных SQLite хранится в `data/tinvest.db`.

### Подключение в Goland/DataGrip

- **Файл:** `data/tinvest.db`
- **Тип:** SQLite
- **Credentials:** не требуются

### Структура таблицы operations

| Поле | Тип | Описание |
|------|-----|----------|
| id | TEXT | Уникальный ID операции |
| operation_type | TEXT | Тип операции (BUY/SELL) |
| status | TEXT | Статус операции |
| date | TEXT | Дата операции |
| name | TEXT | Название инструмента |
| ticker | TEXT | Тикер инструмента |
| figi | TEXT | FIGI инструмента |
| quantity | INTEGER | Количество лотов |
| price | REAL | Цена за лот |
| total | REAL | Общая сумма |
| currency | TEXT | Валюта |
| commission | REAL | Комиссия |
| account_id | TEXT | ID счета |
| created_at | TEXT | Дата создания записи |

### Управление базой данных

Просмотр операций:
```bash
sqlite3 data/tinvest.db "SELECT * FROM operations LIMIT 10;"
```

Удаление операции:
```bash
sqlite3 data/tinvest.db "DELETE FROM operations WHERE id='your_operation_id';"
```

## Использование

### Sandbox режим (тестирование)

Для тестирования используйте sandbox:

```env
USE_SANDBOX=true
```

В sandbox режиме бот будет:
1. Создавать тестовый счет
2. Пополнять баланс
3. Размещать тестовые ордера

### Production режим (реальная работа)

Для работы с реальным счетом:

```env
USE_SANDBOX=false
```

В production режиме бот только мониторит сделки без размещения тестовых ордеров.

## Формат уведомлений

Бот отправляет сообщения в формате:

```
📈 Новая сделка

🏷 Тип: Покупка 📈
📊 Статус: ✅ Исполнена
🕐 Время: 14.03.2026 18:30:00

🎫 Тикер: VTBR
💰 Название: Банк ВТБ (ПАО)
📁 FIGI: BBG004730ZJ9

💵 Количество: 1
💎 Цена: 86.39
💵 Сумма: 86.39 rub
```

## Разработка

### Запуск без Docker

```bash
# Установка зависимостей
go mod download

# Сборка
go build -o tinkoff-bot .

# Запуск
./tinkoff-bot
```

### Тесты

```bash
go test -v ./...
```

### Покрытие тестами

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Логирование

Логи доступны через Docker:

```bash
docker logs tinvest-bot
```

## Устранение проблем

### Не приходят уведомления

1. Проверьте токен бота
2. Проверьте ID группы
3. Убедитесь, что бот добавлен в группу
4. Проверьте права бота на отправку сообщений

### Ошибки подключения к Tinkoff API

1. Проверьте токен Tinkoff
2. Убедитесь, что токен не истек
3. Проверьте права токена (должен быть доступ к Operations)

### База данных недоступна для записи

```bash
sudo chown -R $USER:$USER data/
```

## Лицензия

MIT License
