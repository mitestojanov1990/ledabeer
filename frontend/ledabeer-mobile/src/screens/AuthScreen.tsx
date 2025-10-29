/**
 * Authentication Screen
 *
 * Handles both login and registration flows
 */

import React, { useState } from 'react';
import { View, StyleSheet } from 'react-native';
import { LoginScreen } from './LoginScreen';
import { RegisterScreen } from './RegisterScreen';

interface AuthScreenProps {
  onAuthSuccess: () => void;
}

export function AuthScreen({ onAuthSuccess }: AuthScreenProps) {
  const [isLogin, setIsLogin] = useState(true);

  const handleLoginSuccess = () => {
    onAuthSuccess();
  };

  const handleRegisterSuccess = () => {
    // Switch to login after successful registration
    setIsLogin(true);
  };

  const switchToRegister = () => {
    setIsLogin(false);
  };

  const switchToLogin = () => {
    setIsLogin(true);
  };

  return (
    <View style={styles.container}>
      {isLogin ? (
        <LoginScreen
          onLoginSuccess={handleLoginSuccess}
          onSwitchToRegister={switchToRegister}
        />
      ) : (
        <RegisterScreen
          onRegisterSuccess={handleRegisterSuccess}
          onSwitchToLogin={switchToLogin}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#111827',
  },
});
