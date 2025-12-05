package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	// MaxLogSize - максимальный размер лог-файла (10 MB)
	MaxLogSize = 10 * 1024 * 1024
	// MaxLogFiles - максимальное количество лог-файлов
	MaxLogFiles = 10
	// LogDir - директория для логов
	LogDir = "/app/logs"
)

var (
	logFile *os.File
)

// Init инициализирует логирование в файл + stdout
func Init() error {
	// Создаём директорию если не существует
	if err := os.MkdirAll(LogDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Проверяем нужна ли ротация
	if err := rotateIfNeeded(); err != nil {
		return err
	}

	// Открываем файл для записи
	logPath := filepath.Join(LogDir, "bot.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	logFile = file

	// Логи пишутся и в файл, и в stdout (для docker logs)
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return nil
}

// Close закрывает лог-файл
func Close() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

// rotateIfNeeded проверяет размер лог-файла и ротирует если нужно
func rotateIfNeeded() error {
	logPath := filepath.Join(LogDir, "bot.log")

	// Проверяем существует ли файл
	info, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		return nil // Файл не существует, ротация не нужна
	}
	if err != nil {
		return err
	}

	// Если файл меньше лимита, ротация не нужна
	if info.Size() < MaxLogSize {
		return nil
	}

	// Ротируем логи
	return rotateLogs()
}

// rotateLogs переименовывает старые логи и удаляет самые старые
func rotateLogs() error {
	// Закрываем текущий файл если открыт
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}

	// Удаляем самый старый лог (bot.log.10)
	oldestLog := filepath.Join(LogDir, fmt.Sprintf("bot.log.%d", MaxLogFiles))
	os.Remove(oldestLog) // Игнорируем ошибку если файла нет

	// Переименовываем логи: bot.log.1 -> bot.log.2, bot.log.2 -> bot.log.3, и т.д.
	for i := MaxLogFiles - 1; i >= 1; i-- {
		oldName := filepath.Join(LogDir, fmt.Sprintf("bot.log.%d", i))
		newName := filepath.Join(LogDir, fmt.Sprintf("bot.log.%d", i+1))

		// Переименовываем если файл существует
		if _, err := os.Stat(oldName); err == nil {
			os.Rename(oldName, newName)
		}
	}

	// Переименовываем текущий bot.log -> bot.log.1
	currentLog := filepath.Join(LogDir, "bot.log")
	firstRotated := filepath.Join(LogDir, "bot.log.1")
	if err := os.Rename(currentLog, firstRotated); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to rotate log: %w", err)
	}

	return nil
}

// CheckAndRotate проверяет размер лог-файла при запуске бота
// Фоновая ротация отключена для экономии ресурсов сервера
// Используйте logrotate для автоматической ротации в production
func CheckAndRotate() error {
	return rotateIfNeeded()
}
