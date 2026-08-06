"use client";

import { useEffect, useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { LogOut, Shield, Clock, CheckCircle, XCircle, MapPin } from 'lucide-react';

interface AccessEvent {
  id: number;
  zone_id: number;
  zone_name: string;
  action: string;
  status: string;
  entry_method: string;
  timestamp: string;
  reason: string;
}

interface ZoneAccess {
  zone_id: number;
  zone_name: string;
}

export default function UserDashboard() {
  const { user, token, logout } = useAuth();
  const [events, setEvents] = useState<AccessEvent[]>([]);
  const [access, setAccess] = useState<ZoneAccess[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;

    const base = `http://${window.location.hostname}:8080`;

    Promise.all([
      fetch(`${base}/me/events`, { headers: { Authorization: `Bearer ${token}` } }).then(r => r.ok ? r.json() : []),
      fetch(`${base}/me/access`, { headers: { Authorization: `Bearer ${token}` } }).then(r => r.ok ? r.json() : []),
    ]).then(([evts, acc]) => {
      setEvents(evts || []);
      setAccess(acc || []);
    }).finally(() => setLoading(false));
  }, [token]);

  if (!user) return null;

  return (
    <div className="flex flex-col min-h-screen">
      {/* Header */}
      <header className="nav-header">
        <div className="nav-container">
          <div className="flex items-center gap-2">
            <Shield className="text-accent" size={24} />
            <span className="font-bold text-lg">TraceLock</span>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex flex-col items-end">
              <span className="text-sm font-medium">{user.name}</span>
              <span className="text-xs text-secondary mono uppercase">{user.role}</span>
            </div>
            <button onClick={logout} className="btn btn-danger" style={{ padding: '0.5rem' }} title="Logout">
              <LogOut size={18} />
            </button>
          </div>
        </div>
      </header>

      <main className="container flex-1 w-full mt-8">
        <div className="mb-8">
          <h1>My Access Portal</h1>
          <p className="text-secondary mt-2">Your personal zone access history and permissions</p>
        </div>

        {loading ? (
          <div className="text-secondary">Loading your access data...</div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">

            {/* Authorized Zones Panel */}
            <div className="md:col-span-1">
              <h3 className="mb-4 flex items-center gap-2">
                <MapPin size={18} className="text-accent" />
                Authorized Zones
              </h3>
              <div className="flex flex-col gap-3">
                {access.length === 0 ? (
                  <div className="card text-secondary text-sm">No zone access granted yet.</div>
                ) : (
                  access.map(z => (
                    <div key={z.zone_id} className="card flex items-center gap-3">
                      <div className="status-dot status-active" />
                      <span className="font-medium">{z.zone_name}</span>
                      <span className="text-xs text-secondary mono ml-auto">ZONE {z.zone_id}</span>
                    </div>
                  ))
                )}
              </div>
            </div>

            {/* Event Log Panel */}
            <div className="md:col-span-2">
              <h3 className="mb-4 flex items-center gap-2">
                <Clock size={18} className="text-accent" />
                Recent Access Events
              </h3>
              <div className="table-container">
                <table>
                  <thead>
                    <tr>
                      <th>ZONE</th>
                      <th>ACTION</th>
                      <th>METHOD</th>
                      <th>STATUS</th>
                      <th>TIME</th>
                    </tr>
                  </thead>
                  <tbody>
                    {events.slice(0, 20).map(e => (
                      <tr key={e.id}>
                        <td className="font-medium">{e.zone_name || `Zone ${e.zone_id}`}</td>
                        <td className="mono text-xs uppercase tracking-wider">{e.action}</td>
                        <td className="text-secondary text-sm">{e.entry_method || '—'}</td>
                        <td>
                          <div className="flex items-center gap-1.5">
                            {e.status === 'allowed'
                              ? <CheckCircle size={14} className="text-accent" />
                              : <XCircle size={14} className="text-danger" />
                            }
                            <span className={`text-xs font-medium uppercase ${e.status === 'allowed' ? 'text-accent' : 'text-danger'}`}>
                              {e.status}
                            </span>
                          </div>
                        </td>
                        <td className="text-secondary text-sm mono">
                          {new Date(e.timestamp).toLocaleString()}
                        </td>
                      </tr>
                    ))}
                    {events.length === 0 && (
                      <tr>
                        <td colSpan={5} className="text-center py-8 text-secondary">No access events recorded yet.</td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
