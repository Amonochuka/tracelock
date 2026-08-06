"use client";

import { useEffect, useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { Shield, User, MoreVertical } from 'lucide-react';

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

  useEffect(() => {
    if (!token) return;

    const fetchUsers = async () => {
      try {
        const res = await fetch(`http://${window.location.hostname}:8080/admin/users`, {
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
    };

    fetchUsers();
  }, [token]);

  if (loading) return <div className="text-secondary">Loading personnel data...</div>;

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1>Personnel Access</h1>
          <p className="text-secondary mt-2">Manage user identities and roles</p>
        </div>
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
    </div>
  );
}
