import { api } from '../utils/api.js';
import type { Stats, User, ProductAnalysis } from '../types/index.js';

export class AdminService {
  static async getStats(token: string): Promise<Stats> {
    return api.get<Stats>('/admin/stats', token);
  }

  static async getUsers(token: string): Promise<User[]> {
    return api.get<User[]>('/admin/users', token);
  }

  static async deleteUser(userId: number, token: string): Promise<void> {
    return api.delete(`/admin/users/${userId}`, token);
  }

  static async changeRole(email: string, role: string, token: string): Promise<void> {
    return api.put('/admin/users/role', { email, role }, token);
  }

  static async getProducts(token: string): Promise<ProductAnalysis[]> {
    return api.get<ProductAnalysis[]>('/admin/products', token);
  }

  static async deleteProduct(analysisId: number, token: string): Promise<void> {
    return api.delete(`/admin/products/${analysisId}`, token);
  }
}