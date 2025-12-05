#!/bin/bash

# Скрипт для настройки logrotate на сервере
# Использование: sudo bash setup-logrotate.sh

set -e

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
LOGS_DIR="$PROJECT_DIR/logs"

echo "🔧 Настройка logrotate для WoW Bot..."

# Создаём конфиг logrotate
cat > /etc/logrotate.d/wowbot << EOF
# Auto-generated logrotate config for WoW Bot
$LOGS_DIR/bot.log {
    # Ротация по размеру
    size 10M

    # Хранить 10 файлов
    rotate 10

    # Сжимать старые логи
    compress
    delaycompress

    # Создавать новый файл
    create 0644 $(whoami) $(whoami)

    # Не ругаться если файла нет
    missingok

    # Не ротировать пустые файлы
    notifempty

    # Проверка раз в день
    daily
}
EOF

echo "✅ Конфиг создан: /etc/logrotate.d/wowbot"

# Проверяем конфигурацию
echo "🧪 Тестируем конфигурацию..."
logrotate -d /etc/logrotate.d/wowbot

echo "✅ Logrotate настроен успешно!"
echo ""
echo "📋 Команды для управления:"
echo "  - Тест ротации:  sudo logrotate -d /etc/logrotate.d/wowbot"
echo "  - Ручная ротация: sudo logrotate -f /etc/logrotate.d/wowbot"
echo "  - Просмотр логов: tail -f $LOGS_DIR/bot.log"
