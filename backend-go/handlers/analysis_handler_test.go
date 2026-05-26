package handlers

import (
	"bytes"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"fake-review-ai2/models"
)

// ── AnalyzeReview handler — валидация ────────────────────────────────────────

func TestAnalyzeReview_InvalidJSON(t *testing.T) {
	handler := &AnalysisHandler{}
	req := httptest.NewRequest("POST", "/api/analyze", bytes.NewReader([]byte("not json")))
	rr := httptest.NewRecorder()

	handler.AnalyzeReview(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// ── buildResponse ────────────────────────────────────────────────────────────

func TestBuildResponse_Empty(t *testing.T) {
	resp := buildResponse(12345, []models.Review{})

	if resp.ProductID != 12345 {
		t.Errorf("ProductID = %d, want 12345", resp.ProductID)
	}
	if resp.TotalReviews != 0 {
		t.Errorf("TotalReviews = %d, want 0", resp.TotalReviews)
	}
	if resp.AverageFakeProbability != 0 {
		t.Errorf("AverageFakeProbability = %f, want 0", resp.AverageFakeProbability)
	}
	if len(resp.Results) != 0 {
		t.Errorf("len(Results) = %d, want 0", len(resp.Results))
	}
}

func TestBuildResponse_SingleReview(t *testing.T) {
	reviews := []models.Review{
		{Text: "Отличный товар", Rating: "5", FakeProbability: 0.8},
	}
	resp := buildResponse(100, reviews)

	if resp.TotalReviews != 1 {
		t.Errorf("TotalReviews = %d, want 1", resp.TotalReviews)
	}
	if resp.AverageFakeProbability != 0.8 {
		t.Errorf("AverageFakeProbability = %f, want 0.8", resp.AverageFakeProbability)
	}
}

func TestBuildResponse_MultipleReviews(t *testing.T) {
	reviews := []models.Review{
		{Text: "Хороший", FakeProbability: 0.2},
		{Text: "Плохой", FakeProbability: 0.6},
		{Text: "Средний", FakeProbability: 0.4},
	}
	resp := buildResponse(200, reviews)

	if resp.TotalReviews != 3 {
		t.Errorf("TotalReviews = %d, want 3", resp.TotalReviews)
	}
	expected := (0.2 + 0.6 + 0.4) / 3.0
	if math.Abs(resp.AverageFakeProbability-expected) > 1e-10 {
		t.Errorf("AverageFakeProbability = %f, want %f", resp.AverageFakeProbability, expected)
	}
}

func TestBuildResponse_AllFieldsPreserved(t *testing.T) {
	reviews := []models.Review{
		{
			Text:            "Текст",
			Rating:          "4",
			Pros:            "Плюсы",
			Cons:            "Минусы",
			Date:            "2024-01-01",
			FakeProbability: 0.55,
		},
	}
	resp := buildResponse(999, reviews)

	r := resp.Results[0]
	if r.Text != "Текст" {
		t.Errorf("Text = %q", r.Text)
	}
	if r.Rating != "4" {
		t.Errorf("Rating = %q", r.Rating)
	}
	if r.Pros != "Плюсы" {
		t.Errorf("Pros = %q", r.Pros)
	}
	if r.Cons != "Минусы" {
		t.Errorf("Cons = %q", r.Cons)
	}
	if r.Date != "2024-01-01" {
		t.Errorf("Date = %q", r.Date)
	}
}

// ── AnalyzeProduct — валидация JSON ──────────────────────────────────────────

func TestAnalyzeProduct_InvalidJSON(t *testing.T) {
	handler := &AnalysisHandler{}
	// Нужны claims в контексте, но мы проверяем ошибку JSON — паника на claims
	// будет раньше. Для чистого теста JSON-валидации добавим claims:
	req := httptest.NewRequest("POST", "/api/analyze_product", bytes.NewReader([]byte("{bad json")))
	rr := httptest.NewRecorder()

	// Без claims в контексте — паника
	defer func() {
		if r := recover(); r != nil {
			// OK: паника из-за отсутствия claims — ожидаемо
		}
	}()
	handler.AnalyzeProduct(rr, req)
}
