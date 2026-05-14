import { api } from '../utils/api.js';
import type { Profile } from '../types/index.js';

export class AuthService {
  private static TOKEN_KEY = 'token';

  static saveToken(token: string): void {
    localStorage.setItem(this.TOKEN_KEY, token);
  }

  static getToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY);
  }

  static removeToken(): void {
    localStorage.removeItem(this.TOKEN_KEY);
  }

  static async login(email: string, password: string): Promise<{ token: string }> {
    return api.post<{ token: string }>('/api/login', { email, password });
  }

  static async register(email: string, password: string): Promise<{ token?: string; message?: string }> {
    return api.post<{ token?: string; message?: string }>('/api/register', { email, password });
  }

  static async verify(email: string, code: string): Promise<{ token: string }> {
    return api.post<{ token: string }>('/api/verify', { email, code });
  }

  static async getProfile(token: string): Promise<Profile> {
    return api.get<Profile>('/api/profile', token);
  }
}
