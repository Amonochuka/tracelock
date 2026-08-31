"use client";

import { useCallback, useEffect, useMemo, useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { LogOut, Shield, Clock, CheckCircle, XCircle, MapPin, LogIn, LogOut as Exit, ChevronLeft, ChevronRight, Fingerprint } from 'lucide-react';
import { API_URL } from '@/lib/api';

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

interface Credential {
  id: number;
  entry_method: string;
  credential_hash: string;
  enrolled_at: string;
  revoked: boolean;
}

interface EventResponse {
  events: AccessEvent[];
  total?: number;
}

const PAGE_SIZE = 15;

export default function UserDashboard() {
  const { user, token, logout } = useAuth();
  const [events, setEvents] = useState<AccessEvent[]>([]);
  const [access, setAccess] = useState<ZoneAccess[]>([]);
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [loading, setLoading] = useState(true);
  const [simulationMessage, setSimulationMessage] = useState('');
  const [simulatingZoneId, setSimulatingZoneId] = useState<number | null>(null);
  const [eventsPage, setEventsPage] = useState(0);
  const [eventsTotal, setEventsTotal] = useState(0);

  const loadAccessData = useCallback(async () => {
    if (!token) return;

    const headers = { Authorization: `Bearer ${token}` };
    const [eventsResponse, accessResponse, credResponse] = await Promise.all([
      fetch(`${API_URL}/me/events?limit=${PAGE_SIZE}&offset=${eventsPage * PAGE_SIZE}`, { headers }),
      fetch(`${API_URL}/me/access`, { headers }),
      fetch(`${API_URL}/me/credentials`, { headers }),
    ]);

    const eventsData: EventResponse = eventsResponse.ok ? await eventsResponse.json() : { events: [], total: 0 };
    const accessData: ZoneAccess[] = accessResponse.ok ? await accessResponse.json() : [];
    const credData: Credential[] = credResponse.ok ? await credResponse.json() : [];
    setEvents(eventsData.events || []);
    setEventsTotal(eventsData.total || 0);
    setAccess(accessData || []);
    setCredentials(credData || []);
  }, [token, eventsPage]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadAccessData().finally(() => setLoading(false));
  }, [loadAccessData]);

  const totalPages = Math.ceil(eventsTotal / PAGE_SIZE);

  const activeZoneIds = useMemo(() => {
    const latestEvents = new Map<number, AccessEvent>();
    events.filter((event) => event.status === 'allowed').forEach((event) => {
      const current = latestEvents.get(event.zone_id);
      if (!current || new Date(event.timestamp) > new Date(current.timestamp)) {
        latestEvents.set(event.zone_id, event);
      }
    });
    return new Set([...latestEvents.values()].filter((event) => event.action === 'enter').map((event) => event.zone_id));
  }, [events]);

  const simulateZoneEvent = async (zone: ZoneAccess, action: 'enter' | 'exit') => {
    if (!token) return;

    setSimulationMessage('');
    setSimulatingZoneId(zone.zone_id);
    try {
      const response = await fetch(`${API_URL}/zones/${action}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ zone_id: zone.zone_id, entry_method: 'web_simulator' }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || `Could not ${action} zone`);
      setSimulationMessage(`${action === 'enter' ? 'Entry' : 'Exit'} recorded for ${zone.zone_name}. The live admin feed has been updated.`);
      await loadAccessData();
    } catch (caughtError: unknown) {
      setSimulationMessage(caughtError instanceof Error ? caughtError.message : 'The simulation could not be completed');
    } finally {
      setSimulatingZoneId(null);
    }
  };

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

            <div className="md:col-span-1">
              <h3 className="mb-4 flex items-center gap-2">
                <MapPin size={18} className="text-accent" />
                Authorized Zones
              </h3>
              <p className="text-secondary text-sm" style={{ marginBottom: '1rem' }}>Use the browser simulator while hardware is unavailable. Each action follows the normal access and audit path.</p>
              {simulationMessage && <div className="card text-sm" style={{ marginBottom: '1rem' }}>{simulationMessage}</div>}
              <div className="flex flex-col gap-3">
                {access.length === 0 ? (
                  <div className="card text-secondary text-sm">No zone access granted yet.</div>
                ) : (
                  access.map(z => (
                    <div key={z.zone_id} className="card">
                      <div className="flex items-center gap-3">
                        <div className={`status-dot ${activeZoneIds.has(z.zone_id) ? 'status-active' : 'status-inactive'}`} />
                        <span className="font-medium">{z.zone_name}</span>
                        <span className="text-xs text-secondary mono ml-auto">ZONE {z.zone_id}</span>
                      </div>
                      <button className="btn btn-primary text-sm" style={{ width: '100%', marginTop: '1rem', padding: '0.6rem' }} disabled={simulatingZoneId === z.zone_id} onClick={() => simulateZoneEvent(z, activeZoneIds.has(z.zone_id) ? 'exit' : 'enter')}>
                        {activeZoneIds.has(z.zone_id) ? <Exit size={16} style={{ marginRight: '8px' }} /> : <LogIn size={16} style={{ marginRight: '8px' }} />}
                        {simulatingZoneId === z.zone_id ? 'Recording...' : activeZoneIds.has(z.zone_id) ? 'Simulate exit' : 'Simulate entry'}
                      </button>
                    </div>
                  ))
                )}
              </div>

              <h3 className="mb-4 mt-10 flex items-center gap-2">
                <Fingerprint size={18} className="text-accent" />
                My Credentials
              </h3>
              <p className="text-secondary text-sm" style={{ marginBottom: '1rem' }}>Biometric and card credentials currently enrolled against your account.</p>
              {credentials.length === 0 ? (
                <div className="card text-secondary text-sm">No credentials enrolled yet. Ask an administrator to add one.</div>
              ) : (
                <div className="flex flex-col gap-2">
                  {credentials.map(c => (
                    <div key={c.id} className={`card !p-3 ${c.revoked ? 'opacity-60' : ''}`}>
                      <div className="flex items-center gap-3">
                        <Fingerprint size={16} className={c.revoked ? 'text-danger' : 'text-accent'} />
                        <span className="text-sm font-semibold capitalize">{c.entry_method}</span>
                        {c.revoked && (
                          <span className="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded bg-[rgba(255,77,106,0.15)] text-danger border border-[rgba(255,77,106,0.3)] ml-auto">REVOKED</span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

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
                    {events.map(e => (
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

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-between mt-4 pt-4 border-t border-[var(--border-color)]">
                  <span className="text-xs text-secondary">
                    Page {eventsPage + 1} of {totalPages} · {eventsTotal} events
                  </span>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setEventsPage(p => Math.max(0, p - 1))}
                      disabled={eventsPage === 0}
                      className="btn text-xs px-3 py-1.5 h-auto disabled:opacity-40"
                    >
                      <ChevronLeft size={14} />
                    </button>
                    <button
                      onClick={() => setEventsPage(p => Math.min(totalPages - 1, p + 1))}
                      disabled={eventsPage >= totalPages - 1}
                      className="btn text-xs px-3 py-1.5 h-auto disabled:opacity-40"
                    >
                      <ChevronRight size={14} />
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
