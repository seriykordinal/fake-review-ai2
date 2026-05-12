import { api } from '../utils/api.js';
export class ProductService {
    static async analyzeSingleReview(text) {
        return api.post('/api/analyze', { text });
    }
    static async analyzeProduct(productUrl, token) {
        return api.post('/api/analyze_product', { product_url: productUrl }, token);
    }
}
//# sourceMappingURL=product.js.map