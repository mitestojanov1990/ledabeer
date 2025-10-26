/**
 * Chat List Screen
 *
 * Displays a list of peers/conversations.
 */

import React, { useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  FlatList,
  ActivityIndicator,
} from 'react-native';
import { useChatStore } from '../store/chatStore';
import { Peer } from '../services/mockBackend';

interface ChatListScreenProps {
  onSelectChat: (peerId: string) => void;
}

export function ChatListScreen({ onSelectChat }: ChatListScreenProps) {
  const { conversations, messages, loading, loadConversations, initializeMessageListener } = useChatStore();

  useEffect(() => {
    loadConversations();
    initializeMessageListener();
  }, []);

  const getLastMessage = (peerId: string) => {
    const peerMessages = messages[peerId] || [];
    if (peerMessages.length === 0) return null;
    return peerMessages[peerMessages.length - 1];
  };

  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp);
    const now = new Date();
    const isToday = date.toDateString() === now.toDateString();

    if (isToday) {
      return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
    }
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  const renderChatItem = ({ item }: { item: Peer }) => {
    const lastMessage = getLastMessage(item.id);

    return (
      <TouchableOpacity
        style={styles.chatItem}
        onPress={() => onSelectChat(item.id)}
        activeOpacity={0.7}
      >
        <View style={styles.avatar}>
          <Text style={styles.avatarText}>{item.name.charAt(0).toUpperCase()}</Text>
          {item.online && <View style={styles.onlineIndicator} />}
        </View>

        <View style={styles.chatInfo}>
          <View style={styles.chatHeader}>
            <Text style={styles.chatName}>{item.name}</Text>
            {lastMessage && (
              <Text style={styles.timestamp}>{formatTime(lastMessage.timestamp)}</Text>
            )}
          </View>

          {lastMessage ? (
            <View style={styles.messagePreview}>
              <Text style={styles.messageText} numberOfLines={1}>
                {lastMessage.content}
              </Text>
              {lastMessage.encrypted && (
                <Text style={styles.encryptedBadge}>🔒</Text>
              )}
            </View>
          ) : (
            <Text style={styles.noMessages}>No messages yet</Text>
          )}
        </View>
      </TouchableOpacity>
    );
  };

  if (loading && conversations.length === 0) {
    return (
      <View style={styles.centerContainer}>
        <ActivityIndicator size="large" color="#3B82F6" />
        <Text style={styles.loadingText}>Loading conversations...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Ledabeer</Text>
        <Text style={styles.headerSubtitle}>End-to-End Encrypted Chats</Text>
      </View>

      <FlatList
        data={conversations}
        renderItem={renderChatItem}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.listContent}
        ItemSeparatorComponent={() => <View style={styles.separator} />}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#111827',
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#111827',
  },
  loadingText: {
    marginTop: 16,
    color: '#9CA3AF',
    fontSize: 14,
  },
  header: {
    paddingTop: 60,
    paddingBottom: 20,
    paddingHorizontal: 20,
    backgroundColor: '#1F2937',
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  headerTitle: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#3B82F6',
    marginBottom: 4,
  },
  headerSubtitle: {
    fontSize: 14,
    color: '#9CA3AF',
  },
  listContent: {
    paddingVertical: 8,
  },
  chatItem: {
    flexDirection: 'row',
    padding: 16,
    backgroundColor: '#1F2937',
    alignItems: 'center',
  },
  avatar: {
    width: 50,
    height: 50,
    borderRadius: 25,
    backgroundColor: '#3B82F6',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
    position: 'relative',
  },
  avatarText: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#FFFFFF',
  },
  onlineIndicator: {
    position: 'absolute',
    bottom: 2,
    right: 2,
    width: 12,
    height: 12,
    borderRadius: 6,
    backgroundColor: '#10B981',
    borderWidth: 2,
    borderColor: '#1F2937',
  },
  chatInfo: {
    flex: 1,
  },
  chatHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  chatName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#F9FAFB',
  },
  timestamp: {
    fontSize: 12,
    color: '#6B7280',
  },
  messagePreview: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  messageText: {
    flex: 1,
    fontSize: 14,
    color: '#9CA3AF',
  },
  encryptedBadge: {
    fontSize: 12,
    marginLeft: 8,
  },
  noMessages: {
    fontSize: 14,
    color: '#6B7280',
    fontStyle: 'italic',
  },
  separator: {
    height: 1,
    backgroundColor: '#374151',
    marginLeft: 78,
  },
});
