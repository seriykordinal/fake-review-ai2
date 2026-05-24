package models

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type VerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateRoleRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type AnalyzeRequest struct {
	Text string `json:"text"`
}

type AnalyzeResponse struct {
	FakeProbability float64 `json:"fake_probability"`
}

type ProductAnalysisRequest struct {
	ProductURL string `json:"product_url"`
}

type ProductAnalysisResponse struct {
	ProductID              int      `json:"product_id"`
	TotalReviews           int      `json:"total_reviews"`
	AverageFakeProbability float64  `json:"average_fake_probability"`
	Results                []Review `json:"results"`
}
