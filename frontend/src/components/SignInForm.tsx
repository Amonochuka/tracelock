"use client";

import { useState } from 'react';
import { ShieldAlert, LogIn, Lock, Shield } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { API_URL } from '@/lib/api';

type Portal = 'admin' | 'user';

interface SignInFormProps {
  portal: Portal;
}

export default function SignInForm({ portal }: SignInFormProps) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const isAdminPortal = portal === 'admin';

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await fetch(`${API_URL}/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Unable to sign in');
      }

      const profileResponse = await fetch(`${API_URL}/me`, {
        headers: { Authorization: `Bearer ${data.token}` },
      });

      if (!profileResponse.ok) {
        throw new Error('Unable to verify this account');
      }

      const user = await profileResponse.json();
      if ((isAdminPortal && user.role !== 'admin') || (!isAdminPortal && user.role === 'admin')) {
        throw new Error('This account does not have access to this portal');
      }

      login(data.token, data.refresh_token, user);
    } catch (caughtError: unknown) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to sign in');
    } finally {
      setLoading(false);
    }
  };

  const title = isAdminPortal ? 'TraceLock Administration' : 'TraceLock Access Portal';
  const subtitle = isAdminPortal ? 'Restricted system administration' : 'Authorized personnel only';
  const Icon = isAdminPortal ? Shield : Lock;

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="card" style={{ maxWidth: '400px', width: '100%', margin: '2rem' }}>
        <div className="flex flex-col items-center gap-4" style={{ marginBottom: '2rem' }}>
          <div className="flex items-center justify-center" style={{ width: '64px', height: '64px', borderRadius: '16px', background: 'rgba(0, 212, 170, 0.1)', border: '1px solid rgba(0, 212, 170, 0.3)' }}>
            <Icon className="text-accent" size={32} />
          </div>
          <h1 style={{ fontSize: '1.5rem', margin: 0 }}>{title}</h1>
          <p className="text-secondary text-sm">{subtitle}</p>
        </div>

        {error && <div className="flex items-center gap-2 p-3" style={{ background: 'rgba(255, 77, 106, 0.1)', border: '1px solid var(--danger-primary)', borderRadius: '8px', marginBottom: '1rem' }}><ShieldAlert size={18} className="text-danger" /><span className="text-danger text-sm">{error}</span></div>}

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="form-group">
            <label className="form-label" htmlFor="email">Email</label>
            <input id="email" type="email" className="input mono" value={email} onChange={(event) => setEmail(event.target.value)} placeholder="operator@tracelock.local" required disabled={loading} />
          </div>
          <div className="form-group">
            <label className="form-label" htmlFor="password">Password</label>
            <input id="password" type="password" className="input" value={password} onChange={(event) => setPassword(event.target.value)} placeholder="••••••••" required disabled={loading} />
          </div>
          <button type="submit" className="btn btn-primary mt-4" disabled={loading} style={{ width: '100%' }}>
            {loading ? (
              <><span className="spinner"></span> Authenticating...</>
            ) : (
              <><LogIn size={18} style={{ marginRight: '8px' }} /> Secure Login</>
            )}
          </button>
        </form>
      </div>
    </div>
  );
}
