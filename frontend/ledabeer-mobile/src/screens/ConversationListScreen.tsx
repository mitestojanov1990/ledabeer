import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  TextInput,
  Alert,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import { conversationService, Conversation, UserSearchResult } from '../services/conversationService';
import { useAuth } from '../contexts/AuthContext';
import { useConversationStore, useRealtimeConnection } from '../store/conversationStore';

interface ConversationListScreenProps {
  onSelectConversation: (conversation: Conversation) => void;
}

export const ConversationListScreen: React.FC<ConversationListScreenProps> = ({
  onSelectConversation,
}) => {
  const { user } = useAuth();
  const { conversations, loading, error, setConversations, setLoading, setError } = useConversationStore();
  const { initializeRealtime, cleanupRealtime } = useRealtimeConnection();
  
  const [refreshing, setRefreshing] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<UserSearchResult[]>([]);
  const [showSearch, setShowSearch] = useState(false);
  const [searching, setSearching] = useState(false);

  // Load conversations on mount and initialize realtime
  useEffect(() => {
    loadConversations();
    initializeRealtime();
    
    // Cleanup on unmount
    return () => {
      cleanupRealtime();
    };
  }, []);

  const loadConversations = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await conversationService.getConversations();
      setConversations(response.conversations);
    } catch (error) {
      console.error('Failed to load conversations:', error);
      setError('Failed to load conversations');
      Alert.alert('Error', 'Failed to load conversations');
    } finally {
      setLoading(false);
    }
  };

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await loadConversations();
    setRefreshing(false);
  }, []);

  const handleSearch = async (query: string) => {
    setSearchQuery(query);
    
    if (query.trim().length < 2) {
      setSearchResults([]);
      return;
    }

    try {
      setSearching(true);
      const results = await conversationService.searchUsers(query);
      setSearchResults(results);
    } catch (error) {
      console.error('Search failed:', error);
      Alert.alert('Error', 'Search failed');
    } finally {
      setSearching(false);
    }
  };

  const handleStartConversation = async (user: UserSearchResult) => {
    try {
      const response = await conversationService.createConversation({
        participant_id: user.user_id,
        type: 'direct',
      });
      
      // Add new conversation to store
      setConversations([response.conversation, ...conversations]);
      setShowSearch(false);
      setSearchQuery('');
      setSearchResults([]);
      
      // Navigate to the new conversation
      onSelectConversation(response.conversation);
    } catch (error) {
      console.error('Failed to start conversation:', error);
      Alert.alert('Error', 'Failed to start conversation');
    }
  };

  const renderConversationItem = ({ item }: { item: Conversation }) => {
    const otherParticipant = item.participants.find(p => p.user_id !== user?.id);
    const unreadCount = item.unread_count[user?.id || ''] || 0;

    return (
      <TouchableOpacity
        style={styles.conversationItem}
        onPress={() => onSelectConversation(item)}
      >
        <View style={styles.conversationContent}>
          <View style={styles.conversationHeader}>
            <Text style={styles.conversationName}>
              {item.type === 'direct' ? otherParticipant?.display_name : item.name}
            </Text>
            <Text style={styles.conversationTime}>
              {item.last_message ? formatTime(item.last_message.timestamp) : ''}
            </Text>
          </View>
          
          <View style={styles.conversationPreview}>
            <Text style={styles.lastMessage} numberOfLines={1}>
              {item.last_message ? item.last_message.content : 'No messages yet'}
            </Text>
            {unreadCount > 0 && (
              <View style={styles.unreadBadge}>
                <Text style={styles.unreadCount}>{unreadCount}</Text>
              </View>
            )}
          </View>
          
          <View style={styles.participantInfo}>
            <Text style={styles.participantStatus}>
              {otherParticipant?.is_online ? '🟢 Online' : '⚫ Offline'}
            </Text>
          </View>
        </View>
      </TouchableOpacity>
    );
  };

  const renderSearchResult = ({ item }: { item: UserSearchResult }) => (
    <TouchableOpacity
      style={styles.searchResultItem}
      onPress={() => handleStartConversation(item)}
    >
      <View style={styles.searchResultContent}>
        <Text style={styles.searchResultName}>{item.display_name}</Text>
        <Text style={styles.searchResultUsername}>@{item.username}</Text>
        <Text style={styles.searchResultEmail}>{item.email}</Text>
        <Text style={styles.searchResultStatus}>
          {item.is_online ? '🟢 Online' : '⚫ Offline'}
        </Text>
      </View>
    </TouchableOpacity>
  );

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60);

    if (diffInHours < 24) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } else if (diffInHours < 168) { // 7 days
      return date.toLocaleDateString([], { weekday: 'short' });
    } else {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
    }
  };

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#3B82F6" />
        <Text style={styles.loadingText}>Loading conversations...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Messages</Text>
        <TouchableOpacity
          style={styles.searchButton}
          onPress={() => setShowSearch(!showSearch)}
        >
          <Text style={styles.searchButtonText}>+</Text>
        </TouchableOpacity>
      </View>

      {showSearch && (
        <View style={styles.searchContainer}>
          <TextInput
            style={styles.searchInput}
            placeholder="Search by email or username..."
            value={searchQuery}
            onChangeText={handleSearch}
            autoFocus
          />
          
          {searching && (
            <View style={styles.searchingContainer}>
              <ActivityIndicator size="small" color="#3B82F6" />
              <Text style={styles.searchingText}>Searching...</Text>
            </View>
          )}
          
          {searchResults.length > 0 && (
            <FlatList
              data={searchResults}
              keyExtractor={(item) => item.user_id}
              renderItem={renderSearchResult}
              style={styles.searchResults}
            />
          )}
        </View>
      )}

      <FlatList
        data={conversations}
        keyExtractor={(item) => item.id}
        renderItem={renderConversationItem}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyTitle}>No conversations yet</Text>
            <Text style={styles.emptySubtitle}>
              Start a new conversation by searching for a user
            </Text>
          </View>
        }
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    backgroundColor: '#fff',
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#333',
  },
  searchButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: '#3B82F6',
    justifyContent: 'center',
    alignItems: 'center',
  },
  searchButtonText: {
    color: '#fff',
    fontSize: 20,
    fontWeight: 'bold',
  },
  searchContainer: {
    backgroundColor: '#fff',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0',
  },
  searchInput: {
    height: 40,
    borderWidth: 1,
    borderColor: '#e0e0e0',
    borderRadius: 20,
    paddingHorizontal: 16,
    fontSize: 16,
  },
  searchingContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
  },
  searchingText: {
    marginLeft: 8,
    color: '#666',
  },
  searchResults: {
    maxHeight: 200,
    marginTop: 8,
  },
  searchResultItem: {
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  searchResultContent: {
    flex: 1,
  },
  searchResultName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
  },
  searchResultUsername: {
    fontSize: 14,
    color: '#666',
    marginTop: 2,
  },
  searchResultEmail: {
    fontSize: 14,
    color: '#999',
    marginTop: 2,
  },
  searchResultStatus: {
    fontSize: 12,
    color: '#666',
    marginTop: 4,
  },
  conversationItem: {
    backgroundColor: '#fff',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  conversationContent: {
    flex: 1,
  },
  conversationHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  conversationName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    flex: 1,
  },
  conversationTime: {
    fontSize: 12,
    color: '#999',
  },
  conversationPreview: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  lastMessage: {
    fontSize: 14,
    color: '#666',
    flex: 1,
  },
  unreadBadge: {
    backgroundColor: '#3B82F6',
    borderRadius: 10,
    minWidth: 20,
    height: 20,
    justifyContent: 'center',
    alignItems: 'center',
    marginLeft: 8,
  },
  unreadCount: {
    color: '#fff',
    fontSize: 12,
    fontWeight: 'bold',
  },
  participantInfo: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  participantStatus: {
    fontSize: 12,
    color: '#666',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
    color: '#666',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 32,
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#333',
    marginBottom: 8,
  },
  emptySubtitle: {
    fontSize: 16,
    color: '#666',
    textAlign: 'center',
  },
});
