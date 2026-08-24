"use client";

import { useEffect, useState, useCallback } from 'react';
import { useAuth } from '@/context/AuthContext';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import {
  ArrowLeft, User, Shield, Activity, Clock, LogIn, LogOut, Ban, MapPin,
  CheckCircle, XCircle, ChevronLeft, ChevronRight, RefreshCw
} from 'lucide-react';
import { API_URL } from '@/lib/api';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';

interface UserObj {
  id: number;
  name: string;
  email: string;
  role: string;
  created_at: string;
}

interface ZoneBreakdown {
  zone_id: number;
  zone_name: string;
  entries: number;
  denied: number;
  last_seen: string;
}

interface UserAnalytics {
  total_events: number;
  entries: number;
  exits: number;
  denied: number;
  zones_visited: number;
  last_event_at?: string | null;
  zones: ZoneBreakdown[];
}

interface AccessEvent {
  id: number;
  zone_id: number;
  zone_name?: string;
  action: string;
  status: string;
  reason?: string;
  timestamp: string;
  entry_method?: string;
  hash: string;
}

const PAGE_SIZE = 15;

export default function UserDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const { token } = useAuth();

  const [user, setUser] = useState<UserObj | null>(null);
  const [analytics, setAnalytics] = useState<UserAnalytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Event history pagination
  const [events, setEvents] = useState<AccessEvent[]>([]);
  const [eventsTotal, setEventsTotal] = useState(0);
  const [eventsPage, setEventsPage] = useState(0);
  const [eventsLoading, setEventsLoading] = useState(false);

  useEffect(() => {
    if (!token || !id) return;

    const init = async () => {
      setLoading(true);
      setError('');
      try {
        const headers = { Authorization: `Bearer ${token}` };
        const [usersRes, analyticsRes] = await Promise.all([
          fetch(`${API_URL}/admin/users`, { headers }),
          fetch(`${API_URL}/admin/users/${id}/analytics`, { headers }),
        ]);

        if (!analyticsRes.ok) throw new Error('Failed to fetch user activity');
        setAnalytics(await analyticsRes.json());

        if (usersRes.ok) {
          const users: UserObj[] = await usersRes.json();
          setUser(users.find(u => u.id === Number(id)) || null);
        }
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setLoading(false);
      }
    };

    init();
  }, [token, id]);

  const fetchEvents = useCallback(async (page: number) => {
    if (!token || !id) return;
    setEventsLoading(true);
    try {
      const offset = page * PAGE_SIZE;
      const res = await fetch(`${API_URL}/admin/users/${id}/events?limit=${PAGE_SIZE}&offset=${offset}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (!res.ok) throw new Error('Failed to fetch events');
      const data = await res.json();
      setEvents(data.events || []);
      setEventsTotal(data.total || 0);
    } catch {
      setEvents([]);
    } finally {
      setEventsLoading(false);
    }
  }, [token, id]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchEvents(eventsPage);
  }, [fetchEvents, eventsPage]);

  const totalPages = Math.ceil(eventsTotal / PAGE_SIZE);

  // Chart data: one bar per zone, sorted by entries (already server-sorted).
  const chartData = (analytics?.zones || []).map(z => ({
    name: z.zone_name.length > 18 ? `${z.zone_name.slice(0, 17)}…` : z.zone_name,
    fullName: z.zone_name,
    entries: z.entries,
    denied: z.denied,
  }));
  const hasActivity = !!analytics && analytics.total_events > 0;

  if (loading) return <div className="text-secondary">Loading operator telemetry...</div>;

  return (
    <div>
      {/* Header */}
      <div className="mb-6 flex items-center gap-4">
        <Link href="/admin/users" className="p-2 rounded hover:bg-[rgba(255,255,255,0.05)] text-secondary transition-colors">
          <ArrowLeft size={20} />
        </Link>
        <div className="w-12 h-12 rounded-full bg-[var(--bg-main)] border border-[var(--border-color)] flex items-center justify-center flex-shrink-0">
          <User size={22} className="text-secondary" />
        </div>
        <div>
          <div className="flex items-center gap-3 flex-wrap">
            <h1 className="m-0">{user ? user.name : `Operator #${id}`}</h1>
            {user && (
              <>
                <span className={`px-2 py-1 rounded text-xs font-medium uppercase tracking-wider ${
                  user.role === 'admin'
                    ? 'bg-[rgba(0,212,170,0.1)] text-accent border border-[rgba(0,212,170,0.2)]'
                    : 'bg-[rgba(255,255,255,0.05)] text-secondary'
                }`}>
                  {user.role === 'admin' && <Shield size={10} className="inline mr-1" />}
                  {user.role}
                </span>
                <span className="text-xs font-medium px-2 py-1 rounded bg-[rgba(255,255,255,0.05)] text-secondary mono">
                  ID: {user.id}
                </span>
              </>
            )}
          </div>
          <p className="text-secondary mt-1">
            {user ? (
              <>Operator activity record · enrolled {new Date(user.created_at).toLocaleDateString()}</>
            ) : (
              'This operator no longer exists in the personnel list.'
            )}
          </p>
        </div>
      </div>

      {error && (
        <div className="p-4 mb-6 bg-red-500/10 border border-red-500 rounded-lg text-red-500">{error}</div>
      )}

      {/* Stat cards */}
      <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-4 mb-6">
        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-1">
            <Activity size={16} />
            <span className="text-sm">Total Events</span>
          </div>
          <div className="text-xl font-bold mono">{analytics?.total_events ?? 0}</div>
        </div>

        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-1">
            <LogIn size={16} />
            <span className="text-sm">Entries</span>
          </div>
          <div className="text-xl font-bold mono text-accent">{analytics?.entries ?? 0}</div>
        </div>

        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-1">
            <LogOut size={16} />
            <span className="text-sm">Exits</span>
          </div>
          <div className="text-xl font-bold mono">{analytics?.exits ?? 0}</div>
        </div>

        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-1">
            <Ban size={16} />
            <span className="text-sm">Denied</span>
          </div>
          <div className={`text-xl font-bold mono ${(analytics?.denied ?? 0) > 0 ? 'text-danger' : ''}`}>
            {analytics?.denied ?? 0}
          </div>
        </div>

        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-1">
            <MapPin size={16} />
            <span className="text-sm">Zones Visited</span>
          </div>
          <div className="text-xl font-bold mono">{analytics?.zones_visited ?? 0}</div>
        </div>

        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-1">
            <Clock size={16} />
            <span className="text-sm">Last Seen</span>
          </div>
          <div className="text-sm font-medium mono">
            {analytics?.last_event_at ? new Date(analytics.last_event_at).toLocaleString() : '—'}
          </div>
        </div>
      </div>

      {/* Zone breakdown chart */}
      <div className="card w-full mb-6" style={{ minHeight: hasActivity ? '300px' : 'auto' }}>
        <div className="flex items-center gap-2 mb-6">
          <MapPin size={18} className="text-accent" />
          <h3 className="m-0 text-sm font-semibold">Zone Activity Breakdown</h3>
        </div>
        {!hasActivity ? (
          <div className="py-8 text-center text-secondary text-sm border border-dashed border-[var(--border-color)] rounded-xl">
            No access events recorded for this operator yet.
          </div>
        ) : (
          <div style={{ width: '100%', height: Math.max(180, chartData.length * 48) }}>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData} layout="vertical" margin={{ top: 5, right: 24, left: 8, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" horizontal={false} />
                <XAxis
                  type="number"
                  allowDecimals={false}
                  tick={{ fill: '#8b949e', fontSize: 10 }}
                  axisLine={{ stroke: 'rgba(255,255,255,0.1)' }}
                  tickLine={false}
                />
                <YAxis
                  type="category"
                  dataKey="name"
                  width={130}
                  tick={{ fill: '#8b949e', fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                />
                <Tooltip
                  cursor={{ fill: 'rgba(255,255,255,0.05)' }}
                  contentStyle={{
                    backgroundColor: 'var(--bg-main)',
                    border: '1px solid var(--border-color)',
                    borderRadius: '8px',
                    fontSize: '12px'
                  }}
                  labelFormatter={(_, payload) => payload?.[0]?.payload?.fullName ?? ''}
                  itemStyle={{ color: 'var(--accent-primary)' }}
                />
                <Bar dataKey="entries" fill="var(--accent-primary)" radius={[0, 4, 4, 0]} name="Entries" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>

      {/* Event history */}
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <h3 className="m-0">Access History</h3>
          <button onClick={() => fetchEvents(eventsPage)} disabled={eventsLoading} className="btn text-xs px-3 py-1.5 h-auto">
            <RefreshCw size={13} className={`mr-1 ${eventsLoading ? 'animate-spin' : ''}`} /> Refresh
          </button>
        </div>

        {eventsLoading ? (
          <div className="text-secondary text-sm py-8 text-center">Loading history...</div>
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>TIME</th>
                  <th>ZONE</th>
                  <th>ACTION</th>
                  <th>METHOD</th>
                  <th>STATUS</th>
                  <th>REASON</th>
                  <th>HASH</th>
                </tr>
              </thead>
              <tbody>
                {events.map(ev => (
                  <tr key={ev.id}>
                    <td className="text-xs text-secondary whitespace-nowrap mono">
                      {new Date(ev.timestamp).toLocaleString()}
                    </td>
                    <td className="text-sm">{ev.zone_name || `Zone ${ev.zone_id}`}</td>
                    <td>
                      <span className={`text-xs uppercase font-semibold tracking-wider ${ev.action === 'enter' ? 'text-accent' : 'text-secondary'}`}>
                        {ev.action}
                      </span>
                    </td>
                    <td className="text-xs text-secondary">{ev.entry_method || '—'}</td>
                    <td>
                      <div className="flex items-center gap-1.5">
                        {ev.status === 'allowed'
                          ? <CheckCircle size={14} className="text-accent" />
                          : <XCircle size={14} className="text-danger" />
                        }
                        <span className={`text-xs font-medium uppercase ${ev.status === 'allowed' ? 'text-accent' : 'text-danger'}`}>
                          {ev.status}
                        </span>
                      </div>
                    </td>
                    <td className="text-xs text-secondary">{ev.reason || '—'}</td>
                    <td className="mono text-xs text-secondary" title={ev.hash}>
                      {ev.hash?.slice(0, 10)}…
                    </td>
                  </tr>
                ))}
                {events.length === 0 && (
                  <tr>
                    <td colSpan={7} className="text-center py-8 text-secondary">No access events recorded yet.</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}

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
  );
}
