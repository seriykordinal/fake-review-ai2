package services

import (
	"bytes"
	"encoding/json"
	"fake-review-ai2/models"
	"fmt"
	"net/http"
)

type AnalysisService struct {
	pythonHost string
	pythonPort int
}

func NewAnalysisService(pythonHost string, pythonPort int) *AnalysisService {
	return &AnalysisService{pythonHost: pythonHost, pythonPort: pythonPort}
}

func (s *AnalysisService) AnalyzeText(text string) (float64, error) {
	pythonURL := fmt.Sprintf("http://%s:%d/predict", s.pythonHost, s.pythonPort)
	reqBody := models.AnalyzeRequest{Text: text}
	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(pythonURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var result models.AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.FakeProbability, nil
}

func (s *AnalysisService) AnalyzeTexts(texts []string) ([]float64, error) {
	pythonURL := fmt.Sprintf("http://%s:%d/predict_batch", s.pythonHost, s.pythonPort)
	reqBody := map[string][]string{"texts": texts}
	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(pythonURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Probabilities []float64 `json:"probabilities"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Probabilities, nil
}
