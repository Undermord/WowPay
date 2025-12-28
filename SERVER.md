# Документация по серверу

## Данные доступа к серверу

**IP адрес**: 185.133.40.93
**Логин**: root
**Пароль**: YK4vjI7

**Характеристики**:
- CPU: 2 ядра x 2600 MHz
- RAM: 4 Гб
- Диск: 40 Гб SSD NVMe
- Стоимость: 250 рублей/месяц
- Установлено: Docker + Portainer

## Подключение к серверу

```bash
# SSH подключение
ssh root@185.133.40.93

# Или с использованием пароля напрямую (для автоматизации)
sshpass -p 'YK4vjI7' ssh root@185.133.40.93
```

## Структура проекта на сервере

```
/root/tgWow/              # Основная директория проекта
├── .env                  # Переменные окружения (секреты)
├── docker-compose.yml    # Конфигурация Docker
├── Dockerfile            # Сборка бота
├── cmd/                  # Исходный код бота
├── internal/             # Внутренние пакеты
├── migrations/           # SQL миграции БД
└── logs/                 # Логи приложения
```

## Управление Docker контейнерами

### Базовые команды

```bash
# Перейти в директорию проекта
cd /root/tgWow

# Посмотреть статус контейнеров
docker compose ps

# Посмотреть логи всех сервисов
docker compose logs

# Посмотреть логи бота в реальном времени
docker compose logs -f bot

# Посмотреть последние 50 строк логов
docker compose logs --tail=50 bot
```

### Запуск и остановка

```bash
# Запустить все сервисы
docker compose up -d

# Остановить все сервисы
docker compose down

# Остановить и удалить все данные (включая БД!)
docker compose down -v

# Перезапустить только бота
docker compose restart bot

# Перезапустить все сервисы
docker compose restart
```

### Обновление кода

```bash
# Обновить код из Git репозитория
cd /root/tgWow
git pull

# Пересобрать и перезапустить бота
docker compose up -d --build bot

# Или полностью пересобрать всё
docker compose down
docker compose up -d --build
```

### Сборка образов

```bash
# Пересобрать образ бота
docker compose build bot

# Пересобрать с очисткой кэша
docker compose build --no-cache bot
```

## Доступ к сервисам

### Telegram-бот
- **Username**: @wowpay_app_bot
- **Ссылка**: https://t.me/wowpay_app_bot
- **Админы** (получают уведомления о заказах):
  - User ID: 451978442
  - User ID: 778736273
  - User ID: 486518615
  - User ID: 766427421

### Adminer (веб-интерфейс для PostgreSQL)
- **URL**: http://185.133.40.93:8080
- **Система**: PostgreSQL
- **Сервер**: db
- **Пользователь**: wowbot
- **Пароль**: parol567nasvai
- **База данных**: wowbot

### PostgreSQL (прямое подключение)

```bash
# Через Docker
docker compose exec db psql -U wowbot -d wowbot

# С хоста (если установлен psql)
psql -h 185.133.40.93 -U wowbot -d wowbot
# Пароль: parol567nasvai

# Строка подключения
postgres://wowbot:parol567nasvai@185.133.40.93:5432/wowbot?sslmode=disable
```

## Переменные окружения (.env)

```bash
# Просмотр .env файла
cat /root/tgWow/.env

# Редактирование .env файла
nano /root/tgWow/.env

# После изменений перезапустить бота
docker compose restart bot
```

**Текущие переменные**:
- `BOT_TOKEN`: 8489950679:AAFaa6-LSAmG3wRjI6ljJx--I4Bw1yBN3HU
- `ADMIN_CHAT_ID`: 451978442,778736273,486518615,766427421
- `POSTGRES_PASSWORD`: parol567nasvai
- `DATABASE_URL`: postgres://wowbot:parol567nasvai@db:5432/wowbot?sslmode=disable
- `PAYMENT_CARD_NUMBER`: 2200700977297505

## Полезные SQL запросы

```bash
# Войти в PostgreSQL
docker compose exec db psql -U wowbot -d wowbot
```

```sql
-- Список всех таблиц
\dt

-- Количество товаров
SELECT count(*) FROM products;

-- Список всех регионов
SELECT * FROM regions;

-- Последние 10 заказов
SELECT o.order_id, o.user_id, p.name, o.price, o.status, o.created_at
FROM orders o
JOIN products p ON o.product_id = p.id
ORDER BY o.created_at DESC
LIMIT 10;

-- Статистика по статусам заказов
SELECT status, count(*) as count, sum(price) as total
FROM orders
GROUP BY status;

-- Сделать все товары видимыми
UPDATE products SET is_visible = true;

-- Изменить цену товара
UPDATE products SET price = 999.00 WHERE id = 1;

-- Выход
\q
```

