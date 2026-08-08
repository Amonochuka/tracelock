"use client";

import { useEffect, useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { Users, AlertTriangle, Search } from 'lucide-react';
import { API_URL, WS_URL } from '@/lib/api';

interface Zone {
  id: number;
  name: string;
  max_capacity: number;
}

interface ZoneOccupancy {
  zone: Zone;
  active_count: number;
}

export default function AdminDashboard() {
  const { token } = useAuth();
  const [occupancy, setOccupancy] = useState<Record<number, ZoneOccupancy>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [wsConnected, setWsConnected] = useState(false);
  const [search, setSearch] = useState('');

  useEffect(() => {
    if (!token) return;
    const fetchOccupancy = async () => {
      try {
        const res = await fetch(`${API_URL}/zones/occupancy`, {
          headers: { Authorization: `Bearer ${token}` }
        });
        if (!res.ok) throw new Error('Failed to fetch occupancy');
        const data = await res.json();
        const occMap: Record<number, ZoneOccupancy> = {};
        if (data && data.length > 0) {
          data.forEach((z: { id: number; name: string; max_capacity: number; active_count: number }) => {
            occMap[z.id] = { zone: { id: z.id, name: z.name, max_capacity: z.max_capacity }, active_count: z.active_count || 0 };
          });
        }
        setOccupancy(occMap);
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : 'An unknown error occurred');
      } finally {
        setLoading(false);
      }
    };
    fetchOccupancy();
  }, [token]);

  useEffect(() => {
    if (!token) return;
    const ws = new WebSocket(`${WS_URL}/ws/zones?token=${token}`);
    ws.onopen = () => setWsConnected(true);
    ws.onmessage = (event) => {
      try {
        const update: ZoneOccupancy = JSON.parse(event.data);
        if (update?.zone?.id) {
          setOccupancy(prev => ({ ...prev, [update.zone.id]: update }));
        }
      } catch (e) { console.error('Failed to parse WS message', e); }
    };
    ws.onclose = () => setWsConnected(false);
    return () => ws.close();
  }, [token]);

  const filtered = Object.values(occupancy).filter(occ =>
    occ.zone.name.toLowerCase().includes(search.toLowerCase())
  );

  if (loading) return <div className="text-secondary">Loading dashboard...</div>;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1>Global Zone Occupancy</h1>
          <p className="text-secondary mt-1">Real-time facility monitoring</p>
        </div>
        <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-[var(--bg-card)] border border-[var(--border-color)]">
          <div className={`status-dot ${wsConnected ? 'status-active' : 'status-danger'}`} />
          <span className="text-xs font-medium text-secondary">
            {wsConnected ? 'LIVE FEED CONNECTED' : 'FEED DISCONNECTED'}
          </span>
        </div>
      </div>

      {error && (
        <div className="p-4 mb-6 bg-red-500/10 border border-red-500 rounded-lg text-red-500">{error}</div>
      )}

      {/* Search */}
      <div className="relative mb-4">
        <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-secondary pointer-events-none" />
        <input
          className="input pl-9 h-9 text-sm w-full max-w-xs"
          placeholder="Filter zones..."
          value={search}
          onChange={e => setSearch(e.target.value)}
        />
        {Object.keys(occupancy).length > 0 && (
          <span className="ml-3 text-xs text-secondary">
            {filtered.length} / {Object.keys(occupancy).length} zones
          </span>
        )}
      </div>

      {/* Scrollable zone grid */}
      <div
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 overflow-y-auto pr-1"
        style={{ maxHeight: 'calc(100vh - 260px)' }}
      >
        {filtered.map((occ) => {
          const isAtCapacity = occ.zone.max_capacity > 0 && occ.active_count >= occ.zone.max_capacity;
          const isNearCapacity = occ.zone.max_capacity > 0 && (occ.active_count / occ.zone.max_capacity) > 0.8;
          const statusColor = isAtCapacity ? 'var(--danger-primary)' : isNearCapacity ? 'var(--warning-primary)' : 'var(--accent-primary)';
          const percent = occ.zone.max_capacity > 0 ? Math.min(100, Math.round((occ.active_count / occ.zone.max_capacity) * 100)) : 0;

          return (
            <div key={occ.zone.id} className="card flex flex-col gap-3 relative overflow-hidden py-4">
              {isAtCapacity && (
                <div className="absolute top-3 right-3 animate-pulse">
                  <AlertTriangle size={16} className="text-danger" />
                </div>
              )}
              <div className="flex justify-between items-start">
                <div>
                  <h3 className="text-sm font-semibold mb-0.5">{occ.zone.name}</h3>
                  <span className="text-[10px] text-secondary font-medium px-1.5 py-0.5 rounded" style={{ background: 'rgba(255,255,255,0.05)' }}>
                    ZONE {occ.zone.id}
                  </span>
                </div>
              </div>
              <div className="flex items-end gap-2">
                <Users size={13} className="text-secondary mb-0.5" />
                <span className="text-2xl font-bold mono" style={{ color: statusColor }}>{occ.active_count}</span>
                {occ.zone.max_capacity > 0 && (
                  <span className="text-secondary mono text-xs mb-0.5">/ {occ.zone.max_capacity}</span>
                )}
              </div>
              {occ.zone.max_capacity > 0 && (
                <div className="progress-bg">
                  <div className="progress-fill" style={{ width: `${percent}%`, backgroundColor: statusColor }} />
                </div>
              )}
            </div>
          );
        })}

        {filtered.length === 0 && !loading && (
          <div className="col-span-full p-8 text-center text-secondary border border-dashed border-[var(--border-color)] rounded-xl">
            {search ? `No zones matching "${search}"` : 'No zones configured in the system.'}
          </div>
        )}
      </div>
    </div>
  );
}
