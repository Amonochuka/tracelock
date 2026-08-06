"use client";

import { useEffect, useState, useCallback } from 'react';
import { useAuth } from '@/context/AuthContext';
import { Shield, User, MoreVertical, Plus, UserPlus, X } from 'lucide-react';
import { API_URL } from '@/lib/api';

interface UserObj {
  id: number;
  name: string;
  email: string;
  role: string;
  created_at: string;
}

export default function UsersPage() {
  const { token } = useAuth();
  const [users, setUsers] = useState<UserObj[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Create User State
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newName, setNewName] = useState('');
  const [newEmail, setNewEmail] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [createLoading, setCreateLoading] = useState(false);
  const [createError, setCreateError] = useState('');

  const fetchUsers = useCallback(async () => {
    try {
      const res = await fetch(`${API_URL}/admin/users`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (!res.ok) throw new Error('Failed to fetch users');
      const data = await res.json();
      setUsers(data || []);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('An unknown error occurred');
      }
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (!token) return;
    fetchUsers();
  }, [token, fetchUsers]);

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateError('');
    setCreateLoading(true);

    try {
      const res = await fetch(`${API_URL}/admin/users`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}` 
        },
        body: JSON.stringify({ name: newName, email: newEmail, password: newPassword }),
      });

      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to create user');

      await fetchUsers(); // Refresh list
      
      setShowCreateModal(false);
      setNewName('');
      setNewEmail('');
      setNewPassword('');
    } catch (err: unknown) {
      if (err instanceof Error) {
        setCreateError(err.message);
      } else {
        setCreateError('An unknown error occurred');
      }
    } finally {
      setCreateLoading(false);
    }
  };

  if (loading) return <div className="text-secondary">Loading personnel data...</div>;

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1>Personnel Access</h1>
          <p className="text-secondary mt-2">Manage user identities and roles</p>
        </div>
        <button 
          onClick={() => setShowCreateModal(true)}
          className="btn btn-primary flex items-center gap-2"
        >
          <UserPlus size={18} />
          <span>Add Personnel</span>
        </button>
      </div>

      {error && (
        <div className="p-4 mb-8 bg-red-500/10 border border-red-500 rounded-lg text-red-500">
          {error}
        </div>
      )}

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>OPERATOR</th>
              <th>CONTACT</th>
              <th>CLEARANCE</th>
              <th>ENROLLED</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {users.map(u => (
              <tr key={u.id}>
                <td className="mono text-secondary">#{u.id}</td>
                <td>
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-[var(--bg-main)] border border-[var(--border-color)] flex items-center justify-center">
                      <User size={14} className="text-secondary" />
                    </div>
                    <span className="font-medium">{u.name}</span>
                  </div>
                </td>
                <td className="text-secondary">{u.email}</td>
                <td>
                  <span className={`px-2 py-1 rounded text-xs font-medium uppercase tracking-wider ${
                    u.role === 'admin' 
                      ? 'bg-[rgba(0,212,170,0.1)] text-accent border border-[rgba(0,212,170,0.2)]'
                      : 'bg-[rgba(255,255,255,0.05)] text-secondary'
                  }`}>
                    {u.role === 'admin' && <Shield size={10} className="inline mr-1" />}
                    {u.role}
                  </span>
                </td>
                <td className="text-sm text-secondary mono">
                  {new Date(u.created_at).toLocaleDateString()}
                </td>
                <td className="text-right">
                  <button className="p-2 text-secondary hover:text-white rounded hover:bg-[rgba(255,255,255,0.05)] transition-colors">
                    <MoreVertical size={16} />
                  </button>
                </td>
              </tr>
            ))}
            {users.length === 0 && (
              <tr>
                <td colSpan={6} className="text-center py-8 text-secondary">No personnel records found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="card w-full max-w-md relative">
            <button 
              onClick={() => setShowCreateModal(false)}
              className="absolute top-4 right-4 text-secondary hover:text-white"
            >
              <X size={20} />
            </button>
            
            <div className="flex items-center gap-3 mb-6">
              <div className="w-10 h-10 rounded-lg bg-[rgba(0,212,170,0.1)] flex items-center justify-center border border-[rgba(0,212,170,0.3)]">
                <UserPlus size={20} className="text-accent" />
              </div>
              <div>
                <h2 className="text-lg font-bold">New Personnel</h2>
                <p className="text-sm text-secondary">Create a standard user account</p>
              </div>
            </div>

            {createError && (
              <div className="p-3 mb-4 text-sm bg-[rgba(255,77,106,0.1)] border border-[var(--danger-primary)] rounded text-danger">
                {createError}
              </div>
            )}

            <form onSubmit={handleCreateUser} className="space-y-4">
              <div className="form-group">
                <label className="form-label" htmlFor="name">Full Name</label>
                <input
                  id="name"
                  type="text"
                  className="input"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="Jane Doe"
                  required
                />
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="email">Email Address</label>
                <input
                  id="email"
                  type="email"
                  className="input mono"
                  value={newEmail}
                  onChange={(e) => setNewEmail(e.target.value)}
                  placeholder="jane.doe@tracelock.local"
                  required
                />
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="password">Initial Password</label>
                <input
                  id="password"
                  type="password"
                  className="input mono"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  placeholder="••••••••"
                  required
                  minLength={8}
                />
              </div>

              <div className="pt-4 flex justify-end gap-3">
                <button 
                  type="button" 
                  onClick={() => setShowCreateModal(false)}
                  className="btn bg-[rgba(255,255,255,0.05)] hover:bg-[rgba(255,255,255,0.1)] text-white"
                >
                  Cancel
                </button>
                <button 
                  type="submit" 
                  className="btn btn-primary"
                  disabled={createLoading}
                >
                  {createLoading ? 'Creating...' : 'Create Account'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
