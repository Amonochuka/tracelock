"use client";

import { useEffect, useState, useCallback, useRef } from 'react';
import { useAuth } from '@/context/AuthContext';
import { Shield, User, MoreVertical, Plus, UserPlus, X, Key, Map, Fingerprint, Trash2, PlusCircle } from 'lucide-react';
import { API_URL } from '@/lib/api';

interface UserObj {
  id: number;
  name: string;
  email: string;
  role: string;
  created_at: string;
}

interface ZoneObj {
  id: number;
  name: string;
  description: string;
  max_capacity: number;
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

  // Dropdown state
  const [dropdownOpenId, setDropdownOpenId] = useState<number | null>(null);

  // Access Modal State
  const [accessModalUser, setAccessModalUser] = useState<UserObj | null>(null);
  const [allZones, setAllZones] = useState<ZoneObj[]>([]);
  const [userAccess, setUserAccess] = useState<Set<number>>(new Set());
  const [accessLoading, setAccessLoading] = useState(false);
  const [accessError, setAccessError] = useState('');

  // Credential Modal State
  const [credentialModalUser, setCredentialModalUser] = useState<UserObj | null>(null);
  const [userCredentials, setUserCredentials] = useState<any[]>([]);
  const [credLoading, setCredLoading] = useState(false);
  const [credError, setCredError] = useState('');
  
  const [newCredMethod, setNewCredMethod] = useState('fingerprint');
  const [newCredHash, setNewCredHash] = useState('');
  const [enrollingCred, setEnrollingCred] = useState(false);

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

  const loadZones = async () => {
    if (!token) return;
    const res = await fetch(`${API_URL}/zones`, { headers: { 'Authorization': `Bearer ${token}` } });
    if (res.ok) {
      const data = await res.json();
      setAllZones(data || []);
    }
  };

  const handleManageAccess = async (user: UserObj) => {
    setDropdownOpenId(null);
    setAccessModalUser(user);
    setAccessLoading(true);
    setAccessError('');
    try {
      if (allZones.length === 0) await loadZones();
      const res = await fetch(`${API_URL}/users/${user.id}/access`, { headers: { 'Authorization': `Bearer ${token}` } });
      if (!res.ok) throw new Error('Failed to fetch user access');
      const data = await res.json();
      const accessSet = new Set<number>((data || []).map((z: ZoneObj) => z.id));
      setUserAccess(accessSet);
    } catch (e: unknown) {
      if (e instanceof Error) setAccessError(e.message);
      else setAccessError('Error loading access');
    } finally {
      setAccessLoading(false);
    }
  };

