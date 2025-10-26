/**
 * Mock Backend Service
 *
 * Simulates the Go backend E2EE functionality for frontend development.
 * This allows us to build and test the UI without requiring the actual backend.
 */

export interface Message {
  id: string;
  from: string;
  to: string;
  content: string;
  timestamp: number;
  encrypted: boolean;
}

export interface Peer {
  id: string;
  name: string;
  publicKey: string;
  online: boolean;
  lastSeen: number;
}

export interface Group {
  id: string;
  name: string;
  description: string;
  members: string[];
  admins: string[];
  createdAt: number;
}

class MockBackendService {
  private peers: Map<string, Peer> = new Map();
  private groups: Map<string, Group> = new Map();
  private messages: Message[] = [];
  private currentUserId: string = 'user-1';
  private messageListeners: Array<(message: Message) => void> = [];

  constructor() {
    this.initializeMockData();
  }

  private initializeMockData() {
    // Add some mock peers
    this.peers.set('peer-1', {
      id: 'peer-1',
      name: 'Alice',
      publicKey: 'mock-public-key-alice',
      online: true,
      lastSeen: Date.now(),
    });

    this.peers.set('peer-2', {
      id: 'peer-2',
      name: 'Bob',
      publicKey: 'mock-public-key-bob',
      online: false,
      lastSeen: Date.now() - 3600000, // 1 hour ago
    });

    this.peers.set('peer-3', {
      id: 'peer-3',
      name: 'Charlie',
      publicKey: 'mock-public-key-charlie',
      online: true,
      lastSeen: Date.now(),
    });

    // Add some mock groups
    this.groups.set('group-1', {
      id: 'group-1',
      name: 'Team Alpha',
      description: 'Project discussion group',
      members: [this.currentUserId, 'peer-1', 'peer-2'],
      admins: [this.currentUserId],
      createdAt: Date.now() - 86400000, // 1 day ago
    });

    this.groups.set('group-2', {
      id: 'group-2',
      name: 'Coffee Lovers',
      description: 'Casual chat',
      members: [this.currentUserId, 'peer-1', 'peer-3'],
      admins: [this.currentUserId, 'peer-1'],
      createdAt: Date.now() - 172800000, // 2 days ago
    });

    // Add some mock messages
    this.messages = [
      {
        id: 'msg-1',
        from: 'peer-1',
        to: this.currentUserId,
        content: 'Hey! How are you?',
        timestamp: Date.now() - 300000,
        encrypted: true,
      },
      {
        id: 'msg-2',
        from: this.currentUserId,
        to: 'peer-1',
        content: 'I\'m good, thanks! How about you?',
        timestamp: Date.now() - 240000,
        encrypted: true,
      },
      {
        id: 'msg-3',
        from: 'peer-1',
        to: this.currentUserId,
        content: 'Great! Want to grab coffee later?',
        timestamp: Date.now() - 180000,
        encrypted: true,
      },
    ];
  }

  /**
   * Simulate X3DH key exchange
   */
  async initiateKeyExchange(peerId: string): Promise<{ success: boolean; sharedSecret?: string }> {
    await this.simulateNetworkDelay();

    const peer = this.peers.get(peerId);
    if (!peer) {
      return { success: false };
    }

    // Simulate successful key exchange
    const sharedSecret = `shared-secret-${this.currentUserId}-${peerId}`;
    return { success: true, sharedSecret };
  }

  /**
   * Send an encrypted message
   */
  async sendMessage(peerId: string, content: string): Promise<Message> {
    await this.simulateNetworkDelay();

    const message: Message = {
      id: `msg-${Date.now()}`,
      from: this.currentUserId,
      to: peerId,
      content,
      timestamp: Date.now(),
      encrypted: true,
    };

    this.messages.push(message);
    return message;
  }

  /**
   * Get messages for a specific peer
   */
  async getMessagesWithPeer(peerId: string): Promise<Message[]> {
    await this.simulateNetworkDelay();

    return this.messages.filter(
      (msg) =>
        (msg.from === peerId && msg.to === this.currentUserId) ||
        (msg.from === this.currentUserId && msg.to === peerId)
    );
  }

  /**
   * Get all peers
   */
  async getPeers(): Promise<Peer[]> {
    await this.simulateNetworkDelay();
    return Array.from(this.peers.values());
  }

  /**
   * Get a specific peer
   */
  async getPeer(peerId: string): Promise<Peer | undefined> {
    await this.simulateNetworkDelay();
    return this.peers.get(peerId);
  }

  /**
   * Subscribe to incoming messages
   */
  onMessage(callback: (message: Message) => void) {
    this.messageListeners.push(callback);

    // Simulate receiving messages occasionally
    this.startMessageSimulation();

    return () => {
      const index = this.messageListeners.indexOf(callback);
      if (index > -1) {
        this.messageListeners.splice(index, 1);
      }
    };
  }

  /**
   * Get current user ID
   */
  getCurrentUserId(): string {
    return this.currentUserId;
  }

  /**
   * Get all groups
   */
  async getGroups(): Promise<Group[]> {
    await this.simulateNetworkDelay();
    return Array.from(this.groups.values());
  }

  /**
   * Get a specific group
   */
  async getGroup(groupId: string): Promise<Group | undefined> {
    await this.simulateNetworkDelay();
    return this.groups.get(groupId);
  }

  /**
   * Create a new group
   */
  async createGroup(name: string, description: string, memberIds: string[]): Promise<Group> {
    await this.simulateNetworkDelay();

    const group: Group = {
      id: `group-${Date.now()}`,
      name,
      description,
      members: [this.currentUserId, ...memberIds],
      admins: [this.currentUserId],
      createdAt: Date.now(),
    };

    this.groups.set(group.id, group);
    return group;
  }

  /**
   * Get all conversations (peers + groups)
   */
  async getAllConversations(): Promise<Array<Peer | Group>> {
    await this.simulateNetworkDelay();
    const peers = Array.from(this.peers.values());
    const groups = Array.from(this.groups.values());
    return [...peers, ...groups];
  }

  /**
   * Check if ID is a group
   */
  isGroup(id: string): boolean {
    return id.startsWith('group-');
  }

  /**
   * Send a group message
   */
  async sendGroupMessage(groupId: string, content: string): Promise<Message> {
    await this.simulateNetworkDelay();

    const message: Message = {
      id: `msg-${Date.now()}`,
      from: this.currentUserId,
      to: groupId,
      content,
      timestamp: Date.now(),
      encrypted: true,
    };

    this.messages.push(message);
    return message;
  }

  private startMessageSimulation() {
    // Simulate receiving a message from Alice every 30 seconds
    if (this.messageListeners.length === 1) {
      setInterval(() => {
        const mockMessage: Message = {
          id: `msg-${Date.now()}`,
          from: 'peer-1',
          to: this.currentUserId,
          content: `Mock message at ${new Date().toLocaleTimeString()}`,
          timestamp: Date.now(),
          encrypted: true,
        };

        this.messages.push(mockMessage);
        this.messageListeners.forEach((listener) => listener(mockMessage));
      }, 30000);
    }
  }

  private async simulateNetworkDelay() {
    const delay = Math.random() * 200 + 100; // 100-300ms
    return new Promise((resolve) => setTimeout(resolve, delay));
  }
}

// Singleton instance
export const mockBackend = new MockBackendService();
