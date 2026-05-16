package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"fake-review-ai2/database"
	"fake-review-ai2/middleware"
	"fake-review-ai2/models"
	"fake-review-ai2/services"
)

type AnalysisHandler struct {
	analysisService *services.AnalysisService
	wbParserService *services.WBParserService
}

func NewAnalysisHandler(analysisService *services.AnalysisService, wbParserService *services.WBParserService) *AnalysisHandler {
	return &AnalysisHandler{analysisService: analysisService, wbParserService: wbParserService}
}

func (h *AnalysisHandler) AnalyzeReview(w http.ResponseWriter, r *http.Request) {
	var req models.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	prob, err := h.analysisService.AnalyzeText(req.Text)
	if err != nil {
		writeJSONError(w, "Analysis failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, models.AnalyzeResponse{FakeProbability: prob})
}

func (h *AnalysisHandler) AnalyzeProduct(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*models.Claims)

	var req models.ProductAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	productID, err := h.wbParserService.ExtractProductID(req.ProductURL)
	if err != nil {
		writeJSONError(w, "Invalid product URL", http.StatusBadRequest)
		return
	}

	// Проверяем глобальный кэш
	if cachedResp := h.loadFromCache(productID); cachedResp != nil {
		log.Printf("Returning cached analysis for product %d", productID)
		writeJSON(w, cachedResp)
		return
	}

	// Парсим отзывы
	reviews, err := h.wbParserService.FetchProductReviewsChromedp(productID)
	if err != nil {
		writeJSONError(w, "Failed to parse reviews: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(reviews) == 0 {
		writeJSONError(w, "No reviews found", http.StatusNotFound)
		return
	}

	// Batch-анализ
	texts := make([]string, len(reviews))
	for i, rev := range reviews {
		texts[i] = rev.Text
	}
	probs, err := h.analysisService.AnalyzeTexts(texts)
	if err != nil {
		log.Printf("Batch analysis failed: %v", err)
		writeJSONError(w, "Analysis service error", http.StatusInternalServerError)
		return
	}

	// Проставляем вероятности
	for i := range reviews {
		if i < len(probs) {
			reviews[i].FakeProbability = probs[i]
		}
	}

	// Сохраняем в БД
	analysisID, err := database.SaveProductAnalysis(claims.UserID, req.ProductURL, productID, len(reviews))
	if err != nil {
		log.Printf("Warning: failed to save product analysis: %v", err)
	} else {
		for _, rev := range reviews {
			if err := database.SaveReviewAnalysis(analysisID, rev, rev.FakeProbability); err != nil {
				log.Printf("Warning: failed to save review: %v", err)
			}
		}
	}

	writeJSON(w, buildResponse(productID, reviews))
}

func (h *AnalysisHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*models.Claims)
	analyses, err := database.GetUserProductAnalyses(claims.UserID)
	if err != nil {
		writeJSONError(w, "Failed to get history: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if analyses == nil {
		analyses = []database.ProductAnalysisSummary{}
	}
	writeJSON(w, analyses)
}

func (h *AnalysisHandler) loadFromCache(productID int) *models.ProductAnalysisResponse {
	_, totalReviews, cachedReviews, err := database.GetExistingProductAnalysisGlobal(productID)
	if err != nil || len(cachedReviews) == 0 {
		return nil
	}
	resp := buildResponse(productID, cachedReviews)
	resp.TotalReviews = totalReviews
	return resp
}

func buildResponse(productID int, reviews []models.Review) *models.ProductAnalysisResponse {
	sum := 0.0
	for _, r := range reviews {
		sum += r.FakeProbability
	}
	avg := 0.0
	if len(reviews) > 0 {
		avg = sum / float64(len(reviews))
	}
	return &models.ProductAnalysisResponse{
		ProductID:              productID,
		TotalReviews:           len(reviews),
		AverageFakeProbability: avg,
		Results:                reviews,
	}
}
