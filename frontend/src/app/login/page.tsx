"use client";

import { useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { ShieldAlert, LogIn, Lock } from 'lucide-react';
import Link from 'next/link';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const res = await fetch(`http://${window.location.hostname}:8080/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      const data = await res.json();

      if (!res.ok) {
        throw new Error(data.error || 'Login failed');
      }

      const jwt = data.token;

      // Fetch user profile with the new token
      const meRes = await fetch(`http://${window.location.hostname}:8080/me`, {
        headers: { 'Authorization': `Bearer ${jwt}` }
      });

      if (!meRes.ok) {
        throw new Error('Failed to fetch user profile');
      }

      const user = await meRes.json();
      login(jwt, user);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('An unknown error occurred');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="card" style={{ maxWidth: '400px', width: '100%', margin: '2rem' }}>
        <div className="flex flex-col items-center gap-4" style={{ marginBottom: '2rem' }}>
          <div className="flex items-center justify-center" style={{ 
            width: '64px', 
            height: '64px', 
            borderRadius: '16px',
            background: 'rgba(0, 212, 170, 0.1)',
            border: '1px solid rgba(0, 212, 170, 0.3)'
          }}>
            <Lock className="text-accent" size={32} />
          </div>
          <h1 style={{ fontSize: '1.5rem', margin: 0 }}>TraceLock SOC</h1>
          <p className="text-secondary text-sm">Authorized personnel only</p>
        </div>

        {error && (
          <div className="flex items-center gap-2 p-3" style={{ 
            background: 'rgba(255, 77, 106, 0.1)', 
            border: '1px solid var(--danger-primary)', 
            borderRadius: '8px',
            marginBottom: '1rem'
          }}>
            <ShieldAlert size={18} className="text-danger" />
            <span className="text-danger text-sm">{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="form-group">
            <label className="form-label" htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              className="input mono"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="operator@tracelock.local"
              required
            />
          </div>
          
          <div className="form-group">
            <label className="form-label" htmlFor="password">Password</label>
            <input
              id="password"
              type="password"
              className="input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
            />
          </div>

          <button 
            type="submit" 
            className="btn btn-primary mt-4"
            disabled={loading}
            style={{ width: '100%' }}
          >
            <LogIn size={18} style={{ marginRight: '8px' }} />
            {loading ? 'Authenticating...' : 'Secure Login'}
          </button>
        </form>

        <div style={{ textAlign: 'center', marginTop: '1.5rem' }}>
          <Link href="/bootstrap" className="text-secondary text-sm" style={{ textDecoration: 'none' }}>
            First time? <span className="text-accent">Initialize system</span>
          </Link>
        </div>
      </div>
    </div>
  );
}
