/**
 * Backend Service
 *
 * HTTP-based service layer that connects to the real backend
 * Uses HTTP REST API endpoints instead of gRPC-Web for better compatibility
 */

import { getSmartBackendService } from './smartBackend/smartBackendService';

// Types
export interface Message {
  id: string;
  from: string;
  to: string;
  content: string;
  timestamp: number;
  encrypted: boolean;
  type?: 'text' | 'image' | 'video' | 'audio' | 'file';
}

export interface Peer {
  id: string;
  name: string;
  publicKey: string;
  online: boolean;
  lastSeen: number;
  addresses?: string[];
}

export interface Group {
  id: string;
  name: string;
  description: string;
  members: string[];
  admins: string[];
  createdAt: number;
}

// Configuration
const BACKEND_TIMEOUT = 3000; // 3 seconds timeout for connection

// Types are exported inline above

/**
 * Backend service interface
 */
export interface BackendService {
  // Connection
  connect(): Promise<void>;
  isConnected(): boolean;
  disconnect(): void;

  // User
  getCurrentUserId(): string;

  // Peers
  getPeers(): Promise<Peer[]>;
  getPeer(peerId: string): Promise<Peer | null>;

  // Groups
  getGroups(): Promise<Group[]>;
  getGroup(groupId: string): Promise<Group | null>;
  createGroup(
    name: string,
    description: string,
    memberIds: string[]
  ): Promise<Group>;

  // Conversations
  getAllConversations(): Promise<Array<Peer | Group>>;
  isGroup(id: string): boolean;

  // Messages
  getMessagesWithPeer(peerId: string): Promise<Message[]>;
  sendMessage(peerId: string, content: string): Promise<Message>;
  sendGroupMessage(groupId: string, content: string): Promise<Message>;
  onMessage(callback: (message: Message) => void): void;

  // Debug
  getDebugInfo(): { mode: string; connected: boolean; host: string };
}

/**
 * Real Smart Backend Implementation
 * Uses gRPC primary with HTTP fallback
 */
class RealBackendService implements BackendService {
  private smartClient = getSmartBackendService();
  private messageCallbacks: Array<(message: Message) => void> = [];

  async connect(): Promise<void> {
    return this.smartClient.connect();
  }

  isConnected(): boolean {
    return this.smartClient.isConnected();
  }

  disconnect(): void {
    this.smartClient.disconnect();
    this.messageCallbacks = [];
  }

  getCurrentUserId(): string {
    // For now, use a fixed peer ID
    // In real implementation, this would come from authentication
    return 'peer_local';
  }

  async getPeers(): Promise<Peer[]> {
    console.log('[RealBackend] Getting peers from backend');
    const peers = await this.smartClient.getPeers();
    console.log('[RealBackend] Raw peers from HTTP:', peers);

    // Convert HTTP Peer format to our Peer format
    const convertedPeers: Peer[] = peers.map((peer) => ({
      id: peer.id,
      name: peer.name,
      publicKey: '', // HTTP API doesn't provide publicKey yet
      online: peer.online,
      lastSeen: peer.last_seen || 0, // Convert from Unix timestamp
      addresses: peer.addresses || [],
    }));

    console.log(
      `[RealBackend] Got ${convertedPeers.length} peers from backend`
    );
    return convertedPeers;
  }

  async getPeer(peerId: string): Promise<Peer | null> {
    console.log(`[RealBackend] Getting peer ${peerId} from backend`);
    const peer = await this.smartClient.getPeer(peerId);

    if (!peer) {
      console.log(`[RealBackend] Peer ${peerId} not found in backend`);
      return null;
    }

    // Convert HTTP Peer format to our Peer format
    const convertedPeer: Peer = {
      id: peer.id,
      name: peer.name,
      publicKey: '', // HTTP API doesn't provide publicKey yet
      online: peer.online,
      lastSeen: peer.last_seen || 0, // Keep as number (Unix timestamp)
      addresses: peer.addresses || [],
    };

    console.log(`[RealBackend] Got peer ${peerId} from backend`);
    return convertedPeer;
  }

  async getGroups(): Promise<Group[]> {
    // TODO: Implement real group fetching via backend
    console.log('[RealBackend] Group fetching not yet implemented');
    return [];
  }

  async getGroup(groupId: string): Promise<Group | null> {
    const groups = await this.getGroups();
    return groups.find((g) => g.id === groupId) || null;
  }

  async createGroup(
    name: string,
    description: string,
    memberIds: string[]
  ): Promise<Group> {
    // TODO: Implement real group creation via backend
    console.log('[RealBackend] Group creation not yet implemented');
    throw new Error('Group creation not yet implemented');
  }

