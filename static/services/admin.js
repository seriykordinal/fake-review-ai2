import { api } from '../utils/api.js';
export class AdminService {
    static async getStats(token) {
        return api.get('/admin/stats', token);
    }
    static async getUsers(token) {
        return api.get('/admin/users', token);
    }
    static async deleteUser(userId, token) {
        return api.delete(`/admin/users/${userId}`, token);
    }
    static async changeRole(email, role, token) {
        return api.put('/admin/users/role', { email, role }, token);
    }
    static async getProducts(token) {
        return api.get('/admin/products', token);
    }
    static async deleteProduct(analysisId, token) {
        return api.delete(`/admin/products/${analysisId}`, token);
    }
}
//# sourceMappingURL=admin.js.map