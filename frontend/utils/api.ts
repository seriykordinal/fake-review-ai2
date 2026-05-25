import type { ApiError, ApiResponse } from '../types/index.js';

export class ApiClient {
  private baseUrl: string;

  constructor(baseUrl = '') {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const res = await fetch(`${this.baseUrl}${endpoint}`, options);
    const contentType = res.headers.get('content-type');
    let data: T & ApiError;

    if (contentType?.includes('application/json')) {
      data = await res.json();
    } else {
      const text = await res.text();
      try {
        data = JSON.parse(text);
      } catch {
        data = { message: text } as T & ApiError;
      }
    }

    if (!res.ok) {
      throw new Error(data.error || data.message || `HTTP ${res.status}`);
    }
    return data;
  }

  public get<T>(endpoint: string, token?: string): Promise<T> {
    const headers: HeadersInit = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;
    return this.request<T>(endpoint, { method: 'GET', headers });
  }

  public post<T>(endpoint: string, body: unknown, token?: string): Promise<T> {
    const headers: HeadersInit = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = `Bearer ${token}`;
    return this.request<T>(endpoint, {
      method: 'POST',
      headers,
      body: JSON.stringify(body),
    });
  }

  public put<T>(endpoint: string, body: unknown, token?: string): Promise<T> {
    const headers: HeadersInit = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = `Bearer ${token}`;
    return this.request<T>(endpoint, {
      method: 'PUT',
      headers,
      body: JSON.stringify(body),
    });
  }

  public delete<T>(endpoint: string, token?: string): Promise<T> {
    const headers: HeadersInit = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;
    return this.request<T>(endpoint, { method: 'DELETE', headers });
  }
}

export const api = new ApiClient();