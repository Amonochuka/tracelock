"use client";

import React, { createContext, useContext, useState, useEffect, useRef, useCallback } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { API_URL } from '@/lib/api';

interface User {
  id: number;
  email: string;
  name: string;
  role: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (token: string, refreshToken: string, user: User) => void;
  logout: () => void;
  refreshAccessToken: () => Promise<void>;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Read the `exp` claim from an HS256 JWT payload without verifying it —
// the backend still enforces the token, this only schedules a proactive
// refresh slightly before the session would otherwise expire client-side.
function getTokenExpiry(token: string): number | null {
  try {
    const payload = token.split('.')[1];
    const json = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')));
    return typeof json.exp === 'number' ? json.exp * 1000 : null;
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const refreshingRef = useRef(false);
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    // Check for token on mount
    const storedToken = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');
    
    if (storedToken && storedUser) {
      try {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setToken(storedToken);
        setUser(JSON.parse(storedUser));
      } catch {
        // Handle parse error
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        localStorage.removeItem('refresh_token');
      }
    }
    
    setIsLoading(false);
  }, []);

  // Protect routes based on role
  useEffect(() => {
    if (!isLoading) {
      const isAdmin = user?.role === 'admin';
      const isAuthPage = pathname === '/login' || pathname === '/bootstrap' || pathname === '/admin/login';
      const isAdminRoute = pathname.startsWith('/admin');
      const isAdminLogin = pathname === '/admin/login';
      const isDashboardRoute = pathname.startsWith('/dashboard');

      if (!token) {
        // Not logged in — return each portal to its own door so that
        // signing out from /admin/* (or typing an admin URL while logged
        // out) lands on the admin sign-in, never the personnel one.
        if (isAdminRoute && !isAdminLogin) {
          router.push('/admin/login');
        } else if (isDashboardRoute) {
          router.push('/login');
        }
      } else {
        // Logged in — redirect away from auth pages
        if (isAuthPage) {
          router.push(isAdmin ? '/admin' : '/dashboard');
        }
        // Non-admin trying to access admin routes
        if (!isAdmin && isAdminRoute) {
          router.push('/dashboard');
        }
      }
    }
  }, [token, user, isLoading, pathname, router]);

  const logout = useCallback(() => {
    // Best-effort: tell the backend to revoke the refresh token so the
    // session cannot be resumed later. Failures here never block sign-out.
    const refreshToken = localStorage.getItem('refresh_token');
    if (refreshToken) {
      fetch(`${API_URL}/logout`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token: refreshToken }),
      }).catch(() => {});
    }
    setToken(null);
    setUser(null);
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    localStorage.removeItem('refresh_token');
    // Sign out returns the user to the portal they were using, so admin
    // sessions land on /admin/login instead of the personnel login.
    router.push(pathname.startsWith('/admin') ? '/admin/login' : '/login');
  }, [pathname, router]);

  // Exchange the stored refresh token for a fresh access token, and re-fetch
  // the profile so role changes made by an administrator propagate to the
  // current session. Runs automatically just before the access token expires.
  const refreshAccessToken = useCallback(async () => {
    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken || refreshingRef.current) return;
    refreshingRef.current = true;
    try {
      const res = await fetch(`${API_URL}/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token: refreshToken }),
      });
      if (!res.ok) throw new Error('refresh failed');
      const data = await res.json();
      const newToken = data.token as string;
      localStorage.setItem('token', newToken);
      setToken(newToken);
      // Sync the stored profile with the latest role state.
      try {
        const profileRes = await fetch(`${API_URL}/me`, {
          headers: { Authorization: `Bearer ${newToken}` },
        });
        if (profileRes.ok) {
          const profile = await profileRes.json();
          localStorage.setItem('user', JSON.stringify(profile));
          setUser(profile);
        }
      } catch {
        // Profile sync is best-effort; the fresh token is already in place.
      }
    } catch {
      // The refresh token is expired, revoked, or the API is unreachable —
      // the session can no longer be extended, so sign out cleanly.
      logout();
    } finally {
      refreshingRef.current = false;
    }
  }, [logout]);

  // Proactively refresh ~60 seconds before the access token expires so API
  // calls never observe a stale access token during normal use.
  useEffect(() => {
    if (!token) return;
    const exp = getTokenExpiry(token);
    if (exp === null) return;
    const delay = Math.max(0, exp - Date.now() - 60 * 1000);
    const timer = setTimeout(() => { refreshAccessToken(); }, delay);
    return () => clearTimeout(timer);
  }, [token, refreshAccessToken]);

  const login = (newToken: string, newRefreshToken: string, newUser: User) => {
    setToken(newToken);
    setUser(newUser);
    localStorage.setItem('token', newToken);
    localStorage.setItem('refresh_token', newRefreshToken);
    localStorage.setItem('user', JSON.stringify(newUser));
    router.push(newUser.role === 'admin' ? '/admin' : '/dashboard');
  };

  return (
    <AuthContext.Provider value={{ user, token, login, logout, refreshAccessToken, isLoading }}>
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
