import React, { useEffect, useState } from 'react';
import { User } from '../types';
import { api, setTokens } from '../lib/api';
import { AuthContext } from './auth-context';

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem('automata_access_token');
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

    const expired = () => setUser(null);
    window.addEventListener('automata:session-expired', expired);
    void initAuth();
    return () => window.removeEventListener('automata:session-expired', expired);
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
      await api.auth.logout();
    } catch {
      // ignore
    } finally {
      localStorage.removeItem('automata_demo_mode');
      setUser(null);
    }
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
        error,
        clearError: () => setError(null),
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
