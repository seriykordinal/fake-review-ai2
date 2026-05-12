package models

type AnalyzeRequest struct {
	Text string `json:"text"`
}

type AnalyzeResponse struct {
	FakeProbability float64 `json:"fake_probability"`
}

type ProductAnalysisRequest struct {
	ProductURL string `json:"product_url"`
}

type Review struct {
	Rating          string  `json:"rating"`
	Pros            string  `json:"pros"`
	Cons            string  `json:"cons"`
	Text            string  `json:"text"`
	Date            string  `json:"date"`
	FakeProbability float64 `json:"fake_probability"`
}

type ProductAnalysisResponse struct {
	ProductID              int      `json:"product_id"`
	TotalReviews           int      `json:"total_reviews"`
	AverageFakeProbability float64  `json:"average_fake_probability"`
	Results                []Review `json:"results"`
}
