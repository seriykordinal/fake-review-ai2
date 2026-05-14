import { api } from '../utils/api.js';
export class AuthService {
    static saveToken(token) {
        localStorage.setItem(this.TOKEN_KEY, token);
    }
    static getToken() {
        return localStorage.getItem(this.TOKEN_KEY);
    }
    static removeToken() {
        localStorage.removeItem(this.TOKEN_KEY);
    }
    static async login(email, password) {
        return api.post('/api/login', { email, password });
    }
    static async register(email, password) {
        return api.post('/api/register', { email, password });
    }
    static async verify(email, code) {
        return api.post('/api/verify', { email, code });
    }
    static async getProfile(token) {
        return api.get('/api/profile', token);
    }
}
AuthService.TOKEN_KEY = 'token';
//# sourceMappingURL=auth.js.map