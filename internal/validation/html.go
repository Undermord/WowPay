package validation

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// MaxPrice - максимальная разумная цена для товара (100,000 руб)
	MaxPrice = 100000.0
	// MinPrice - минимальная цена для товара
	MinPrice = 0.0
)

// Telegram поддерживает ограниченный набор HTML-тегов
var allowedTags = map[string]bool{
	"b":      true,
	"i":      true,
	"u":      true,
	"s":      true,
	"code":   true,
	"pre":    true,
	"a":      true,
	"strong": true,
	"em":     true,
}

// ValidateHTML проверяет HTML на корректность для Telegram
func ValidateHTML(html string) error {
	// Проверяем баланс тегов
	if err := checkTagBalance(html); err != nil {
		return err
	}

	// Проверяем разрешенные теги
	if err := checkAllowedTags(html); err != nil {
		return err
	}

	// Проверяем атрибуты ссылок
	if err := checkLinkAttributes(html); err != nil {
		return err
	}

	return nil
}

// checkTagBalance проверяет, что все теги правильно закрыты
func checkTagBalance(html string) error {
	tagPattern := regexp.MustCompile(`</?([a-zA-Z]+)[^>]*>`)
	matches := tagPattern.FindAllStringSubmatch(html, -1)

	stack := []string{}

	for _, match := range matches {
		fullTag := match[0]
		tagName := strings.ToLower(match[1])

		// Пропускаем самозакрывающиеся теги (хотя в Telegram их нет)
		if strings.HasSuffix(fullTag, "/>") {
			continue
		}

		// Закрывающий тег
		if strings.HasPrefix(fullTag, "</") {
			if len(stack) == 0 {
				return fmt.Errorf("закрывающий тег <%s> без открывающего", tagName)
			}
			lastTag := stack[len(stack)-1]
			if lastTag != tagName {
				return fmt.Errorf("неправильное закрытие тега: ожидается </%s>, получено </%s>", lastTag, tagName)
			}
			stack = stack[:len(stack)-1]
		} else {
			// Открывающий тег
			stack = append(stack, tagName)
		}
	}

	if len(stack) > 0 {
		return fmt.Errorf("незакрытые теги: %v", stack)
	}

	return nil
}

// checkAllowedTags проверяет, что используются только разрешенные теги
func checkAllowedTags(html string) error {
	tagPattern := regexp.MustCompile(`</?([a-zA-Z]+)[^>]*>`)
	matches := tagPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		tagName := strings.ToLower(match[1])
		if !allowedTags[tagName] {
			return fmt.Errorf("недопустимый тег: <%s>. Разрешены только: b, i, u, s, code, pre, a, strong, em", tagName)
		}
	}

	return nil
}

// checkLinkAttributes проверяет атрибуты ссылок
func checkLinkAttributes(html string) error {
	// Ищем все теги <a> (с атрибутами или без)
	linkPattern := regexp.MustCompile(`<a\s*([^>]*)>`)
	matches := linkPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		attributes := strings.TrimSpace(match[1])

		// Если атрибуты пустые, значит href отсутствует
		if attributes == "" {
			return fmt.Errorf("тег <a> должен содержать атрибут href")
		}

		// Проверяем наличие атрибута href
		hrefPattern := regexp.MustCompile(`href\s*=\s*["']([^"']*)["']`)
		hrefMatch := hrefPattern.FindStringSubmatch(attributes)

		if hrefMatch == nil {
			return fmt.Errorf("тег <a> должен содержать атрибут href")
		}

		href := hrefMatch[1]

		// Проверяем, что href не пустой
		if strings.TrimSpace(href) == "" {
			return fmt.Errorf("атрибут href не может быть пустым")
		}

		// Проверяем, что href начинается с http:// или https:// или mailto: или tg://
		validPrefixes := []string{"http://", "https://", "mailto:", "tg://"}
		isValid := false
		for _, prefix := range validPrefixes {
			if strings.HasPrefix(href, prefix) {
				isValid = true
				break
			}
		}

		if !isValid {
			return fmt.Errorf("некорректный URL в href: %s. URL должен начинаться с http://, https://, mailto: или tg://", href)
		}
	}

	return nil
}

// SanitizeHTML удаляет опасные или недопустимые теги (базовая очистка)
func SanitizeHTML(html string) string {
	// Удаляем script и style теги
	html = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`).ReplaceAllString(html, "")

	// Удаляем комментарии
	html = regexp.MustCompile(`<!--.*?-->`).ReplaceAllString(html, "")

	return html
}

// ValidatePrice проверяет корректность цены товара
func ValidatePrice(price float64) error {
	if price < MinPrice {
		return fmt.Errorf("цена не может быть отрицательной")
	}

	if price > MaxPrice {
		return fmt.Errorf("цена слишком велика (максимум %.2f руб). Возможно, вы допустили опечатку?", MaxPrice)
	}

	return nil
}
