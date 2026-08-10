"use client";

import { useEffect, useState, useCallback } from 'react';
import { useAuth } from '@/context/AuthContext';
import { useParams, useRouter } from 'next/navigation';
import {
  ArrowLeft, User, Activity, Clock, List, Settings, ShieldCheck, CheckCircle,
  XCircle, AlertTriangle, Trash2, Save, RefreshCw, ChevronLeft, ChevronRight
} from 'lucide-react';
import Link from 'next/link';
import { API_URL } from '@/lib/api';

interface UserObj {
  id: number;
  name: string;
  email: string;
  role: string;
}

interface ZoneObj {
  id: number;
  name: string;
  description: string;
  max_capacity: number;
  requires_exit_scan: boolean;
}

interface AccessEvent {
  id: number;
  user_id: number;
  zone_id: number;
  action: string;
  status: string;
  reason?: string;
  timestamp: string;
  entry_method: string;
  hash: string;
  previous_hash: string;
}

type Tab = 'overview' | 'events' | 'settings';

const PAGE_SIZE = 15;

export default function ZoneDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const { token } = useAuth();
  const router = useRouter();

  const [tab, setTab] = useState<Tab>('overview');
  const [zone, setZone] = useState<ZoneObj | null>(null);
  const [activeUsers, setActiveUsers] = useState<UserObj[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Event history
  const [events, setEvents] = useState<AccessEvent[]>([]);
  const [eventsTotal, setEventsTotal] = useState(0);
  const [eventsPage, setEventsPage] = useState(0);
  const [eventsLoading, setEventsLoading] = useState(false);

  // Chain verification
  const [chainResult, setChainResult] = useState<{ valid: boolean; events_checked: number; message: string } | null>(null);
  const [chainLoading, setChainLoading] = useState(false);

  // Zone editing
  const [editName, setEditName] = useState('');
  const [editDescription, setEditDescription] = useState('');
  const [editMaxCapacity, setEditMaxCapacity] = useState('');
  const [editLoading, setEditLoading] = useState(false);
  const [editFeedback, setEditFeedback] = useState({ type: '', msg: '' });

  // Deletion
  const [deleteConfirm, setDeleteConfirm] = useState('');
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [deleteError, setDeleteError] = useState('');

  const headers = { Authorization: `Bearer ${token}` };

  const fetchZone = useCallback(async () => {
    if (!token || !id) return;
    const zRes = await fetch(`${API_URL}/zones/${id}`, { headers: { Authorization: `Bearer ${token}` } });
    if (!zRes.ok) throw new Error('Failed to fetch zone details');
    const zData = await zRes.json();
    setZone(zData);
    setEditName(zData.name);
    setEditDescription(zData.description || '');
    setEditMaxCapacity(zData.max_capacity?.toString() || '0');
  }, [token, id]);

  const fetchActiveUsers = useCallback(async () => {
    if (!token || !id) return;
    const uRes = await fetch(`${API_URL}/admin/zones/${id}/active-users`, { headers: { Authorization: `Bearer ${token}` } });
    if (!uRes.ok) throw new Error('Failed to fetch active users');
    const uData = await uRes.json();
    setActiveUsers(uData || []);
  }, [token, id]);

  const fetchEvents = useCallback(async (page: number) => {
    if (!token || !id) return;
    setEventsLoading(true);
    const offset = page * PAGE_SIZE;
    const res = await fetch(`${API_URL}/zones/${id}/events?limit=${PAGE_SIZE}&offset=${offset}`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (res.ok) {
      const data = await res.json();
      setEvents(data.events || []);
      setEventsTotal(data.total || 0);
    }
    setEventsLoading(false);
  }, [token, id]);

  useEffect(() => {
    if (!token || !id) return;
    const init = async () => {
      try {
        await Promise.all([fetchZone(), fetchActiveUsers()]);
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setLoading(false);
      }
    };
    init();
  }, [token, id]);

  useEffect(() => {
    if (tab === 'events') fetchEvents(eventsPage);
  }, [tab, eventsPage]);

  const handleVerifyChain = async () => {
    setChainLoading(true);
    setChainResult(null);
    try {
      const res = await fetch(`${API_URL}/admin/zones/${id}/verify-chain`, { headers });
      const data = await res.json();
      setChainResult(data);
    } catch {
      setChainResult({ valid: false, events_checked: 0, message: 'Verification request failed.' });
    } finally {
      setChainLoading(false);
    }
  };

  const handleUpdateZone = async (e: React.FormEvent) => {
    e.preventDefault();
    setEditLoading(true);
    setEditFeedback({ type: '', msg: '' });
    try {
      const res = await fetch(`${API_URL}/admin/zones/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...headers },
        body: JSON.stringify({ name: editName, description: editDescription, max_capacity: parseInt(editMaxCapacity) || 0 })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to update zone');
      setZone(data);
      setEditFeedback({ type: 'success', msg: 'Zone settings saved successfully.' });
    } catch (err: unknown) {
      setEditFeedback({ type: 'error', msg: err instanceof Error ? err.message : 'Update failed' });
    } finally {
      setEditLoading(false);
    }
  };

  const handleDeleteZone = async () => {
    if (deleteConfirm !== zone?.name) return;
    setDeleteLoading(true);
    setDeleteError('');
    try {
      const res = await fetch(`${API_URL}/admin/zones/${id}`, { method: 'DELETE', headers });
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || 'Failed to delete zone');
      }
      router.push('/admin/zones');
    } catch (err: unknown) {
      setDeleteError(err instanceof Error ? err.message : 'Deletion failed');
      setDeleteLoading(false);
    }
  };

  const totalPages = Math.ceil(eventsTotal / PAGE_SIZE);

  if (loading) return <div className="text-secondary">Loading zone telemetry...</div>;
  if (!zone) return <div className="text-danger">Zone not found</div>;

  const tabClass = (t: Tab) =>
    `px-1 pb-3 text-sm font-medium border-b-2 transition-all ${
      tab === t
        ? 'border-[var(--accent-primary)] text-white'
        : 'border-transparent text-gray-500 hover:text-gray-300'
    }`;

  return (
    <div>
      {/* Header */}
      <div className="mb-6 flex items-center gap-4">
        <Link href="/admin/zones" className="p-2 rounded hover:bg-[rgba(255,255,255,0.05)] text-secondary transition-colors">
          <ArrowLeft size={20} />
        </Link>
        <div>
          <div className="flex items-center gap-3">
            <h1>{zone.name}</h1>
            <span className="text-xs font-medium px-2 py-1 rounded bg-[rgba(255,255,255,0.05)] text-secondary">
              ID: {zone.id}
            </span>
          </div>
          <p className="text-secondary mt-1">{zone.description || 'No description provided'}</p>
        </div>
      </div>

      {error && (
        <div className="p-4 mb-6 bg-red-500/10 border border-red-500 rounded-lg text-red-500">{error}</div>
      )}

      {/* Stat cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-1">
            <Activity size={16} />
            <span className="text-sm">Current Status</span>
          </div>
          <div className={`text-xl font-bold ${activeUsers.length > 0 ? 'text-warning' : 'text-accent'}`}>
            {activeUsers.length > 0 ? 'OCCUPIED' : 'SECURE'}
          </div>
        </div>

        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-1">
            <User size={16} />
            <span className="text-sm">Occupancy</span>
          </div>
          <div className="text-xl font-bold mono">
            {activeUsers.length}{' '}
            <span className="text-secondary text-sm font-normal">/ {zone.max_capacity > 0 ? zone.max_capacity : '∞'}</span>
          </div>
        </div>

        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-1">
            <Clock size={16} />
            <span className="text-sm">Protocol</span>
          </div>
          <div className="text-sm font-medium">
            {zone.requires_exit_scan ? 'Strict (Entry/Exit Scan)' : 'Standard (Auto-Exit)'}
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-8 mb-6 border-b border-[var(--border-color)]">
        <button type="button" onClick={() => setTab('overview')} className={tabClass('overview')}>Overview</button>
        <button type="button" onClick={() => setTab('events')} className={tabClass('events')}>Event History</button>
        <button type="button" onClick={() => setTab('settings')} className={tabClass('settings')}>Settings</button>
      </div>

      {/* ── Overview Tab ── */}
      {tab === 'overview' && (
        <div className="card">
          <h3 className="mb-4">Active Personnel</h3>
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>OPERATOR</th>
                  <th>CLEARANCE</th>
                  <th>STATUS</th>
                </tr>
              </thead>
              <tbody>
                {activeUsers.map(u => (
                  <tr key={u.id}>
                    <td className="mono text-secondary">#{u.id}</td>
                    <td><span className="font-medium">{u.name}</span></td>
                    <td><span className="text-xs uppercase tracking-wider text-secondary">{u.role}</span></td>
                    <td>
                      <div className="flex items-center gap-2">
                        <div className="status-dot status-active"></div>
                        <span className="text-xs font-medium text-accent">INSIDE</span>
                      </div>
                    </td>
                  </tr>
                ))}
                {activeUsers.length === 0 && (
                  <tr>
                    <td colSpan={4} className="text-center py-8 text-secondary">Zone is currently empty</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* ── Event History Tab ── */}
      {tab === 'events' && (
        <div className="space-y-4">
          {/* Chain Integrity */}
          <div className="card">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <ShieldCheck size={18} className="text-accent" />
                <h3 className="m-0">Hash Chain Integrity</h3>
              </div>
              <button
                onClick={handleVerifyChain}
                disabled={chainLoading}
                className="btn text-sm px-4 py-2"
              >
                <RefreshCw size={14} className={`mr-2 ${chainLoading ? 'animate-spin' : ''}`} />
                {chainLoading ? 'Verifying...' : 'Verify Chain'}
              </button>
            </div>
            {chainResult && (
              <div className={`flex items-start gap-3 p-4 rounded-lg border ${chainResult.valid
                ? 'bg-[rgba(0,212,170,0.07)] border-accent/30'
                : 'bg-[rgba(255,77,106,0.07)] border-danger/30'
              }`}>
                {chainResult.valid
                  ? <CheckCircle size={20} className="text-accent mt-0.5 flex-shrink-0" />
                  : <XCircle size={20} className="text-danger mt-0.5 flex-shrink-0" />
                }
                <div>
                  <div className={`font-semibold text-sm ${chainResult.valid ? 'text-accent' : 'text-danger'}`}>
                    {chainResult.valid ? 'Chain Intact' : 'Integrity Violation'}
                  </div>
                  <div className="text-secondary text-xs mt-0.5">
                    {chainResult.message} — {chainResult.events_checked} events checked.
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Event Log */}
          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <h3 className="m-0">Access Event Log</h3>
              <button onClick={() => fetchEvents(eventsPage)} className="btn text-xs px-3 py-1.5 h-auto">
                <RefreshCw size={13} className="mr-1" /> Refresh
              </button>
            </div>

            {eventsLoading ? (
              <div className="text-secondary text-sm py-8 text-center">Loading events...</div>
            ) : (
              <div className="table-container">
                <table>
                  <thead>
                    <tr>
                      <th>TIME</th>
                      <th>USER</th>
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
                        <td className="text-sm">#{ev.user_id}</td>
                        <td>
                          <span className={`text-xs uppercase font-semibold tracking-wider ${ev.action === 'enter' ? 'text-accent' : 'text-secondary'}`}>
                            {ev.action}
                          </span>
                        </td>
                        <td className="text-xs text-secondary">{ev.entry_method || '—'}</td>
                        <td>
                          <span className={`text-xs font-medium uppercase ${ev.status === 'allowed' ? 'text-accent' : 'text-danger'}`}>
                            {ev.status}
                          </span>
                        </td>
                        <td className="text-xs text-secondary">{ev.reason || '—'}</td>
                        <td className="mono text-xs text-secondary" title={ev.hash}>
                          {ev.hash?.slice(0, 10)}…
                        </td>
                      </tr>
                    ))}
                    {events.length === 0 && (
                      <tr>
                        <td colSpan={7} className="text-center py-8 text-secondary">No events recorded yet.</td>
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
      )}

      {/* ── Settings Tab ── */}
      {tab === 'settings' && (
        <div className="space-y-6">
          {/* Edit Form */}
          <div className="card">
            <div className="flex items-center gap-2 mb-4">
              <Save size={18} className="text-accent" />
              <h3 className="m-0">Zone Configuration</h3>
            </div>
            {editFeedback.msg && (
              <div className={`p-3 mb-4 text-sm rounded border ${editFeedback.type === 'error'
                ? 'bg-[rgba(255,77,106,0.1)] border-danger/40 text-danger'
                : 'bg-[rgba(0,212,170,0.1)] border-accent/40 text-accent'
              }`}>
                {editFeedback.msg}
              </div>
            )}
            <form onSubmit={handleUpdateZone} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="form-group">
                  <label className="form-label" htmlFor="edit-name">Zone Name</label>
                  <input
                    id="edit-name"
                    className="input"
                    value={editName}
                    onChange={e => setEditName(e.target.value)}
                    required
                  />
                </div>
                <div className="form-group">
                  <label className="form-label" htmlFor="edit-capacity">Max Capacity (0 = unlimited)</label>
                  <input
                    id="edit-capacity"
                    type="number"
                    min="0"
                    className="input mono"
                    value={editMaxCapacity}
                    onChange={e => setEditMaxCapacity(e.target.value)}
                  />
                </div>
              </div>
              <div className="form-group">
                <label className="form-label" htmlFor="edit-description">Description</label>
                <input
                  id="edit-description"
                  className="input"
                  value={editDescription}
                  onChange={e => setEditDescription(e.target.value)}
                />
              </div>
              <div className="pt-2">
                <button type="submit" className="btn btn-primary" disabled={editLoading}>
                  <Save size={15} className="mr-2" />
                  {editLoading ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </form>
          </div>

          {/* Danger Zone */}
          <div className="card border border-danger/30 bg-[rgba(255,77,106,0.03)]">
            <div className="flex items-start gap-3 mb-4">
              <AlertTriangle size={20} className="text-danger mt-0.5 flex-shrink-0" />
              <div>
                <h3 className="m-0 text-danger">Danger Zone</h3>
                <p className="text-secondary text-sm mt-1">
                  Deleting a zone is permanent. Zones with active sessions cannot be deleted.
                  All historical access events will remain in the audit log.
                </p>
              </div>
            </div>
            {deleteError && (
              <div className="p-3 mb-4 text-sm bg-[rgba(255,77,106,0.1)] border border-danger/40 rounded text-danger">
                {deleteError}
              </div>
            )}
            <div className="form-group mb-3">
              <label className="form-label" htmlFor="delete-confirm">
                Type <span className="text-danger font-semibold mono">"{zone.name}"</span> to confirm deletion
              </label>
              <input
                id="delete-confirm"
                className="input border-danger/30 focus:border-danger"
                placeholder={zone.name}
                value={deleteConfirm}
                onChange={e => setDeleteConfirm(e.target.value)}
              />
            </div>
            <button
              onClick={handleDeleteZone}
              disabled={deleteConfirm !== zone.name || deleteLoading}
              className="btn border border-danger/40 text-danger hover:bg-danger/10 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              <Trash2 size={15} className="mr-2" />
              {deleteLoading ? 'Deleting...' : 'Delete Zone Permanently'}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
