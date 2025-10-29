/**
 * HTTP Backend Client Implementation
 *
 * Implements HTTP REST API calls to the backend through Envoy proxy
 * This replaces the gRPC-Web client to avoid compatibility issues on Mac ARM
 */

import { authService } from '../authService';

// Backend HTTP endpoints (Envoy proxy)
const HTTP_HOST = 'http://192.168.0.140:9001'; // Alice node by default

// Message types matching backend responses
export interface Message {
  message_id: string;
  from_peer_id: string;
  content: Uint8Array;
  timestamp: number;
}

export interface SendMessageResponse {
  message_id: string;
  timestamp: number;
}

export interface Peer {
  id: string;
  name: string;
  online: boolean;
  last_seen: number;
  addresses: string[];
}

export interface GetPeersResponse {
  peers: Peer[];
}

export interface SendMessageRequest {
  to: string;
  content: string;
}

/**
 * HTTP Backend Client
 */
export class HttpBackendClient {
  private host: string;
  private connected: boolean = false;

  constructor(host: string = HTTP_HOST) {
    this.host = host;
  }

  private getAuthHeaders(): Record<string, string> {
    const token = authService.getAccessToken();
    if (token) {
      return {
        Authorization: `Bearer ${token}`,
      };
    }
    return {};
  }

  /**
   * Connect to backend
   */
  async connect(): Promise<void> {
    try {
      console.log('[HttpBackendClient] Connecting to', this.host);

      // Test connection by calling /api/peers
      const response = await fetch(`${this.host}/api/peers`, {
        method: 'GET',
        headers: {
          'Cache-Control': 'no-cache',
          ...this.getAuthHeaders(),
        },
      });

      if (response.ok) {
        this.connected = true;
        console.log('[HttpBackendClient] Connected successfully');
      } else {
        throw new Error(
          `Connection failed: ${response.status} ${response.statusText}`
        );
      }
    } catch (error) {
      console.error('[HttpBackendClient] Connection failed:', error);
      throw error;
    }
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.connected;
  }

  /**
   * Send a text message to a peer
   */
  async sendMessage(
    toPeerId: string,
    content: string
  ): Promise<SendMessageResponse> {
    try {
      console.log('[HttpBackendClient] Sending message to', toPeerId);

      const request: SendMessageRequest = {
        to: toPeerId,
        content: content,
      };

      const response = await fetch(`${this.host}/api/send-message`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...this.getAuthHeaders(),
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        throw new Error(
          `Send message failed: ${response.status} ${response.statusText}`
        );
      }

      const data = await response.json();
      console.log('[HttpBackendClient] Message sent:', data.messageId);

      return {
        message_id: data.messageId,
        timestamp: Date.now(), // Backend should return this
      };
    } catch (error) {
      console.error('[HttpBackendClient] sendMessage error:', error);
      throw error;
    }
  }

  /**
   * Send a message to a group
   */
  async sendGroupMessage(
    groupId: string,
    content: string
  ): Promise<SendMessageResponse> {
    // For now, treat group messages as regular peer messages
    // TODO: Implement proper group messaging
    console.log(
      '[HttpBackendClient] Group messaging not yet implemented, treating as peer message'
    );
    return this.sendMessage(groupId, content);
  }

  /**
   * Get message history with a peer
   */
  async getMessageHistory(
    peerId: string,
    limit: number = 50
  ): Promise<Message[]> {
    // TODO: Implement message history endpoint
    console.log('[HttpBackendClient] Message history not yet implemented');
    return [];
  }

  /**
   * Subscribe to incoming messages
   *
   * Note: This is a simplified implementation using polling
   * In a real implementation, you'd use WebSockets or Server-Sent Events
   */
  subscribeToMessages(callback: (message: Message) => void): () => void {
    console.log('[HttpBackendClient] Subscribing to messages (polling)');

    let cancelled = false;
    let pollInterval: NodeJS.Timeout | null = null;

    // Start polling for messages
    const pollMessages = async () => {
      if (cancelled) return;

      try {
        // TODO: Implement message polling endpoint
        // For now, we'll just log that we're polling
        console.log('[HttpBackendClient] Polling for messages...');
      } catch (error) {
        if (!cancelled) {
          console.error('[HttpBackendClient] Polling error:', error);
        }
      }
    };

    // Poll every 2 seconds
    pollInterval = setInterval(pollMessages, 2000);

    // Return unsubscribe function
    return () => {
      cancelled = true;
      if (pollInterval) {
        clearInterval(pollInterval);
      }
      console.log('[HttpBackendClient] Unsubscribed from messages');
    };
  }

  /**
   * Get all peers
   */
  async getPeers(): Promise<Peer[]> {
    try {
      console.log('[HttpBackendClient] Getting peers');

      const response = await fetch(`${this.host}/api/peers`, {
        method: 'GET',
        headers: {
          'Cache-Control': 'no-cache',
          ...this.getAuthHeaders(),
        },
      });

      if (!response.ok) {
        throw new Error(
          `Get peers failed: ${response.status} ${response.statusText}`
        );
      }

      const data: GetPeersResponse = await response.json();
      console.log(`[HttpBackendClient] Got ${data.peers.length} peers`);

      return data.peers;
    } catch (error) {
      console.error('[HttpBackendClient] getPeers error:', error);
      throw error;
    }
  }

  /**
   * Get a specific peer by ID
   */
  async getPeer(peerId: string): Promise<Peer | null> {
    try {
      console.log(`[HttpBackendClient] Getting peer: ${peerId}`);

      const peers = await this.getPeers();
      const peer = peers.find((p) => p.id === peerId);

      console.log(`[HttpBackendClient] Peer found: ${!!peer}`);
      return peer || null;
    } catch (error) {
      console.error('[HttpBackendClient] getPeer error:', error);
      throw error;
    }
  }

  /**
   * Get only connected peers
   */
  async getConnectedPeers(): Promise<Peer[]> {
    try {
      console.log('[HttpBackendClient] Getting connected peers');

      const peers = await this.getPeers();
      const connectedPeers = peers.filter((p) => p.online);

      console.log(
        `[HttpBackendClient] Got ${connectedPeers.length} connected peers`
      );
      return connectedPeers;
    } catch (error) {
      console.error('[HttpBackendClient] getConnectedPeers error:', error);
      throw error;
    }
  }

  /**
   * Disconnect from backend
   */
  disconnect(): void {
    this.connected = false;
    console.log('[HttpBackendClient] Disconnected');
  }
}

// Singleton instance
let clientInstance: HttpBackendClient | null = null;

/**
 * Get singleton client instance
 */
export function getHttpBackendClient(): HttpBackendClient {
  if (!clientInstance) {
    clientInstance = new HttpBackendClient();
  }
  return clientInstance;
}

// Alias for compatibility
export const getHttpClient = getHttpBackendClient;
