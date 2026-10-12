import React, { useEffect, useState } from 'react';
import { useAuth as useClerkAuth, useClerk } from '@clerk/react';
import { User } from '../types';
import { api, setClerkTokenGetter, setTokens } from '../lib/api';
import { AuthContext } from './auth-context';

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isLoaded, isSignedIn, sessionId, getToken } = useClerkAuth();
  const { signOut } = useClerk();
  const [account, setAccount] = useState<{ sessionId: string; user: User | null; error: string | null } | null>(null);

  useEffect(() => {
    let cancelled = false;
    setClerkTokenGetter(getToken);
    setTokens(null);
    if (!isLoaded || !isSignedIn || !sessionId) return;
    api.auth.me().then(user => {
      if (!cancelled) setAccount({ sessionId, user, error: null });
    }).catch(() => {
      if (!cancelled) setAccount({ sessionId, user: null, error: 'You are signed in, but Automata could not load your account. Please try again or contact your administrator.' });
    });
    return () => { cancelled = true; };
  }, [isLoaded, isSignedIn, sessionId, getToken]);

  const current = isSignedIn && account?.sessionId === sessionId ? account : null;
  return <AuthContext.Provider value={{
    user: current?.user ?? null,
    isAuthenticated: !!current?.user,
    isLoading: !isLoaded || (!!isSignedIn && !current),
    error: current?.error ?? null,
    clearError: () => setAccount(previous => previous ? { ...previous, error: null } : null),
    logout: async () => { await signOut(); },
  }}>{children}</AuthContext.Provider>;
};
