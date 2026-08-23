"use client";

import React, { createContext, useContext, useState, useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';

interface User {
  id: number;
  email: string;
  name: string;
  role: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (token: string, user: User) => void;
  logout: () => void;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
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

  const login = (newToken: string, newUser: User) => {
    setToken(newToken);
    setUser(newUser);
    localStorage.setItem('token', newToken);
    localStorage.setItem('user', JSON.stringify(newUser));
    router.push(newUser.role === 'admin' ? '/admin' : '/dashboard');
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    // Sign out returns the user to the portal they were using, so admin
    // sessions land on /admin/login instead of the personnel login.
    router.push(pathname.startsWith('/admin') ? '/admin/login' : '/login');
  };

  return (
    <AuthContext.Provider value={{ user, token, login, logout, isLoading }}>
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
