import { conversationService, Conversation, MessagePreview } from './conversationService';
import { authService } from './authService';

export interface RealtimeMessage {
  conversation_id: string;
  message: MessagePreview;
  type: 'new_message' | 'conversation_update' | 'user_online' | 'user_offline';
}

export interface RealtimeServiceConfig {
  baseUrl: string;
  reconnectInterval: number;
  maxReconnectAttempts: number;
}

class RealtimeService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private isConnecting = false;
  private messageHandlers: ((message: RealtimeMessage) => void)[] = [];
  private conversationUpdateHandlers: ((conversation: Conversation) => void)[] = [];
  private config: RealtimeServiceConfig;

  constructor(config: RealtimeServiceConfig) {
    this.config = config;
  }

  // Connect to WebSocket
  async connect(): Promise<void> {
    if (this.ws?.readyState === WebSocket.OPEN || this.isConnecting) {
      return;
    }

    this.isConnecting = true;

    try {
      const token = authService.getAccessToken();
      if (!token) {
        throw new Error('No authentication token available');
      }

      const wsUrl = `${this.config.baseUrl.replace('http', 'ws')}/ws?token=${encodeURIComponent(token)}`;
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('[RealtimeService] Connected to WebSocket');
        this.reconnectAttempts = 0;
        this.isConnecting = false;
      };

      this.ws.onmessage = (event) => {
        try {
          const message: RealtimeMessage = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (error) {
          console.error('[RealtimeService] Failed to parse message:', error);
        }
      };

      this.ws.onclose = (event) => {
        console.log('[RealtimeService] WebSocket closed:', event.code, event.reason);
        this.isConnecting = false;
        this.scheduleReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('[RealtimeService] WebSocket error:', error);
        this.isConnecting = false;
      };

    } catch (error) {
      console.error('[RealtimeService] Failed to connect:', error);
      this.isConnecting = false;
      throw error;
    }
  }

  // Disconnect from WebSocket
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.reconnectAttempts = 0;
  }

  // Add message handler
  onMessage(handler: (message: RealtimeMessage) => void): void {
    this.messageHandlers.push(handler);
  }

  // Remove message handler
  removeMessageHandler(handler: (message: RealtimeMessage) => void): void {
    const index = this.messageHandlers.indexOf(handler);
    if (index > -1) {
      this.messageHandlers.splice(index, 1);
    }
  }

  // Add conversation update handler
  onConversationUpdate(handler: (conversation: Conversation) => void): void {
    this.conversationUpdateHandlers.push(handler);
  }

  // Remove conversation update handler
  removeConversationUpdateHandler(handler: (conversation: Conversation) => void): void {
    const index = this.conversationUpdateHandlers.indexOf(handler);
    if (index > -1) {
      this.conversationUpdateHandlers.splice(index, 1);
    }
  }

  // Send message through WebSocket
  sendMessage(message: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('[RealtimeService] WebSocket not connected, cannot send message');
    }
  }

  // Handle incoming messages
  private handleMessage(message: RealtimeMessage): void {
    console.log('[RealtimeService] Received message:', message);

    // Notify message handlers
    this.messageHandlers.forEach(handler => {
      try {
        handler(message);
      } catch (error) {
        console.error('[RealtimeService] Error in message handler:', error);
      }
    });

    // Handle specific message types
    switch (message.type) {
      case 'new_message':
        this.handleNewMessage(message);
        break;
      case 'conversation_update':
        this.handleConversationUpdate(message);
        break;
      case 'user_online':
      case 'user_offline':
        this.handleUserStatusUpdate(message);
        break;
    }
  }

  // Handle new message
  private async handleNewMessage(message: RealtimeMessage): Promise<void> {
    try {
      // Get updated conversation
      const conversation = await conversationService.getConversation(message.conversation_id);
      
      // Notify conversation update handlers
      this.conversationUpdateHandlers.forEach(handler => {
        try {
          handler(conversation);
        } catch (error) {
          console.error('[RealtimeService] Error in conversation update handler:', error);
        }
      });
    } catch (error) {
      console.error('[RealtimeService] Failed to handle new message:', error);
    }
  }

  // Handle conversation update
  private handleConversationUpdate(message: RealtimeMessage): void {
    // This would typically contain the full conversation object
    // For now, we'll just log it
    console.log('[RealtimeService] Conversation updated:', message);
  }

  // Handle user status update
  private handleUserStatusUpdate(message: RealtimeMessage): void {
    console.log('[RealtimeService] User status updated:', message);
    // This could trigger UI updates for online/offline status
  }

  // Schedule reconnection
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      console.log('[RealtimeService] Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.config.reconnectInterval * this.reconnectAttempts;

    console.log(`[RealtimeService] Scheduling reconnection in ${delay}ms (attempt ${this.reconnectAttempts})`);

    setTimeout(() => {
      this.connect().catch(error => {
        console.error('[RealtimeService] Reconnection failed:', error);
      });
    }, delay);
  }

  // Check if connected
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  // Get connection state
  getConnectionState(): string {
    if (!this.ws) return 'disconnected';
    
    switch (this.ws.readyState) {
      case WebSocket.CONNECTING:
        return 'connecting';
      case WebSocket.OPEN:
        return 'connected';
      case WebSocket.CLOSING:
        return 'closing';
      case WebSocket.CLOSED:
        return 'closed';
      default:
        return 'unknown';
    }
  }
}

// Create singleton instance
export const realtimeService = new RealtimeService({
  baseUrl: 'http://192.168.0.140:9001',
  reconnectInterval: 3000, // 3 seconds
  maxReconnectAttempts: 5,
});
