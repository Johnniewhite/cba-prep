'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import api from '@/lib/api';
import wsClient from '@/lib/websocket';

interface User {
  id: string;
  email: string;
  username: string;
  first_name: string;
  last_name: string;
  avatar?: string;
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email_or_username: string, password: string) => Promise<void>;
  register: (data: {
    email: string;
    username: string;
    password: string;
    first_name: string;
    last_name: string;
  }) => Promise<void>;
  logout: () => Promise<void>;
  updateUser: (data: Partial<User>) => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    try {
      const token = localStorage.getItem('access_token');
      if (token) {
        api.setToken(token);
        const userData = await api.getCurrentUser();
        setUser(userData);
        wsClient.connect(token);
      }
    } catch (error) {
      console.error('Auth check failed:', error);
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
    } finally {
      setLoading(false);
    }
  };

  const login = async (email_or_username: string, password: string) => {
    const response = await api.login(email_or_username, password);
    setUser(response.user);
    if (response.access_token) {
      wsClient.connect(response.access_token);
    }
  };

  const register = async (data: {
    email: string;
    username: string;
    password: string;
    first_name: string;
    last_name: string;
  }) => {
    const response = await api.register(data);
    setUser(response.user);
    if (response.access_token) {
      wsClient.connect(response.access_token);
    }
  };

  const logout = async () => {
    await api.logout();
    setUser(null);
    wsClient.disconnect();
  };

  const updateUser = async (data: Partial<User>) => {
    const updatedUser = await api.updateCurrentUser(data);
    setUser(updatedUser);
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout, updateUser }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}