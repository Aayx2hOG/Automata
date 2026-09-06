import React, { createContext, useContext, useEffect, useState } from 'react';
import { User } from '../types';
import { api, setTokens } from '../lib/api';

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, pass: string) => Promise<void>;
  register: (email: string, pass: string) => Promise<void>;
  logout: () => Promise<void>;
  loginAsDemo: () => void;
  error: string | null;
  clearError: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const DEMO_USER: User = {
  id: '00000000-0000-0000-0000-000000000001',
  email: 'demo@automata.io',
  created_at: new Date().toISOString(),
};

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem('automata_access_token');
      const isDemo = localStorage.getItem('automata_demo_mode');

      if (isDemo === 'true') {
        setUser(DEMO_USER);
        setIsLoading(false);
        return;
      }

      if (token) {
        try {
          const userData = await api.auth.me();
          setUser(userData);
        } catch (err: any) {
          console.warn('Failed to load active user from token', err);
          setTokens(null);
        }
      }
      setIsLoading(false);
    };

    initAuth();
  }, []);

  const login = async (email: string, pass: string) => {
    setError(null);
    try {
      localStorage.removeItem('automata_demo_mode');
      await api.auth.login(email, pass);
      const userData = await api.auth.me();
      setUser(userData);
    } catch (err: any) {
      setError(err.message || 'Login failed');
      throw err;
    }
  };

  const register = async (email: string, pass: string) => {
    setError(null);
    try {
      localStorage.removeItem('automata_demo_mode');
      const data = await api.auth.register(email, pass);
      setUser(data.user);
    } catch (err: any) {
      setError(err.message || 'Registration failed');
      throw err;
    }
  };

  const logout = async () => {
    try {
      if (localStorage.getItem('automata_demo_mode') !== 'true') {
        await api.auth.logout();
      }
    } catch (e) {
      // ignore
    } finally {
      localStorage.removeItem('automata_demo_mode');
      setUser(null);
    }
  };

  const loginAsDemo = () => {
    localStorage.setItem('automata_demo_mode', 'true');
    setUser(DEMO_USER);
    setError(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        register,
        logout,
        loginAsDemo,
        error,
        clearError: () => setError(null),
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within an AuthProvider');
  return ctx;
};
