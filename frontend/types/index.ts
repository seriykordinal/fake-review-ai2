// ========== Пользователи и роли ==========
export type UserRole = 'user' | 'admin' | 'super_admin';

export interface User {
  id: number;
  email: string;
  role: UserRole;
  created_at?: string;
}

export interface Profile extends User {
  created_at: string;
}

// ========== Статистика ==========
export interface Stats {
  total_users: number;
  total_analyses: number;
  total_reviews_analyzed: number;
  global_avg_fake: number;
}

// ========== Анализы товаров ==========
export interface ProductAnalysis {
  analysis_id: number;
  user_email: string;
  product_id: string;
  total_reviews: number;
  analyzed_at: string;
}

export interface ReviewResult {
  rating?: number;
  date?: string;
  text?: string;
  pros?: string;
  cons?: string;
  fake_probability: number;
}

export interface AnalyzeProductResponse {
  total_reviews: number;
  average_fake_probability?: number;
  results: ReviewResult[];
}

// ========== Ответы API ==========
export interface ApiError {
  error?: string;
  message?: string;
}

export type ApiResponse<T> = T & ApiError;