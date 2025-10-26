/**
 * HTTP Backend Client
 *
 * REST API client for connecting to the Ledabeer backend HTTP server
 * Uses port 8080 (not gRPC which doesn't work in browsers)
 */

const BACKEND_URL = 'http://localhost:8080';

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
}

/**
 * HTTP Client for backend API
 */
class HttpBackendClient {
  private baseUrl: string;
  private connected: boolean = false;

  constructor(baseUrl: string = BACKEND_URL) {
    this.baseUrl = baseUrl;
  }

  /**
   * Test connection to backend
   */
  async connect(): Promise<void> {
    try {
      console.log('[HttpBackendClient] Connecting to backend at', this.baseUrl);

      // Test if server is accessible
      const response = await fetch(`${this.baseUrl}/health`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        this.connected = true;
        console.log('[HttpBackendClient] Connected successfully');
      } else {
        throw new Error(`Server returned ${response.status}`);
      }
    } catch (error) {
      console.error('[HttpBackendClient] Connection failed:', error);
      throw new Error(`Failed to connect to backend: ${error}`);
    }
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.connected;
  }

  /**
   * Make GET request
   */
  async get<T = any>(endpoint: string): Promise<ApiResponse<T>> {
    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      const data = await response.json();

      if (response.ok) {
        return { success: true, data };
      } else {
        return { success: false, error: data.error || 'Request failed' };
      }
    } catch (error) {
      console.error('[HttpBackendClient] GET error:', error);
      return { success: false, error: String(error) };
    }
  }

  /**
   * Make POST request
   */
  async post<T = any>(endpoint: string, body: any): Promise<ApiResponse<T>> {
    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
      });

      const data = await response.json();

      if (response.ok) {
        return { success: true, data };
      } else {
        return { success: false, error: data.error || 'Request failed' };
      }
    } catch (error) {
      console.error('[HttpBackendClient] POST error:', error);
      return { success: false, error: String(error) };
    }
  }

  /**
   * Get WebSocket URL
   */
  getWebSocketUrl(): string {
    return this.baseUrl.replace('http', 'ws');
  }
}

// Singleton instance
let httpClientInstance: HttpBackendClient | null = null;

/**
 * Get singleton HTTP client instance
 */
export function getHttpClient(): HttpBackendClient {
  if (!httpClientInstance) {
    httpClientInstance = new HttpBackendClient();
  }
  return httpClientInstance;
}
