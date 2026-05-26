package models

import (
	"encoding/json"
	"testing"
)

// ── User — JSON сериализация ────────────────────────────────────────────────

func TestUser_PasswordHashHiddenInJSON(t *testing.T) {
	u := User{
		ID:           1,
		Email:        "test@test.com",
		PasswordHash: "secret-hash",
		Role:         "user",
		IsVerified:   true,
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	json.Unmarshal(data, &m)

	if _, exists := m["password_hash"]; exists {
		t.Error("password_hash should not be in JSON (json:\"-\")")
	}
	if _, exists := m["verification_code"]; exists {
		t.Error("verification_code should not be in JSON")
	}
}

func TestUser_EmailInJSON(t *testing.T) {
	u := User{ID: 1, Email: "user@mail.com", Role: "admin"}
	data, _ := json.Marshal(u)

	var m map[string]any
	json.Unmarshal(data, &m)

	if m["email"] != "user@mail.com" {
		t.Errorf("email = %q, want user@mail.com", m["email"])
	}
	if m["role"] != "admin" {
		t.Errorf("role = %q, want admin", m["role"])
	}
}

// ── Review — JSON ────────────────────────────────────────────────────────────

func TestReview_JSONRoundTrip(t *testing.T) {
	r := Review{
		Rating:          "5",
		Pros:            "Хорошее качество",
		Cons:            "Нет",
		Text:            "Отличный товар",
		Date:            "2024-01-15",
		FakeProbability: 0.85,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Review
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Rating != r.Rating {
		t.Errorf("Rating = %q, want %q", decoded.Rating, r.Rating)
	}
	if decoded.FakeProbability != r.FakeProbability {
		t.Errorf("FakeProbability = %f, want %f", decoded.FakeProbability, r.FakeProbability)
	}
	if decoded.Text != r.Text {
		t.Errorf("Text = %q, want %q", decoded.Text, r.Text)
	}
}

// ── Requests — JSON десериализация ───────────────────────────────────────────

func TestRegisterRequest_Decode(t *testing.T) {
	input := `{"email":"user@test.com","password":"pass123"}`
	var req RegisterRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatal(err)
	}
	if req.Email != "user@test.com" {
		t.Errorf("Email = %q", req.Email)
	}
	if req.Password != "pass123" {
		t.Errorf("Password = %q", req.Password)
	}
}

func TestAnalyzeRequest_Decode(t *testing.T) {
	input := `{"text":"Отличный товар!"}`
	var req AnalyzeRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatal(err)
	}
	if req.Text != "Отличный товар!" {
		t.Errorf("Text = %q", req.Text)
	}
}

func TestProductAnalysisRequest_Decode(t *testing.T) {
	input := `{"product_url":"https://www.wildberries.ru/catalog/12345/detail.aspx"}`
	var req ProductAnalysisRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatal(err)
	}
	if req.ProductURL != "https://www.wildberries.ru/catalog/12345/detail.aspx" {
		t.Errorf("ProductURL = %q", req.ProductURL)
	}
}

func TestUpdateRoleRequest_Decode(t *testing.T) {
	input := `{"email":"admin@test.com","role":"admin"}`
	var req UpdateRoleRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatal(err)
	}
	if req.Email != "admin@test.com" {
		t.Errorf("Email = %q", req.Email)
	}
	if req.Role != "admin" {
		t.Errorf("Role = %q", req.Role)
	}
}

// ── AnalyzeResponse ──────────────────────────────────────────────────────────

func TestAnalyzeResponse_JSON(t *testing.T) {
	resp := AnalyzeResponse{FakeProbability: 0.92}
	data, _ := json.Marshal(resp)

	var m map[string]float64
	json.Unmarshal(data, &m)

	if m["fake_probability"] != 0.92 {
		t.Errorf("fake_probability = %f, want 0.92", m["fake_probability"])
	}
}

// ── ProductAnalysisResponse ─────────────────────────────────────────────────

func TestProductAnalysisResponse_JSON(t *testing.T) {
	resp := ProductAnalysisResponse{
		ProductID:              12345,
		TotalReviews:           10,
		AverageFakeProbability: 0.65,
		Results:                []Review{{Text: "test", FakeProbability: 0.5}},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var decoded ProductAnalysisResponse
	json.Unmarshal(data, &decoded)

	if decoded.ProductID != 12345 {
		t.Errorf("ProductID = %d", decoded.ProductID)
	}
	if decoded.TotalReviews != 10 {
		t.Errorf("TotalReviews = %d", decoded.TotalReviews)
	}
	if len(decoded.Results) != 1 {
		t.Errorf("len(Results) = %d", len(decoded.Results))
	}
}

// ── GlobalStats ──────────────────────────────────────────────────────────────

func TestGlobalStats_JSON(t *testing.T) {
	stats := GlobalStats{
		TotalUsers:           100,
		TotalAnalyses:        50,
		TotalReviewsAnalyzed: 5000,
		GlobalAvgFake:        0.42,
	}
	data, _ := json.Marshal(stats)

	var m map[string]any
	json.Unmarshal(data, &m)

	if m["total_users"].(float64) != 100 {
		t.Error("total_users mismatch")
	}
	if m["total_analyses"].(float64) != 50 {
		t.Error("total_analyses mismatch")
	}
}
