export class ApiClient {
    constructor(baseUrl = '') {
        this.baseUrl = baseUrl;
    }
    async request(endpoint, options = {}) {
        const res = await fetch(`${this.baseUrl}${endpoint}`, options);
        const contentType = res.headers.get('content-type');
        let data;
        if (contentType?.includes('application/json')) {
            data = await res.json();
        }
        else {
            const text = await res.text();
            try {
                data = JSON.parse(text);
            }
            catch {
                data = { message: text };
            }
        }
        if (!res.ok) {
            throw new Error(data.error || data.message || `HTTP ${res.status}`);
        }
        return data;
    }
    get(endpoint, token) {
        const headers = {};
        if (token)
            headers['Authorization'] = `Bearer ${token}`;
        return this.request(endpoint, { method: 'GET', headers });
    }
    post(endpoint, body, token) {
        const headers = { 'Content-Type': 'application/json' };
        if (token)
            headers['Authorization'] = `Bearer ${token}`;
        return this.request(endpoint, {
            method: 'POST',
            headers,
            body: JSON.stringify(body),
        });
    }
    put(endpoint, body, token) {
        const headers = { 'Content-Type': 'application/json' };
        if (token)
            headers['Authorization'] = `Bearer ${token}`;
        return this.request(endpoint, {
            method: 'PUT',
            headers,
            body: JSON.stringify(body),
        });
    }
    delete(endpoint, token) {
        const headers = {};
        if (token)
            headers['Authorization'] = `Bearer ${token}`;
        return this.request(endpoint, { method: 'DELETE', headers });
    }
}
export const api = new ApiClient();
//# sourceMappingURL=api.js.map