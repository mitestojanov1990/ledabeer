/**
 * Smart Backend Service
 *
 * Implements intelligent backend selection with gRPC primary and HTTP fallback
 * Tries gRPC first for better performance, falls back to HTTP on failure
 */

import {
  getGrpcBackendClient,
  GrpcBackendClient,
} from '../grpcBackend/grpcClient';
import {
  getHttpBackendClient,
  HttpBackendClient,
} from '../httpBackend/httpClient';

// Re-export types for convenience
export type {
  Message,
  SendMessageResponse,
  Peer,
  Group,
} from '../grpcBackend/grpcClient';

export class SmartBackendService {
  private grpcClient: GrpcBackendClient;
  private httpClient: HttpBackendClient;
  private currentMode: 'grpc' | 'http' | 'unknown' = 'unknown';
  private connectionAttempts: number = 0;
  private maxRetries: number = 3;
  private retryDelay: number = 1000; // 1 second

  constructor() {
    this.grpcClient = getGrpcBackendClient();
    this.httpClient = getHttpBackendClient();
  }

  async connect(): Promise<void> {
    console.log('[SmartBackend] Attempting to connect...');

    // Try gRPC first
    if (await this.tryGrpcConnection()) {
      this.currentMode = 'grpc';
      console.log('[SmartBackend] ✅ Connected via gRPC');
      return;
    }

    // Fallback to HTTP
    if (await this.tryHttpConnection()) {
      this.currentMode = 'http';
      console.log('[SmartBackend] ✅ Connected via HTTP (fallback)');
      return;
    }

    throw new Error('Failed to connect via both gRPC and HTTP');
  }

  private async tryGrpcConnection(): Promise<boolean> {
    try {
      console.log('[SmartBackend] Trying gRPC connection...');
      await this.grpcClient.connect();
      return true;
    } catch (error) {
      console.warn('[SmartBackend] gRPC connection failed:', error);
      return false;
    }
  }

  private async tryHttpConnection(): Promise<boolean> {
    try {
      console.log('[SmartBackend] Trying HTTP connection...');
      await this.httpClient.connect();
      return true;
    } catch (error) {
      console.warn('[SmartBackend] HTTP connection failed:', error);
      return false;
    }
  }

  isConnected(): boolean {
    if (this.currentMode === 'grpc') {
      return this.grpcClient.isConnected();
    } else if (this.currentMode === 'http') {
      return this.httpClient.isConnected();
    }
    return false;
  }

  disconnect(): void {
    if (this.currentMode === 'grpc') {
      this.grpcClient.disconnect();
    } else if (this.currentMode === 'http') {
      this.httpClient.disconnect();
    }
    this.currentMode = 'unknown';
    console.log('[SmartBackend] Disconnected');
  }

  getCurrentMode(): 'grpc' | 'http' | 'unknown' {
    return this.currentMode;
  }

  private async executeWithFallback<T>(
    grpcOperation: () => Promise<T>,
    httpOperation: () => Promise<T>,
    operationName: string
  ): Promise<T> {
    // If we're in gRPC mode, try gRPC first
    if (this.currentMode === 'grpc') {
      try {
        return await grpcOperation();
      } catch (error) {
        console.warn(
          `[SmartBackend] gRPC ${operationName} failed, falling back to HTTP:`,
          error
        );

        // Try to switch to HTTP mode
        if (await this.tryHttpConnection()) {
          this.currentMode = 'http';
          console.log('[SmartBackend] Switched to HTTP mode');
          return await httpOperation();
        }

        throw error;
      }
    }

    // If we're in HTTP mode or unknown, use HTTP
    if (this.currentMode === 'http' || this.currentMode === 'unknown') {
      try {
        return await httpOperation();
      } catch (error) {
        console.warn(
          `[SmartBackend] HTTP ${operationName} failed, trying gRPC:`,
          error
        );

        // Try to switch to gRPC mode
        if (await this.tryGrpcConnection()) {
          this.currentMode = 'grpc';
          console.log('[SmartBackend] Switched to gRPC mode');
          return await grpcOperation();
        }

        throw error;
      }
    }

    throw new Error('No backend mode available');
  }

