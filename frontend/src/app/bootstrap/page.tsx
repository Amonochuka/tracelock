"use client";

import { useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { ShieldCheck, UserPlus } from 'lucide-react';
import Link from 'next/link';

export default function BootstrapPage() {
  const [name, setName] = useState('');
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
      const res = await fetch(`http://${window.location.hostname}:8080/bootstrap`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, email, password }),
      });

      const data = await res.json();

      if (!res.ok) {
        // Backend returns 404 "not found" when admin already exists (security: no info disclosure)
        if (res.status === 404) {
          throw new Error('System has already been initialized');
        }
        throw new Error(data.error || 'Bootstrap failed');
      }

      // Step 2: Log in with the new credentials
      const loginRes = await fetch(`http://${window.location.hostname}:8080/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      const loginData = await loginRes.json();

      if (!loginRes.ok) {
        throw new Error(loginData.error || 'Login after bootstrap failed');
      }

      const jwt = loginData.token;

      // Step 3: Fetch user profile
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
      <div className="card" style={{ maxWidth: '420px', width: '100%', margin: '2rem' }}>
        <div className="flex flex-col items-center gap-4" style={{ marginBottom: '2rem' }}>
          <div className="flex items-center justify-center" style={{
            width: '64px',
            height: '64px',
            borderRadius: '16px',
            background: 'rgba(0, 212, 170, 0.1)',
            border: '1px solid rgba(0, 212, 170, 0.3)'
          }}>
            <ShieldCheck className="text-accent" size={32} />
          </div>
          <h1 style={{ fontSize: '1.5rem', margin: 0 }}>System Bootstrap</h1>
          <p className="text-secondary text-sm" style={{ textAlign: 'center' }}>
            Create the first admin account to initialize the system.
            This can only be done once.
          </p>
        </div>

        {error && (
          <div className="flex items-center gap-2 p-3" style={{
            background: 'rgba(255, 77, 106, 0.1)',
            border: '1px solid var(--danger-primary)',
            borderRadius: '8px',
            marginBottom: '1rem'
          }}>
            <span className="text-danger text-sm">{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="form-group">
            <label className="form-label" htmlFor="name">Full Name</label>
            <input
              id="name"
              type="text"
              className="input"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="System Administrator"
              required
            />
          </div>

          <div className="form-group">
            <label className="form-label" htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              className="input mono"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="admin@tracelock.local"
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
              minLength={8}
            />
          </div>

          <button
            type="submit"
            className="btn btn-primary mt-4"
            disabled={loading}
            style={{ width: '100%' }}
          >
            <UserPlus size={18} style={{ marginRight: '8px' }} />
            {loading ? 'Initializing...' : 'Initialize System'}
          </button>
        </form>

        <div style={{ textAlign: 'center', marginTop: '1.5rem' }}>
          <Link href="/login" className="text-secondary text-sm" style={{ textDecoration: 'none' }}>
            Already bootstrapped? <span className="text-accent">Sign in</span>
          </Link>
        </div>
      </div>
    </div>
  );
}
