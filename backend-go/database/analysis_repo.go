package database

import (
	"database/sql"
	"fake-review-ai2/models"
)

func (s *Store) CreateAnalysisTables() error {
	_, err := s.DB.Exec(`
        CREATE TABLE IF NOT EXISTS product_analysis (
            id SERIAL PRIMARY KEY,
            user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
            product_url TEXT NOT NULL,
            product_id INTEGER NOT NULL,
            total_reviews INTEGER DEFAULT 0,
            analyzed_at TIMESTAMP DEFAULT NOW()
        )`)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`
        CREATE TABLE IF NOT EXISTS review_analysis (
            id SERIAL PRIMARY KEY,
            analysis_id INTEGER REFERENCES product_analysis(id) ON DELETE CASCADE,
            review_text TEXT,
            review_rating VARCHAR(10),
            review_pros TEXT,
            review_cons TEXT,
            review_date VARCHAR(100),
            fake_probability DOUBLE PRECISION DEFAULT 0
        )`)
	return err
}

func (s *Store) SaveProductAnalysis(userID int, productURL string, productID int, totalReviews int) (int, error) {
	var analysisID int
	err := s.DB.QueryRow(`
        INSERT INTO product_analysis (user_id, product_url, product_id, total_reviews)
        VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, productURL, productID, totalReviews).Scan(&analysisID)
	return analysisID, err
}

func (s *Store) SaveReviewAnalysis(analysisID int, review models.Review, fakeProbability float64) error {
	_, err := s.DB.Exec(`
        INSERT INTO review_analysis (analysis_id, review_text, review_rating, review_pros, review_cons, review_date, fake_probability)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		analysisID, review.Text, review.Rating, review.Pros, review.Cons, review.Date, fakeProbability)
	return err
}

func (s *Store) GetExistingProductAnalysisGlobal(productID int) (analysisID int, totalReviews int, reviews []models.Review, err error) {
	row := s.DB.QueryRow(`
        SELECT id, total_reviews FROM product_analysis
        WHERE product_id = $1
        ORDER BY analyzed_at DESC LIMIT 1`, productID)
	err = row.Scan(&analysisID, &totalReviews)
	if err == sql.ErrNoRows {
		return 0, 0, nil, nil
	}
	if err != nil {
		return 0, 0, nil, err
	}
	rows, err := s.DB.Query(`
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

func (s *Store) GetAllProductAnalysis() ([]models.ProductAnalysisSummary, error) {
	rows, err := s.DB.Query(`
        SELECT pa.id, pa.user_id, u.email, pa.product_id, pa.product_url, pa.total_reviews, pa.analyzed_at
        FROM product_analysis pa
        JOIN users u ON pa.user_id = u.id
        ORDER BY pa.analyzed_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.ProductAnalysisSummary
	for rows.Next() {
		var a models.ProductAnalysisSummary
		if err := rows.Scan(&a.AnalysisID, &a.UserID, &a.UserEmail, &a.ProductID, &a.ProductURL, &a.TotalReviews, &a.AnalyzedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *Store) GetProductAnalysisByID(analysisID int) (*models.ProductAnalysisSummary, error) {
	var a models.ProductAnalysisSummary
	row := s.DB.QueryRow(`
        SELECT pa.id, pa.user_id, u.email, pa.product_id, pa.product_url, pa.total_reviews, pa.analyzed_at
        FROM product_analysis pa
        JOIN users u ON pa.user_id = u.id
        WHERE pa.id = $1`, analysisID)
	err := row.Scan(&a.AnalysisID, &a.UserID, &a.UserEmail, &a.ProductID, &a.ProductURL, &a.TotalReviews, &a.AnalyzedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Store) GetUserProductAnalysis(userID int) ([]models.ProductAnalysisSummary, error) {
	rows, err := s.DB.Query(`
        SELECT pa.id, pa.user_id, u.email, pa.product_id, pa.product_url, pa.total_reviews, pa.analyzed_at
        FROM product_analysis pa
        JOIN users u ON pa.user_id = u.id
        WHERE pa.user_id = $1
        ORDER BY pa.analyzed_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.ProductAnalysisSummary
	for rows.Next() {
		var a models.ProductAnalysisSummary
		if err := rows.Scan(&a.AnalysisID, &a.UserID, &a.UserEmail, &a.ProductID, &a.ProductURL, &a.TotalReviews, &a.AnalyzedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *Store) DeleteProductAnalysis(analysisID int) error {
	_, err := s.DB.Exec(`DELETE FROM product_analysis WHERE id = $1`, analysisID)
	return err
}
