package validation

import (
	"testing"
)

func TestValidateHTML_ValidHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
	}{
		{
			name: "Simple bold",
			html: "Hello <b>world</b>",
		},
		{
			name: "Nested tags",
			html: "<b>Bold <i>and italic</i></b>",
		},
		{
			name: "Link with href",
			html: `Visit <a href="https://example.com">example</a>`,
		},
		{
			name: "Multiple tags",
			html: `<b>Bold</b> <i>Italic</i> <u>Underline</u> <s>Strike</s> <code>Code</code>`,
		},
		{
			name: "Link with tg://",
			html: `<a href="tg://user?id=123">User</a>`,
		},
		{
			name: "Placeholder",
			html: `Hello {name}!`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHTML(tt.html)
			if err != nil {
				t.Errorf("ValidateHTML() error = %v, want nil", err)
			}
		})
	}
}

func TestValidateHTML_InvalidHTML(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantErr string
	}{
		{
			name:    "Unclosed tag",
			html:    "<b>Hello world",
			wantErr: "незакрытые теги",
		},
		{
			name:    "Wrong closing tag",
			html:    "<b>Hello <i>world</b></i>",
			wantErr: "неправильное закрытие тега",
		},
		{
			name:    "Disallowed tag",
			html:    "<script>alert('xss')</script>",
			wantErr: "недопустимый тег",
		},
		{
			name:    "Link without href",
			html:    "<a>Click here</a>",
			wantErr: "должен содержать атрибут href",
		},
		{
			name:    "Link with empty href",
			html:    `<a href="">Click here</a>`,
			wantErr: "атрибут href не может быть пустым",
		},
		{
			name:    "Link with invalid URL",
			html:    `<a href="javascript:alert('xss')">Click</a>`,
			wantErr: "некорректный URL",
		},
		{
			name:    "Div tag",
			html:    "<div>Hello</div>",
			wantErr: "недопустимый тег",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHTML(tt.html)
			if err == nil {
				t.Errorf("ValidateHTML() error = nil, want error containing %q", tt.wantErr)
				return
			}
			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("ValidateHTML() error = %q, want error containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "Remove script",
			html: "Hello <script>alert('xss')</script> world",
			want: "Hello  world",
		},
		{
			name: "Remove style",
			html: "Hello <style>body{color:red}</style> world",
			want: "Hello  world",
		},
		{
			name: "Remove comments",
			html: "Hello <!-- comment --> world",
			want: "Hello  world",
		},
		{
			name: "Keep allowed tags",
			html: "<b>Hello</b> <i>world</i>",
			want: "<b>Hello</b> <i>world</i>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeHTML(tt.html)
			if got != tt.want {
				t.Errorf("SanitizeHTML() = %q, want %q", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