## Мониторинг и отладка

### Просмотр логов

```bash
# Логи бота
docker compose logs -f bot

# Логи базы данных
docker compose logs -f db

# Логи с временными метками
docker compose logs -f -t bot

# Поиск ошибок в логах
docker compose logs bot | grep -i error
```

### Использование ресурсов

```bash
# Статистика контейнеров (CPU, RAM, Network)
docker stats

# Использование диска
df -h

# Размер Docker образов
docker images

# Очистка неиспользуемых ресурсов
docker system prune -a
```

### Проверка работоспособности

```bash
# Проверить, работает ли PostgreSQL
docker compose exec db pg_isready -U wowbot

# Проверить подключение к Telegram API
docker compose exec bot wget -O- https://api.telegram.org/bot8489950679:AAFaa6-LSAmG3wRjI6ljJx--I4Bw1yBN3HU/getMe
```

## Резервное копирование

### Бэкап базы данных

```bash
# Создать бэкап
docker compose exec db pg_dump -U wowbot wowbot > backup_$(date +%Y%m%d_%H%M%S).sql

# Или с сжатием
docker compose exec db pg_dump -U wowbot wowbot | gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz
```

### Восстановление из бэкапа

```bash
# Восстановить из бэкапа
cat backup_20251211_103000.sql | docker compose exec -T db psql -U wowbot -d wowbot

# Или из сжатого
gunzip -c backup_20251211_103000.sql.gz | docker compose exec -T db psql -U wowbot -d wowbot
```

### Бэкап .env и конфигурации

```bash
# Скопировать важные файлы с сервера на локальную машину
scp root@185.133.40.93:/root/tgWow/.env ./backup/.env
scp root@185.133.40.93:/root/tgWow/docker-compose.yml ./backup/docker-compose.yml
```

## Безопасность

### Обновление секретов

```bash
# Изменить пароль БД (требует пересоздания контейнера)
# 1. Остановить контейнеры
docker compose down -v

# 2. Отредактировать .env (изменить POSTGRES_PASSWORD и DATABASE_URL)
nano .env

# 3. Запустить снова
docker compose up -d
```

### Firewall (если требуется)

```bash
# Открыть только необходимые порты
ufw allow 22/tcp    # SSH
ufw allow 8080/tcp  # Adminer (если нужен извне)
ufw enable
```

## Мониторинг через Portainer

Если установлен Portainer:
- **URL**: http://185.133.40.93:9000 (или другой порт)
- Позволяет управлять контейнерами через веб-интерфейс

## Git репозиторий

- **URL**: https://github.com/Undermord/WowPay.git
- **Ветка**: main

```bash
# Посмотреть статус Git
cd /root/tgWow
git status

# Посмотреть последние коммиты
git log --oneline -10

# Откатить изменения
git reset --hard origin/main

# Обновить код
git pull
```

## Troubleshooting

### Бот не отвечает

```bash
# Проверить статус
docker compose ps

# Проверить логи на ошибки
docker compose logs bot | tail -100

# Перезапустить бота
docker compose restart bot
```

### База данных не доступна

```bash
# Проверить статус БД
docker compose ps db

# Проверить healthcheck
docker compose exec db pg_isready -U wowbot

# Перезапустить БД
docker compose restart db
```

### Нет места на диске

```bash
# Проверить использование диска
df -h

# Очистить логи Docker
docker system prune -a

# Очистить старые образы
docker image prune -a
```

### Adminer не подключается к БД

Убедитесь что:
1. Пароль в Adminer: `parol567nasvai`
2. Сервер: `db` (не localhost!)
3. Пользователь: `wowbot`
4. База данных: `wowbot`

```bash
# Проверить, что пароль совпадает
cat .env | grep POSTGRES_PASSWORD
cat docker-compose.yml | grep POSTGRES_PASSWORD

# Перезапустить оба сервиса
docker compose restart db adminer
```

## Контакты и поддержка

При возникновении проблем:
1. Проверьте логи: `docker compose logs bot`
2. Проверьте статус: `docker compose ps`
3. Перезапустите сервисы: `docker compose restart`
4. Если проблема не решается - обратитесь к разработчику

---

**Последнее обновление**: 2025-12-11
**Версия бота**: @wowpay_app_bot