  const toggleZoneAccess = async (zoneId: number, currentStatus: boolean) => {
    if (!accessModalUser || !token) return;
    try {
      if (currentStatus) {
        const res = await fetch(`${API_URL}/admin/access`, {
          method: 'DELETE',
          headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
          body: JSON.stringify({ user_id: accessModalUser.id, zone_id: zoneId })
        });
        if (!res.ok) throw new Error('Failed to revoke access');
        setUserAccess(prev => {
          const next = new Set(prev);
          next.delete(zoneId);
          return next;
        });
      } else {
        const res = await fetch(`${API_URL}/admin/access`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
          body: JSON.stringify({ user_id: accessModalUser.id, zone_id: zoneId })
        });
        if (!res.ok) throw new Error('Failed to grant access');
        setUserAccess(prev => {
          const next = new Set(prev);
          next.add(zoneId);
          return next;
        });
      }
    } catch (e: unknown) {
      if (e instanceof Error) setAccessError(e.message);
      else setAccessError('Error updating access');
    }
  };

  const loadCredentials = async (user: UserObj) => {
    if (!token) return;
    setCredLoading(true);
    setCredError('');
    try {
      const res = await fetch(`${API_URL}/admin/users/${user.id}/credentials`, { headers: { 'Authorization': `Bearer ${token}` } });
      if (!res.ok) throw new Error('Failed to fetch credentials');
      const data = await res.json();
      setUserCredentials(data || []);
    } catch (e: unknown) {
      if (e instanceof Error) setCredError(e.message);
      else setCredError('Error loading credentials');
    } finally {
      setCredLoading(false);
    }
  };

  const handleManageCredentials = async (user: UserObj) => {
    setDropdownOpenId(null);
    setCredentialModalUser(user);
    await loadCredentials(user);
  };

  const handleEnrollCredential = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!credentialModalUser || !token) return;
    setEnrollingCred(true);
    setCredError('');
    try {
      const res = await fetch(`${API_URL}/admin/users/${credentialModalUser.id}/credentials`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
        body: JSON.stringify({ entry_method: newCredMethod, credential_hash: newCredHash })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to enroll credential');
      setNewCredHash('');
      await loadCredentials(credentialModalUser);
    } catch (e: unknown) {
      if (e instanceof Error) setCredError(e.message);
      else setCredError('Error enrolling credential');
    } finally {
      setEnrollingCred(false);
    }
  };

  const handleRevokeCredential = async (method: string) => {
    if (!credentialModalUser || !token) return;
    try {
      const res = await fetch(`${API_URL}/admin/users/${credentialModalUser.id}/credentials/${method}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (!res.ok) throw new Error('Failed to revoke credential');
      await loadCredentials(credentialModalUser);
    } catch (e: unknown) {
      if (e instanceof Error) setCredError(e.message);
      else setCredError('Error revoking credential');
    }
  };

  if (loading) return <div className="text-secondary">Loading personnel data...</div>;

  return (
    <div onClick={() => setDropdownOpenId(null)}>
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

      <div className="table-container pb-32">
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
                <td className="text-right relative">
                  <button 
                    onClick={(e) => {
                      e.stopPropagation();
                      setDropdownOpenId(dropdownOpenId === u.id ? null : u.id);
                    }}
                    className="p-2 text-secondary hover:text-white rounded hover:bg-[rgba(255,255,255,0.05)] transition-colors"
                  >
                    <MoreVertical size={16} />
                  </button>

                  {dropdownOpenId === u.id && (
                    <div className="absolute right-8 top-1/2 -translate-y-1/2 w-48 bg-[var(--bg-card)] border border-[var(--border-color)] rounded-lg shadow-xl z-10 overflow-hidden">
                      <button 
                        onClick={(e) => {
                          e.stopPropagation();
                          handleManageAccess(u);
                        }}
                        className="w-full text-left px-4 py-2.5 text-sm hover:bg-[rgba(255,255,255,0.05)] flex items-center gap-2"
                      >
                        <Key size={14} className="text-accent" />
                        Manage Access
                      </button>
                      <button 
                        onClick={(e) => {
                          e.stopPropagation();
                          handleManageCredentials(u);
                        }}
                        className="w-full text-left px-4 py-2.5 text-sm hover:bg-[rgba(255,255,255,0.05)] flex items-center gap-2 border-t border-[rgba(255,255,255,0.05)]"
                      >
                        <Fingerprint size={14} className="text-accent" />
                        Manage Credentials
                      </button>
                    </div>
                  )}
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

      {/* Access Modal */}
      {accessModalUser && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="card w-full max-w-lg relative max-h-[85vh] flex flex-col">
            <button 
              onClick={() => setAccessModalUser(null)}
              className="absolute top-4 right-4 text-secondary hover:text-white"
            >
              <X size={20} />
            </button>
            
            <div className="flex items-center gap-3 mb-6 flex-shrink-0">
              <div className="w-10 h-10 rounded-lg bg-[rgba(0,212,170,0.1)] flex items-center justify-center border border-[rgba(0,212,170,0.3)]">
                <Key size={20} className="text-accent" />
              </div>
              <div>
                <h2 className="text-lg font-bold">Manage Access</h2>
                <p className="text-sm text-secondary">Operator: <span className="text-white">{accessModalUser.name}</span></p>
              </div>
            </div>

            {accessError && (
              <div className="p-3 mb-4 text-sm bg-[rgba(255,77,106,0.1)] border border-[var(--danger-primary)] rounded text-danger flex-shrink-0">
                {accessError}
              </div>
            )}

            {accessLoading ? (
              <div className="py-10 text-center text-secondary">Loading zone policies...</div>
            ) : (
              <div className="flex-1 overflow-y-auto pr-2 space-y-3">
                {accessModalUser.role === 'admin' && (
                  <div className="p-4 bg-[rgba(0,212,170,0.05)] border border-[rgba(0,212,170,0.2)] rounded-lg mb-4">
                    <div className="flex items-start gap-2">
                      <Shield size={16} className="text-accent mt-0.5" />
                      <div>
                        <div className="font-semibold text-sm mb-1">Administrator Clearance</div>
                        <div className="text-xs text-secondary">This operator has global admin privileges. They inherently have access to all zones. Toggling access below sets explicit grants for if they are ever demoted to a regular user.</div>
                      </div>
                    </div>
                  </div>
                )}
                
                {allZones.length === 0 ? (
                  <div className="text-center text-secondary py-4 text-sm border border-dashed border-[var(--border-color)] rounded-lg">
                    No zones configured in the system.
                  </div>
                ) : (
                  allZones.map(zone => {
                    const hasAccess = userAccess.has(zone.id);
                    return (
                      <div key={zone.id} className="flex items-center justify-between p-3 rounded-lg border border-[var(--border-color)] bg-[rgba(255,255,255,0.02)]">
                        <div className="flex items-start gap-3">
                          <Map size={16} className="text-secondary mt-0.5" />
                          <div>
                            <div className="text-sm font-semibold">{zone.name}</div>
                            <div className="text-xs text-secondary mono">ZONE ID: {zone.id}</div>
                          </div>
                        </div>
                        <label className="flex items-center cursor-pointer">
                          <div className="relative">
                            <input 
                              type="checkbox" 
                              className="sr-only"
                              checked={hasAccess}
                              onChange={() => toggleZoneAccess(zone.id, hasAccess)}
                            />
                            <div className={`block w-10 h-6 rounded-full transition-colors ${hasAccess ? 'bg-accent' : 'bg-gray-700'}`}></div>
                            <div className={`dot absolute left-1 top-1 bg-white w-4 h-4 rounded-full transition-transform ${hasAccess ? 'transform translate-x-4' : ''}`}></div>
                          </div>
                        </label>
                      </div>
                    );
                  })
                )}
              </div>
            )}
            
            <div className="pt-6 flex justify-end flex-shrink-0">
              <button 
                type="button" 
                onClick={() => setAccessModalUser(null)}
                className="btn bg-[rgba(255,255,255,0.05)] hover:bg-[rgba(255,255,255,0.1)] text-white"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Credentials Modal */}
      {credentialModalUser && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="card w-full max-w-lg relative max-h-[85vh] flex flex-col">
            <button 
              onClick={() => setCredentialModalUser(null)}
              className="absolute top-4 right-4 text-secondary hover:text-white"
            >
              <X size={20} />
            </button>
            
            <div className="flex items-center gap-3 mb-6 flex-shrink-0">
              <div className="w-10 h-10 rounded-lg bg-[rgba(0,212,170,0.1)] flex items-center justify-center border border-[rgba(0,212,170,0.3)]">
                <Fingerprint size={20} className="text-accent" />
              </div>
              <div>
                <h2 className="text-lg font-bold">Manage Credentials</h2>
                <p className="text-sm text-secondary">Operator: <span className="text-white">{credentialModalUser.name}</span></p>
              </div>
            </div>

            {credError && (
              <div className="p-3 mb-4 text-sm bg-[rgba(255,77,106,0.1)] border border-[var(--danger-primary)] rounded text-danger flex-shrink-0">
                {credError}
              </div>
            )}

            {credLoading ? (
              <div className="py-10 text-center text-secondary">Loading credentials...</div>
            ) : (
              <div className="flex-1 overflow-y-auto pr-2 space-y-6">
                <div>
                  <h3 className="text-sm font-semibold mb-3">Enrolled Credentials</h3>
                  {userCredentials.length === 0 ? (
                    <div className="text-center text-secondary py-4 text-sm border border-dashed border-[var(--border-color)] rounded-lg">
                      No credentials enrolled for this user.
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {userCredentials.map((cred: {id: number; entry_method: string; credential_hash: string; revoked: boolean; enrolled_at: string}) => (
                        <div key={cred.id} className={`flex items-center justify-between p-3 rounded-lg border ${cred.revoked ? 'border-[rgba(255,77,106,0.2)] bg-[rgba(255,77,106,0.03)] opacity-60' : 'border-[var(--border-color)] bg-[rgba(255,255,255,0.02)]'}`}>
                          <div className="flex items-center gap-3">
                            <Fingerprint size={16} className={cred.revoked ? 'text-danger' : 'text-secondary'} />
                            <div>
                              <div className="flex items-center gap-2">
                                <span className="text-sm font-semibold capitalize">{cred.entry_method}</span>
                                {cred.revoked && (
                                  <span className="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded bg-[rgba(255,77,106,0.15)] text-danger border border-[rgba(255,77,106,0.3)]">REVOKED</span>
                                )}
                              </div>
                              <div className="text-xs text-secondary mono" title={cred.credential_hash}>
                                Hash: {cred.credential_hash.slice(0, 12)}…
                              </div>
                            </div>
                          </div>
                          {!cred.revoked && (
                            <button
                              onClick={() => handleRevokeCredential(cred.entry_method)}
                              className="p-1.5 text-danger hover:bg-[rgba(255,77,106,0.1)] rounded transition-colors"
                              title="Revoke Credential"
                            >
                              <Trash2 size={16} />
                            </button>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                <div className="border-t border-[var(--border-color)] pt-4">
                  <h3 className="text-sm font-semibold mb-3 flex items-center gap-2">
                    <PlusCircle size={16} className="text-accent" />
                    Enroll New Credential
                  </h3>
                  <form onSubmit={handleEnrollCredential} className="space-y-3">
                    <div className="grid grid-cols-2 gap-3">
                      <div className="form-group mb-0">
                        <label className="form-label text-xs">Method</label>
                        <select className="input h-9 text-sm" value={newCredMethod} onChange={e => setNewCredMethod(e.target.value)}>
                          <option value="fingerprint">Fingerprint</option>
                          <option value="face">Face Recognition</option>
                          <option value="iris">Iris Scanner</option>
                          <option value="card">Card Reader</option>
                          <option value="pin">PIN Pad</option>
                        </select>
                      </div>
                      <div className="form-group mb-0">
                        <label className="form-label text-xs">Raw Value / Hash</label>
                        <input className="input h-9 text-sm mono" value={newCredHash} onChange={e => setNewCredHash(e.target.value)} required placeholder="e.g. hash_123" />
                      </div>
                    </div>
                    <div className="flex justify-end pt-1">
                      <button type="submit" className="btn btn-primary text-sm px-4 h-9" disabled={enrollingCred}>
                        {enrollingCred ? 'Enrolling...' : 'Enroll Credential'}
                      </button>
                    </div>
                  </form>
                </div>
              </div>
            )}
            
            <div className="pt-6 flex justify-end flex-shrink-0">
              <button 
                type="button" 
                onClick={() => setCredentialModalUser(null)}
                className="btn bg-[rgba(255,255,255,0.05)] hover:bg-[rgba(255,255,255,0.1)] text-white"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
