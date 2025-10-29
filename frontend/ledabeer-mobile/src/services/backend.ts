/**
 * Backend Service
 *
 * Unified service layer that can use either real gRPC-Web backend or mock backend
 * Automatically falls back to mock if backend is unavailable
 */

import { getRealGrpcWebClient } from './grpcWeb/realClient';
import {
  mockBackend,
  Message as MockMessage,
  Peer,
  Group,
} from './mockBackend';

// Configuration
const USE_REAL_BACKEND = true; // Set to false to force mock backend
const BACKEND_TIMEOUT = 3000; // 3 seconds timeout for connection

export type { Message as MockMessage, Peer, Group } from './mockBackend';

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
  getMessagesWithPeer(peerId: string): Promise<MockMessage[]>;
  sendMessage(peerId: string, content: string): Promise<MockMessage>;
  sendGroupMessage(groupId: string, content: string): Promise<MockMessage>;
  onMessage(callback: (message: MockMessage) => void): void;
}

/**
 * Real gRPC-Web Backend Implementation
 */
class RealBackendService implements BackendService {
  private grpcClient = getRealGrpcWebClient();
  private messageCallbacks: Array<(message: MockMessage) => void> = [];

  async connect(): Promise<void> {
    return this.grpcClient.connect();
  }

  isConnected(): boolean {
    return this.grpcClient.isConnected();
  }

  disconnect(): void {
    this.grpcClient.disconnect();
    this.messageCallbacks = [];
  }

  getCurrentUserId(): string {
    // For now, use a fixed peer ID
    // In real implementation, this would come from authentication
    return 'peer_local';
  }

  async getPeers(): Promise<Peer[]> {
    try {
      console.log('[RealBackend] Getting peers from backend');
      console.log('[RealBackend] USE_REAL_BACKEND:', USE_REAL_BACKEND);
      const grpcClient = getRealGrpcWebClient();
      console.log('[RealBackend] gRPC client created:', !!grpcClient);
      const peers = await grpcClient.getPeers();
      console.log('[RealBackend] Raw peers from gRPC:', peers);

      // Convert gRPC Peer format to our Peer format
      const convertedPeers: Peer[] = peers.map((peer) => ({
        id: peer.id,
        name: peer.name,
        publicKey: peer.publicKey || '', // Add publicKey field
        online: peer.online,
        lastSeen: peer.lastSeen, // Keep as number (Unix timestamp)
        addresses: peer.addresses,
      }));

      console.log(
        `[RealBackend] Got ${convertedPeers.length} peers from backend`
      );
      return convertedPeers;
    } catch (error) {
      console.error('[RealBackend] Failed to get peers from backend:', error);
      console.log('[RealBackend] Falling back to mock data');
      return mockBackend.getPeers();
    }
  }

  async getPeer(peerId: string): Promise<Peer | null> {
    try {
      console.log(`[RealBackend] Getting peer ${peerId} from backend`);
      const grpcClient = getRealGrpcWebClient();
      const peer = await grpcClient.getPeer(peerId);

      if (!peer) {
        console.log(`[RealBackend] Peer ${peerId} not found in backend`);
        return null;
      }

      // Convert gRPC Peer format to our Peer format
      const convertedPeer: Peer = {
        id: peer.id,
        name: peer.name,
        online: peer.online,
        lastSeen: new Date(peer.lastSeen * 1000), // Convert from Unix timestamp
        addresses: peer.addresses,
      };

      console.log(`[RealBackend] Got peer ${peerId} from backend`);
      return convertedPeer;
    } catch (error) {
      console.error(
        `[RealBackend] Failed to get peer ${peerId} from backend:`,
        error
      );
      console.log('[RealBackend] Falling back to mock data');
      const peers = await mockBackend.getPeers();
      return peers.find((p) => p.id === peerId) || null;
    }
  }

