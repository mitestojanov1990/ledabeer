/**
 * Ledabeer Mobile App
 *
 * Main entry point for the React Native application.
 * Provides E2E encrypted P2P chat with voice/video calling.
 */

import React, { useState } from 'react';
import { StatusBar } from 'expo-status-bar';
import { StyleSheet } from 'react-native';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { ChatListScreen } from './src/screens/ChatListScreen';
import { ChatConversationScreen } from './src/screens/ChatConversationScreen';

export default function App() {
  const [selectedPeerId, setSelectedPeerId] = useState<string | null>(null);

  return (
    <SafeAreaProvider>
      <GestureHandlerRootView style={styles.container}>
        {selectedPeerId ? (
          <ChatConversationScreen
            peerId={selectedPeerId}
            onBack={() => setSelectedPeerId(null)}
          />
        ) : (
          <ChatListScreen onSelectChat={setSelectedPeerId} />
        )}
        <StatusBar style="light" />
      </GestureHandlerRootView>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#111827',
  },
});
