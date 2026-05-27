package services

import (
	"testing"
)

func TestExtractProductID_ValidURLs(t *testing.T) {
	svc := NewWBParserService()

	tests := []struct {
		url      string
		expected int
	}{
		{"https://www.wildberries.ru/catalog/12345678/detail.aspx", 12345678},
		{"https://wildberries.ru/catalog/99999/detail.aspx", 99999},
		{"https://www.wildberries.ru/catalog/1/detail.aspx", 1},
		{"https://www.wildberries.ru/catalog/100200300/detail.aspx?targetUrl=GP", 100200300},
		{"https://www.wildberries.ru/catalog/55555/feedbacks", 55555},
	}
	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			id, err := svc.ExtractProductID(tc.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tc.expected {
				t.Errorf("got %d, want %d", id, tc.expected)
			}
		})
	}
}

func TestExtractProductID_InvalidURLs(t *testing.T) {
	svc := NewWBParserService()

	tests := []string{
		"https://www.wildberries.ru/",
		"https://www.wildberries.ru/catalog/",
		"https://www.wildberries.ru/catalog/abc/detail.aspx",
		"https://ozon.ru/product/12345",
		"not-a-url",
		"",
		"https://www.wildberries.ru/catalogdetail.aspx",
	}
	for _, url := range tests {
		t.Run(url, func(t *testing.T) {
			_, err := svc.ExtractProductID(url)
			if err == nil {
				t.Errorf("expected error for URL %q", url)
			}
		})
	}
}
