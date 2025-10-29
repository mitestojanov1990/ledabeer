import { authService } from './authService';

const API_BASE_URL = 'http://192.168.0.140:9001/api';

export interface Conversation {
  id: string;
  type: 'direct' | 'group';
  name: string;
  participants: ConversationMember[];
  last_message?: MessagePreview;
  created_at: string;
  updated_at: string;
  unread_count: Record<string, number>;
}

export interface ConversationMember {
  user_id: string;
  username: string;
  display_name: string;
  avatar_url: string;
  joined_at: string;
  is_online: boolean;
}

export interface MessagePreview {
  id: string;
  content: string;
  from_user: string;
  from_name: string;
  timestamp: string;
  type: string;
}

export interface UserSearchResult {
  user_id: string;
  username: string;
  display_name: string;
  email: string;
  avatar_url: string;
  is_online: boolean;
}

export interface CreateConversationRequest {
  participant_id: string;
  type?: 'direct' | 'group';
}

export interface CreateConversationResponse {
  conversation: Conversation;
  message: string;
}

export interface GetConversationsResponse {
  conversations: Conversation[];
  total: number;
  has_more: boolean;
}

export interface AddMessageRequest {
  conversation_id: string;
  content: string;
  type?: string;
}

export interface AddMessageResponse {
  message_id: string;
  conversation_id: string;
  timestamp: string;
}

class ConversationService {
  private getAuthHeaders(): Record<string, string> {
    const token = authService.getAccessToken();
    if (token) {
      return {
        'Authorization': `Bearer ${token}`,
        'X-User-ID': authService.getUser()?.id || '', // For testing
      };
    }
    return {};
  }

  // Get all conversations for the current user
  async getConversations(limit = 50, offset = 0): Promise<GetConversationsResponse> {
    try {
      const response = await fetch(
        `${API_BASE_URL}/conversations/list?limit=${limit}&offset=${offset}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            ...this.getAuthHeaders(),
          },
        }
      );

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Failed to get conversations: ${error}`);
      }

      return await response.json();
    } catch (error) {
      console.error('[ConversationService] Get conversations failed:', error);
      throw error;
    }
  }

  // Create a new conversation
  async createConversation(request: CreateConversationRequest): Promise<CreateConversationResponse> {
    try {
      const response = await fetch(`${API_BASE_URL}/conversations`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...this.getAuthHeaders(),
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Failed to create conversation: ${error}`);
      }

      return await response.json();
    } catch (error) {
      console.error('[ConversationService] Create conversation failed:', error);
      throw error;
    }
  }

  // Get a specific conversation
  async getConversation(conversationId: string): Promise<Conversation> {
    try {
      const response = await fetch(
        `${API_BASE_URL}/conversations/get?conversation_id=${conversationId}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            ...this.getAuthHeaders(),
          },
        }
      );

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Failed to get conversation: ${error}`);
      }

      return await response.json();
    } catch (error) {
      console.error('[ConversationService] Get conversation failed:', error);
      throw error;
    }
  }

  // Add a message to a conversation
  async addMessage(request: AddMessageRequest): Promise<AddMessageResponse> {
    try {
      const response = await fetch(`${API_BASE_URL}/conversations/messages`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...this.getAuthHeaders(),
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Failed to add message: ${error}`);
      }

      return await response.json();
    } catch (error) {
      console.error('[ConversationService] Add message failed:', error);
      throw error;
    }
  }

  // Mark messages as read
  async markAsRead(conversationId: string): Promise<void> {
    try {
      const response = await fetch(`${API_BASE_URL}/conversations/mark-read`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...this.getAuthHeaders(),
        },
        body: JSON.stringify({ conversation_id: conversationId }),
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Failed to mark as read: ${error}`);
      }
    } catch (error) {
      console.error('[ConversationService] Mark as read failed:', error);
      throw error;
    }
  }

  // Search for users by email or username
  async searchUsers(query: string, limit = 20): Promise<UserSearchResult[]> {
    try {
      const response = await fetch(`${API_BASE_URL}/users/search`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...this.getAuthHeaders(),
        },
        body: JSON.stringify({ query, limit }),
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Failed to search users: ${error}`);
      }

      const result = await response.json();
      return result.users || [];
    } catch (error) {
      console.error('[ConversationService] Search users failed:', error);
      throw error;
    }
  }

  // Find user by email
  async findUserByEmail(email: string): Promise<UserSearchResult | null> {
    try {
      const response = await fetch(
        `${API_BASE_URL}/users/find-by-email?email=${encodeURIComponent(email)}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            ...this.getAuthHeaders(),
          },
        }
      );

      if (response.status === 404) {
        return null;
      }

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Failed to find user by email: ${error}`);
      }

      return await response.json();
    } catch (error) {
      console.error('[ConversationService] Find user by email failed:', error);
      throw error;
    }
  }

  // Find user by username
  async findUserByUsername(username: string): Promise<UserSearchResult | null> {
    try {
      const response = await fetch(
        `${API_BASE_URL}/users/find-by-username?username=${encodeURIComponent(username)}`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            ...this.getAuthHeaders(),
          },
        }
      );

      if (response.status === 404) {
        return null;
      }

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Failed to find user by username: ${error}`);
      }

      return await response.json();
    } catch (error) {
      console.error('[ConversationService] Find user by username failed:', error);
      throw error;
    }
  }

  // Start a conversation with a user by email or username
  async startConversationWithUser(identifier: string): Promise<Conversation> {
    try {
      // First try to find by email
      let user = await this.findUserByEmail(identifier);
      
      // If not found by email, try username
      if (!user) {
        user = await this.findUserByUsername(identifier);
      }

      if (!user) {
        throw new Error('User not found');
      }

      // Create conversation with the found user
      const response = await this.createConversation({
        participant_id: user.user_id,
        type: 'direct',
      });

      return response.conversation;
    } catch (error) {
      console.error('[ConversationService] Start conversation failed:', error);
      throw error;
    }
  }
}

export const conversationService = new ConversationService();