  async getAllConversations(): Promise<Array<Peer | Group>> {
    const [peers, groups] = await Promise.all([
      this.getPeers(),
      this.getGroups(),
    ]);
    return [...peers, ...groups];
  }

  isGroup(id: string): boolean {
    return id.startsWith('group_');
  }

  async getMessagesWithPeer(peerId: string): Promise<Message[]> {
    const messages = await this.smartClient.getMessageHistory(peerId);

    // Convert HTTP messages to our format
    const decoder = new TextDecoder();
    return messages.map((msg) => ({
      id: msg.message_id,
      from: msg.from_peer_id,
      to: peerId,
      content: decoder.decode(msg.content),
      timestamp: msg.timestamp,
      encrypted: true,
      type: 'text' as const,
    }));
  }

  async sendMessage(peerId: string, content: string): Promise<Message> {
    const response = await this.smartClient.sendMessage(peerId, content);

    const message: Message = {
      id: response.message_id,
      from: this.getCurrentUserId(),
      to: peerId,
      content,
      timestamp: response.timestamp,
      encrypted: true,
    };

    // Notify listeners
    this.messageCallbacks.forEach((cb) => cb(message));

    return message;
  }

  async sendGroupMessage(groupId: string, content: string): Promise<Message> {
    const response = await this.smartClient.sendGroupMessage(groupId, content);

    const message: Message = {
      id: response.message_id,
      from: this.getCurrentUserId(),
      to: groupId,
      content,
      timestamp: response.timestamp,
      encrypted: true,
    };

    // Notify listeners
    this.messageCallbacks.forEach((cb) => cb(message));

    return message;
  }

  onMessage(callback: (message: Message) => void): void {
    this.messageCallbacks.push(callback);

    // Subscribe to HTTP polling
    this.smartClient.subscribeToMessages((httpMsg) => {
      const decoder = new TextDecoder();
      const message: Message = {
        id: httpMsg.message_id,
        from: httpMsg.from_peer_id,
        to: this.getCurrentUserId(),
        content: decoder.decode(httpMsg.content),
        timestamp: httpMsg.timestamp,
        encrypted: true,
      };
      callback(message);
    });
  }

  getDebugInfo(): { mode: string; connected: boolean; host: string } {
    return this.smartClient.getDebugInfo();
  }
}

/**
 * HTTP Backend Service (simplified - no fallback needed)
 */
class HttpBackendService implements BackendService {
  private realBackend = new RealBackendService();

  async connect(): Promise<void> {
    console.log('[HttpBackend] Connecting to real backend...');
    await this.realBackend.connect();
    console.log('[HttpBackend] Connected to real backend');
  }

  isConnected(): boolean {
    return this.realBackend.isConnected();
  }

  disconnect(): void {
    this.realBackend.disconnect();
  }

  getCurrentUserId(): string {
    return this.realBackend.getCurrentUserId();
  }

  async getPeers(): Promise<Peer[]> {
    return this.realBackend.getPeers();
  }

  async getPeer(peerId: string): Promise<Peer | null> {
    return this.realBackend.getPeer(peerId);
  }

  async getGroups(): Promise<Group[]> {
    return this.realBackend.getGroups();
  }

  async getGroup(groupId: string): Promise<Group | null> {
    return this.realBackend.getGroup(groupId);
  }

  async createGroup(
    name: string,
    description: string,
    memberIds: string[]
  ): Promise<Group> {
    return this.realBackend.createGroup(name, description, memberIds);
  }

  async getAllConversations(): Promise<Array<Peer | Group>> {
    return this.realBackend.getAllConversations();
  }

  isGroup(id: string): boolean {
    return this.realBackend.isGroup(id);
  }

  async getMessagesWithPeer(peerId: string): Promise<Message[]> {
    return this.realBackend.getMessagesWithPeer(peerId);
  }

  async sendMessage(peerId: string, content: string): Promise<Message> {
    return this.realBackend.sendMessage(peerId, content);
  }

  async sendGroupMessage(groupId: string, content: string): Promise<Message> {
    return this.realBackend.sendGroupMessage(groupId, content);
  }

  onMessage(callback: (message: Message) => void): void {
    this.realBackend.onMessage(callback);
  }

  getDebugInfo(): { mode: string; connected: boolean; host: string } {
    return this.realBackend.getDebugInfo();
  }
}

// Singleton instance
let backendServiceInstance: HttpBackendService | null = null;

/**
 * Get the backend service instance
 */
export function getBackendService(): BackendService {
  if (!backendServiceInstance) {
    backendServiceInstance = new HttpBackendService();
  }
  return backendServiceInstance;
}

/**
 * Initialize backend service
 */
export async function initializeBackend(): Promise<void> {
  const service = getBackendService();
  await service.connect();
}
