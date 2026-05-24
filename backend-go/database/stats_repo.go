package database

import "fake-review-ai2/models"

func (s *Store) GetGlobalStats() (models.GlobalStats, error) {
	var stats models.GlobalStats
	row := s.DB.QueryRow(`
        SELECT
            (SELECT COUNT(*) FROM users) AS total_users,
            (SELECT COUNT(*) FROM product_analysis) AS total_analyses,
            (SELECT COUNT(*) FROM review_analysis) AS total_reviews_analyzed,
            COALESCE((SELECT AVG(fake_probability) FROM review_analysis), 0) AS global_avg_fake
    `)
	err := row.Scan(&stats.TotalUsers, &stats.TotalAnalyses, &stats.TotalReviewsAnalyzed, &stats.GlobalAvgFake)
	return stats, err
}
