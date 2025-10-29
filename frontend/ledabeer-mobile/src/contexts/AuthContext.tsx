/**
 * Authentication Context
 *
 * Provides authentication state and methods throughout the app
 */

import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from 'react';
import { authService, User } from '../services/authService';

interface AuthContextType {
  isAuthenticated: boolean;
  user: User | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  register: (
    username: string,
    email: string,
    password: string,
    displayName: string
  ) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if user is already authenticated
    const checkAuth = async () => {
      try {
        if (authService.isAuthenticated()) {
          const currentUser = await authService.getCurrentUser();
          setUser(currentUser);
          setIsAuthenticated(true);
        }
      } catch (error) {
        console.error('[AuthContext] Auth check failed:', error);
        // Clear auth data if check fails
        await authService.logout();
      } finally {
        setLoading(false);
      }
    };

    checkAuth();

    // Listen for auth state changes
    const unsubscribe = authService.addAuthListener((authenticated) => {
      setIsAuthenticated(authenticated);
      if (authenticated) {
        setUser(authService.getCurrentUserSync());
      } else {
        setUser(null);
      }
    });

    return unsubscribe;
  }, []);

  const login = async (username: string, password: string) => {
    try {
      await authService.login({ username, password });
      // Auth state will be updated via the listener
    } catch (error) {
      throw error;
    }
  };

  const register = async (
    username: string,
    email: string,
    password: string,
    displayName: string
  ) => {
    try {
      await authService.register({
        username,
        email,
        password,
        display_name: displayName,
      });
      // Registration successful, but user needs to login
    } catch (error) {
      throw error;
    }
  };

  const logout = async () => {
    try {
      await authService.logout();
      // Auth state will be updated via the listener
    } catch (error) {
      console.error('[AuthContext] Logout failed:', error);
      // Force logout even if server request fails
      setIsAuthenticated(false);
      setUser(null);
    }
  };

  const value: AuthContextType = {
    isAuthenticated,
    user,
    loading,
    login,
    register,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
