/**
 * Backend Client
 *
 * Main gRPC client for connecting to the Ledabeer backend
 */

import * as grpc from '@grpc/grpc-js';
import { loadMessageProto, loadCallProto, loadMediaProto, getServiceClient } from './protoLoader';

// Backend server address
const BACKEND_URL = 'localhost:50051';

// Connection timeout
const CONNECTION_TIMEOUT_MS = 5000;

export class BackendClient {
  private messageClient: grpc.Client | null = null;
  private callClient: grpc.Client | null = null;
  private mediaClient: grpc.Client | null = null;
  private connected: boolean = false;

  /**
   * Initialize connection to backend
   */
  async connect(): Promise<void> {
    try {
      console.log('[BackendClient] Connecting to backend at', BACKEND_URL);

      // Load proto definitions
      const messageProto = loadMessageProto();
      const callProto = loadCallProto();
      const mediaProto = loadMediaProto();

      // Get service constructors
      const MessageService = getServiceClient(messageProto, 'ledabeer', 'MessageService');
      const CallService = getServiceClient(callProto, 'ledabeer', 'CallService');
      const MediaService = getServiceClient(mediaProto, 'ledabeer', 'MediaService');

      // Create insecure credentials for local development
      const credentials = grpc.credentials.createInsecure();

      // Create service clients
      this.messageClient = new MessageService(BACKEND_URL, credentials) as grpc.Client;
      this.callClient = new CallService(BACKEND_URL, credentials) as grpc.Client;
      this.mediaClient = new MediaService(BACKEND_URL, credentials) as grpc.Client;

      // Wait for connection
      await this.waitForReady(this.messageClient);

      this.connected = true;
      console.log('[BackendClient] Connected to backend successfully');
    } catch (error) {
      console.error('[BackendClient] Failed to connect:', error);
      throw new Error(`Failed to connect to backend: ${error}`);
    }
  }

  /**
   * Wait for client to be ready
   */
  private async waitForReady(client: grpc.Client): Promise<void> {
    return new Promise((resolve, reject) => {
      const deadline = new Date();
      deadline.setMilliseconds(deadline.getMilliseconds() + CONNECTION_TIMEOUT_MS);

      client.waitForReady(deadline, (error) => {
        if (error) {
          reject(error);
        } else {
          resolve();
        }
      });
    });
  }

  /**
   * Check if connected to backend
   */
  isConnected(): boolean {
    return this.connected;
  }

  /**
   * Get message service client
   */
  getMessageClient(): grpc.Client {
    if (!this.messageClient) {
      throw new Error('Message client not initialized. Call connect() first.');
    }
    return this.messageClient;
  }

  /**
   * Get call service client
   */
  getCallClient(): grpc.Client {
    if (!this.callClient) {
      throw new Error('Call client not initialized. Call connect() first.');
    }
    return this.callClient;
  }

  /**
   * Get media service client
   */
  getMediaClient(): grpc.Client {
    if (!this.mediaClient) {
      throw new Error('Media client not initialized. Call connect() first.');
    }
    return this.mediaClient;
  }

  /**
   * Disconnect from backend
   */
  disconnect(): void {
    if (this.messageClient) {
      this.messageClient.close();
      this.messageClient = null;
    }
    if (this.callClient) {
      this.callClient.close();
      this.callClient = null;
    }
    if (this.mediaClient) {
      this.mediaClient.close();
      this.mediaClient = null;
    }
    this.connected = false;
    console.log('[BackendClient] Disconnected from backend');
  }
}

// Singleton instance
let backendClientInstance: BackendClient | null = null;

/**
 * Get singleton backend client instance
 */
export function getBackendClient(): BackendClient {
  if (!backendClientInstance) {
    backendClientInstance = new BackendClient();
  }
  return backendClientInstance;
}
