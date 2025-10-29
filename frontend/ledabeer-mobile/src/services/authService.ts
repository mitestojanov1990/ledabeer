/**
 * Authentication Service
 *
 * Handles user authentication, token management, and session persistence
 */

import AsyncStorage from '@react-native-async-storage/async-storage';

// Types
export interface User {
  id: string;
  username: string;
  email: string;
  display_name: string;
  avatar_url: string;
  is_active: boolean;
  created_at: string;
  last_login_at?: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  display_name: string;
}

export interface AuthResponse {
  user: User;
  token: AuthTokens;
}

// Configuration
const AUTH_BASE_URL = 'http://192.168.0.140:9001/api/auth';
const TOKEN_STORAGE_KEY = 'auth_tokens';
const USER_STORAGE_KEY = 'current_user';

class AuthService {
  private tokens: AuthTokens | null = null;
  private currentUser: User | null = null;
  private listeners: Array<(isAuthenticated: boolean) => void> = [];

  constructor() {
    this.loadStoredAuth();
  }

  // Public methods
  async register(data: RegisterRequest): Promise<AuthResponse> {
    try {
      const response = await fetch(`${AUTH_BASE_URL}/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });

      // Check if response is JSON
      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text();
        throw new Error(
          `Server returned non-JSON response: ${text.substring(0, 100)}`
        );
      }

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Registration failed');
      }

      return result;
    } catch (error) {
      console.error('[AuthService] Registration failed:', error);
      throw error;
    }
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    try {
      const response = await fetch(`${AUTH_BASE_URL}/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });

      // Check if response is JSON
      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text();
        throw new Error(
          `Server returned non-JSON response: ${text.substring(0, 100)}`
        );
      }

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Login failed');
      }

      // Store tokens and user data
      await this.setAuthData(result.user, result.token);

      return result;
    } catch (error) {
      console.error('[AuthService] Login failed:', error);
      throw error;
    }
  }

  async logout(): Promise<void> {
    try {
      // Call logout endpoint if we have a token
      if (this.tokens?.access_token) {
        await fetch(`${AUTH_BASE_URL}/logout`, {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${this.tokens.access_token}`,
          },
        });
      }
    } catch (error) {
      console.error('[AuthService] Logout request failed:', error);
      // Continue with local logout even if server request fails
    } finally {
      // Clear local data
      await this.clearAuthData();
    }
  }

  async refreshToken(): Promise<AuthTokens> {
    if (!this.tokens?.refresh_token) {
      throw new Error('No refresh token available');
    }

    try {
      const response = await fetch(`${AUTH_BASE_URL}/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          refresh_token: this.tokens.refresh_token,
        }),
      });

      // Check if response is JSON
      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text();
        throw new Error(
          `Server returned non-JSON response: ${text.substring(0, 100)}`
        );
      }

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Token refresh failed');
      }

      // Update stored tokens
      this.tokens = result;
      await AsyncStorage.setItem(
        TOKEN_STORAGE_KEY,
        JSON.stringify(this.tokens)
      );

      return result;
    } catch (error) {
      console.error('[AuthService] Token refresh failed:', error);
      // If refresh fails, clear auth data
      await this.clearAuthData();
      throw error;
    }
  }

  async getCurrentUser(): Promise<User> {
    if (!this.tokens?.access_token) {
      throw new Error('No access token available');
    }

    try {
      const response = await fetch(`${AUTH_BASE_URL}/me`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${this.tokens.access_token}`,
        },
      });

      // Check if response is JSON
      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text();
        throw new Error(
          `Server returned non-JSON response: ${text.substring(0, 100)}`
        );
      }

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to get user info');
      }

      // Update stored user data
      this.currentUser = result;
      await AsyncStorage.setItem(
        USER_STORAGE_KEY,
        JSON.stringify(this.currentUser)
      );

      return result;
    } catch (error) {
      console.error('[AuthService] Get current user failed:', error);
      throw error;
    }
  }

  isAuthenticated(): boolean {
    return this.tokens !== null && this.currentUser !== null;
  }

  getCurrentUserSync(): User | null {
    return this.currentUser;
  }

  getAccessToken(): string | null {
    return this.tokens?.access_token || null;
  }

  // Event listeners
  addAuthListener(listener: (isAuthenticated: boolean) => void): () => void {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter((l) => l !== listener);
    };
  }

  private notifyListeners(): void {
    const isAuthenticated = this.isAuthenticated();
    this.listeners.forEach((listener) => listener(isAuthenticated));
  }

  // Private methods
  private async setAuthData(user: User, tokens: AuthTokens): Promise<void> {
    this.currentUser = user;
    this.tokens = tokens;

    // Store in AsyncStorage
    await Promise.all([
      AsyncStorage.setItem(USER_STORAGE_KEY, JSON.stringify(user)),
      AsyncStorage.setItem(TOKEN_STORAGE_KEY, JSON.stringify(tokens)),
    ]);

    this.notifyListeners();
  }

  private async clearAuthData(): Promise<void> {
    this.currentUser = null;
    this.tokens = null;

    // Clear from AsyncStorage
    await Promise.all([
      AsyncStorage.removeItem(USER_STORAGE_KEY),
      AsyncStorage.removeItem(TOKEN_STORAGE_KEY),
    ]);

    this.notifyListeners();
  }

  private async loadStoredAuth(): Promise<void> {
    try {
      const [userData, tokenData] = await Promise.all([
        AsyncStorage.getItem(USER_STORAGE_KEY),
        AsyncStorage.getItem(TOKEN_STORAGE_KEY),
      ]);

      if (userData && tokenData) {
        this.currentUser = JSON.parse(userData);
        this.tokens = JSON.parse(tokenData);

        // Verify token is still valid by getting current user
        try {
          await this.getCurrentUser();
        } catch (error) {
          // Token is invalid, clear auth data
          await this.clearAuthData();
        }
      }
    } catch (error) {
      console.error('[AuthService] Failed to load stored auth:', error);
      await this.clearAuthData();
    }
  }
}

// Export singleton instance
export const authService = new AuthService();
export default authService;
