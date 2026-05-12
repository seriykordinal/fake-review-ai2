package database

import (
	"database/sql"
	"fake-review-ai2/models"
	"time"
)

func SaveProductAnalysis(userID int, productURL string, productID int, totalReviews int) (int, error) {
	var analysisID int
	err := DB.QueryRow(`
        INSERT INTO product_analysis (user_id, product_url, product_id, total_reviews)
        VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, productURL, productID, totalReviews).Scan(&analysisID)
	return analysisID, err
}

func SaveReviewAnalysis(analysisID int, review models.Review, fakeProbability float64) error {
	_, err := DB.Exec(`
        INSERT INTO review_analysis (analysis_id, review_text, review_rating, review_pros, review_cons, review_date, fake_probability)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		analysisID, review.Text, review.Rating, review.Pros, review.Cons, review.Date, fakeProbability)
	return err
}

func GetExistingProductAnalysisGlobal(productID int) (analysisID int, totalReviews int, reviews []models.Review, err error) {
	row := DB.QueryRow(`
        SELECT id, total_reviews FROM product_analysis
        WHERE product_id = $1
        ORDER BY analyzed_at DESC LIMIT 1`,
		productID)
	err = row.Scan(&analysisID, &totalReviews)
	if err == sql.ErrNoRows {
		return 0, 0, nil, nil
	}
	if err != nil {
		return 0, 0, nil, err
	}
	rows, err := DB.Query(`
        SELECT review_text, review_rating, review_pros, review_cons, review_date, fake_probability
        FROM review_analysis
        WHERE analysis_id = $1`, analysisID)
	if err != nil {
		return 0, 0, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var rev models.Review
		rows.Scan(&rev.Text, &rev.Rating, &rev.Pros, &rev.Cons, &rev.Date, &rev.FakeProbability)
		reviews = append(reviews, rev)
	}
	return analysisID, totalReviews, reviews, nil
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

func GetAllProductAnalyses() ([]ProductAnalysisSummary, error) {
	rows, err := DB.Query(`
        SELECT pa.id, pa.user_id, u.email, pa.product_id, pa.product_url, pa.total_reviews, pa.analyzed_at
        FROM product_analysis pa
        JOIN users u ON pa.user_id = u.id
        ORDER BY pa.analyzed_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ProductAnalysisSummary
	for rows.Next() {
		var a ProductAnalysisSummary
		err := rows.Scan(&a.AnalysisID, &a.UserID, &a.UserEmail, &a.ProductID, &a.ProductURL, &a.TotalReviews, &a.AnalyzedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func DeleteProductAnalysis(analysisID int) error {
	_, err := DB.Exec(`DELETE FROM product_analysis WHERE id = $1`, analysisID)
	return err
}

type GlobalStats struct {
	TotalUsers           int     `json:"total_users"`
	TotalAnalyses        int     `json:"total_analyses"`
	TotalReviewsAnalyzed int     `json:"total_reviews_analyzed"`
	GlobalAvgFake        float64 `json:"global_avg_fake"`
}

func GetGlobalStats() (GlobalStats, error) {
	var stats GlobalStats
	row := DB.QueryRow(`
        SELECT 
            (SELECT COUNT(*) FROM users) AS total_users,
            (SELECT COUNT(*) FROM product_analysis) AS total_analyses,
            (SELECT COUNT(*) FROM review_analysis) AS total_reviews_analyzed,
            COALESCE((SELECT AVG(fake_probability) FROM review_analysis), 0) AS global_avg_fake
    `)
	err := row.Scan(&stats.TotalUsers, &stats.TotalAnalyses, &stats.TotalReviewsAnalyzed, &stats.GlobalAvgFake)
	return stats, err
}

func GetUserProductAnalyses(userID int) ([]ProductAnalysisSummary, error) {
	rows, err := DB.Query(`
        SELECT pa.id, pa.user_id, u.email, pa.product_id, pa.product_url, pa.total_reviews, pa.analyzed_at
        FROM product_analysis pa
        JOIN users u ON pa.user_id = u.id
        WHERE pa.user_id = $1
        ORDER BY pa.analyzed_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ProductAnalysisSummary
	for rows.Next() {
		var a ProductAnalysisSummary
		err := rows.Scan(&a.AnalysisID, &a.UserID, &a.UserEmail, &a.ProductID, &a.ProductURL, &a.TotalReviews, &a.AnalyzedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}
