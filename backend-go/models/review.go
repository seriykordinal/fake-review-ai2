package models

import "time"

type Review struct {
	Rating          string  `json:"rating"`
	Pros            string  `json:"pros"`
	Cons            string  `json:"cons"`
	Text            string  `json:"text"`
	Date            string  `json:"date"`
	FakeProbability float64 `json:"fake_probability"`
}

type ProductAnalysisSummary struct {
	AnalysisID   int       `json:"analysis_id"`
	UserID       int       `json:"user_id"`
	UserEmail    string    `json:"user_email"`
	ProductID    int       `json:"product_id"`
	ProductURL   string    `json:"product_url"`
	TotalReviews int       `json:"total_reviews"`
	AnalyzedAt   time.Time `json:"analyzed_at"`
}

type GlobalStats struct {
	TotalUsers           int     `json:"total_users"`
	TotalAnalyses        int     `json:"total_analyses"`
	TotalReviewsAnalyzed int     `json:"total_reviews_analyzed"`
	GlobalAvgFake        float64 `json:"global_avg_fake"`
}