  async getGroups(): Promise<Group[]> {
    // TODO: Implement real group fetching via backend
    console.log('[RealBackend] Group fetching not yet implemented, using mock');
    return mockBackend.getGroups();
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
    console.log('[RealBackend] Group creation not yet implemented, using mock');
    return mockBackend.createGroup(name, description, memberIds);
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

  async getMessagesWithPeer(peerId: string): Promise<MockMessage[]> {
    try {
      const messages = await this.grpcClient.getMessageHistory(peerId);

      // Convert gRPC messages to our format
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
    } catch (error) {
      console.error('[RealBackend] Failed to get messages, using mock:', error);
      return mockBackend.getMessagesWithPeer(peerId);
    }
  }

  async sendMessage(peerId: string, content: string): Promise<MockMessage> {
    try {
      const response = await this.grpcClient.sendMessage(peerId, content);

      const message: MockMessage = {
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
    } catch (error) {
      console.error('[RealBackend] Failed to send message, using mock:', error);
      return mockBackend.sendMessage(peerId, content);
    }
  }

  async sendGroupMessage(
    groupId: string,
    content: string
  ): Promise<MockMessage> {
    try {
      const response = await this.grpcClient.sendGroupMessage(groupId, content);

      const message: MockMessage = {
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
    } catch (error) {
      console.error(
        '[RealBackend] Failed to send group message, using mock:',
        error
      );
      return mockBackend.sendGroupMessage(groupId, content);
    }
  }

  onMessage(callback: (message: MockMessage) => void): void {
    this.messageCallbacks.push(callback);

    // Subscribe to gRPC stream
    this.grpcClient.subscribeToMessages((grpcMsg) => {
      const decoder = new TextDecoder();
      const message: MockMessage = {
        id: grpcMsg.message_id,
        from: grpcMsg.from_peer_id,
        to: this.getCurrentUserId(),
        content: decoder.decode(grpcMsg.content),
        timestamp: grpcMsg.timestamp,
        encrypted: true,
      };
      callback(message);
    });
  }
}

/**
 * Mock Backend Service (wraps mockBackend)
 */
class MockBackendService implements BackendService {
  async connect(): Promise<void> {
    console.log('[MockBackend] Connected (mock)');
  }

  isConnected(): boolean {
    return true;
  }

  disconnect(): void {
    console.log('[MockBackend] Disconnected (mock)');
  }

  getCurrentUserId(): string {
    return mockBackend.getCurrentUserId();
  }

  async getPeers(): Promise<Peer[]> {
    return mockBackend.getPeers();
  }

  async getPeer(peerId: string): Promise<Peer | null> {
    const peers = await this.getPeers();
    return peers.find((p) => p.id === peerId) || null;
  }

  async getGroups(): Promise<Group[]> {
    return mockBackend.getGroups();
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
    return mockBackend.createGroup(name, description, memberIds);
  }

  async getAllConversations(): Promise<Array<Peer | Group>> {
    return mockBackend.getAllConversations();
  }

  isGroup(id: string): boolean {
    return mockBackend.isGroup(id);
  }

  async getMessagesWithPeer(peerId: string): Promise<MockMessage[]> {
    return mockBackend.getMessagesWithPeer(peerId);
  }

  async sendMessage(peerId: string, content: string): Promise<MockMessage> {
    return mockBackend.sendMessage(peerId, content);
  }

  async sendGroupMessage(
    groupId: string,
    content: string
  ): Promise<MockMessage> {
    return mockBackend.sendGroupMessage(groupId, content);
  }

  onMessage(callback: (message: MockMessage) => void): void {
    mockBackend.onMessage(callback);
  }
}

/**
 * Auto-selecting backend service
 */
class AutoBackendService implements BackendService {
  private realBackend = new RealBackendService();
  private mockBackend = new MockBackendService();
  private currentBackend: BackendService = this.mockBackend;
  private connectionAttempted = false;

  async connect(): Promise<void> {
    if (this.connectionAttempted) {
      return;
    }

    this.connectionAttempted = true;

    if (!USE_REAL_BACKEND) {
      console.log('[AutoBackend] Using mock backend (configured)');
      this.currentBackend = this.mockBackend;
      return;
    }

    try {
      console.log('[AutoBackend] Attempting to connect to real backend...');

      // Try to connect with timeout
      const timeoutPromise = new Promise<never>((_, reject) =>
        setTimeout(
          () => reject(new Error('Connection timeout')),
          BACKEND_TIMEOUT
        )
      );

      await Promise.race([this.realBackend.connect(), timeoutPromise]);

      if (this.realBackend.isConnected()) {
        console.log('[AutoBackend] Connected to real backend');
        this.currentBackend = this.realBackend;
      } else {
        throw new Error('Backend not connected');
      }
    } catch (error) {
      console.warn(
        '[AutoBackend] Failed to connect to real backend, using mock:',
        error
      );
      this.currentBackend = this.mockBackend;
      await this.mockBackend.connect();
    }
  }

  isConnected(): boolean {
    return this.currentBackend.isConnected();
  }

  disconnect(): void {
    this.currentBackend.disconnect();
  }

  getCurrentUserId(): string {
    return this.currentBackend.getCurrentUserId();
  }

  async getPeers(): Promise<Peer[]> {
    return this.currentBackend.getPeers();
  }

  async getPeer(peerId: string): Promise<Peer | null> {
    return this.currentBackend.getPeer(peerId);
  }

  async getGroups(): Promise<Group[]> {
    return this.currentBackend.getGroups();
  }

  async getGroup(groupId: string): Promise<Group | null> {
    return this.currentBackend.getGroup(groupId);
  }

  async createGroup(
    name: string,
    description: string,
    memberIds: string[]
  ): Promise<Group> {
    return this.currentBackend.createGroup(name, description, memberIds);
  }

  async getAllConversations(): Promise<Array<Peer | Group>> {
    return this.currentBackend.getAllConversations();
  }

  isGroup(id: string): boolean {
    return this.currentBackend.isGroup(id);
  }

  async getMessagesWithPeer(peerId: string): Promise<MockMessage[]> {
    return this.currentBackend.getMessagesWithPeer(peerId);
  }

  async sendMessage(peerId: string, content: string): Promise<MockMessage> {
    return this.currentBackend.sendMessage(peerId, content);
  }

  async sendGroupMessage(
    groupId: string,
    content: string
  ): Promise<MockMessage> {
    return this.currentBackend.sendGroupMessage(groupId, content);
  }

  onMessage(callback: (message: MockMessage) => void): void {
    this.currentBackend.onMessage(callback);
  }
}

// Singleton instance
let backendServiceInstance: AutoBackendService | null = null;

/**
 * Get the backend service instance
 */
export function getBackendService(): BackendService {
  if (!backendServiceInstance) {
    backendServiceInstance = new AutoBackendService();
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