  async getPeers(): Promise<import('../grpcBackend/grpcClient').Peer[]> {
    return this.executeWithFallback(
      () => this.grpcClient.getPeers(),
      () => this.httpClient.getPeers(),
      'getPeers'
    );
  }

  async getPeer(peerId: string): Promise<import('../grpcBackend/grpcClient').Peer | null> {
    return this.executeWithFallback(
      () => this.grpcClient.getPeer(peerId),
      () => this.httpClient.getPeer(peerId),
      'getPeer'
    );
  }

  async getMessageHistory(peerId: string): Promise<import('../grpcBackend/grpcClient').Message[]> {
    return this.executeWithFallback(
      () => this.grpcClient.getMessageHistory(peerId),
      () => this.httpClient.getMessageHistory(peerId),
      'getMessageHistory'
    );
  }

  async sendMessage(
    toPeerId: string,
    content: string
  ): Promise<import('../grpcBackend/grpcClient').SendMessageResponse> {
    return this.executeWithFallback(
      () => this.grpcClient.sendMessage(toPeerId, content),
      () => this.httpClient.sendMessage(toPeerId, content),
      'sendMessage'
    );
  }

  async sendGroupMessage(
    groupId: string,
    content: string
  ): Promise<import('../grpcBackend/grpcClient').SendMessageResponse> {
    return this.executeWithFallback(
      () => this.grpcClient.sendGroupMessage(groupId, content),
      () => this.httpClient.sendGroupMessage(groupId, content),
      'sendGroupMessage'
    );
  }

  subscribeToMessages(callback: (message: import('../grpcBackend/grpcClient').Message) => void): () => void {
    if (this.currentMode === 'grpc') {
      console.log('[SmartBackend] Using gRPC streaming for messages');
      return this.grpcClient.subscribeToMessages(callback);
    } else {
      console.log('[SmartBackend] Using HTTP polling for messages');
      return this.httpClient.subscribeToMessages(callback);
    }
  }

  async getConnectedPeers(): Promise<import('../grpcBackend/grpcClient').Peer[]> {
    return this.executeWithFallback(
      () => this.grpcClient.getConnectedPeers(),
      () => this.httpClient.getConnectedPeers(),
      'getConnectedPeers'
    );
  }

  // Debug method to show current status
  getDebugInfo(): { mode: string; connected: boolean; host: string } {
    return {
      mode: this.currentMode,
      connected: this.isConnected(),
      host: this.currentMode === 'grpc' ? 'gRPC' : 'HTTP',
    };
  }

  // Health check method
  async healthCheck(): Promise<{
    mode: string;
    status: string;
    details?: any;
  }> {
    try {
      if (this.currentMode === 'grpc') {
        const isConnected = this.grpcClient.isConnected();
        return {
          mode: 'grpc',
          status: isConnected ? 'healthy' : 'unhealthy',
          details: { connected: isConnected },
        };
      } else if (this.currentMode === 'http') {
        const isConnected = this.httpClient.isConnected();
        return {
          mode: 'http',
          status: isConnected ? 'healthy' : 'unhealthy',
          details: { connected: isConnected },
        };
      } else {
        return {
          mode: 'unknown',
          status: 'unhealthy',
          details: { error: 'No backend mode selected' },
        };
      }
    } catch (error) {
      return {
        mode: this.currentMode,
        status: 'unhealthy',
        details: { error: (error as Error).message },
      };
    }
  }
}

let serviceInstance: SmartBackendService | null = null;

export function getSmartBackendService(): SmartBackendService {
  if (!serviceInstance) {
    serviceInstance = new SmartBackendService();
  }
  return serviceInstance;
}
