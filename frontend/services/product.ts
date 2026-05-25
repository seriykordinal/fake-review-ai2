import { api } from '../utils/api.js';
import type { AnalyzeProductResponse } from '../types/index.js';

export class ProductService {
  static async analyzeSingleReview(text: string): Promise<{ fake_probability: number }> {
    return api.post<{ fake_probability: number }>('/api/analyze', { text });
  }

  static async analyzeProduct(productUrl: string, token: string): Promise<AnalyzeProductResponse> {
    return api.post<AnalyzeProductResponse>('/api/analyze_product', { product_url: productUrl }, token);
  }
}