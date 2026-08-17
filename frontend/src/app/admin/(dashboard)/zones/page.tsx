"use client";

import { FormEvent, useCallback, useEffect, useState } from 'react';
import Link from 'next/link';
import { Map, Plus, Users, Search, Lock, Unlock } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { API_URL } from '@/lib/api';

interface Zone {
  id: number;
  name: string;
  description: string;
  max_capacity: number;
  requires_exit_scan: boolean;
}

export default function ZonesPage() {
  const { token } = useAuth();
  const [zones, setZones] = useState<Zone[]>([]);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [maxCapacity, setMaxCapacity] = useState('');
  const [requiresExitScan, setRequiresExitScan] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [showForm, setShowForm] = useState(false);

  const loadZones = useCallback(async () => {
    if (!token) return;
    const response = await fetch(`${API_URL}/zones`, { headers: { Authorization: `Bearer ${token}` } });
    if (!response.ok) throw new Error('Unable to load zones');
    setZones(await response.json());
  }, [token]);

  useEffect(() => {
    loadZones()
      .catch((e: unknown) => setError(e instanceof Error ? e.message : 'Unable to load zones'))
      .finally(() => setLoading(false));
  }, [loadZones]);

  const createZone = async (event: FormEvent) => {
    event.preventDefault();
    if (!token) return;
    setError('');
    setSaving(true);
    try {
      const response = await fetch(`${API_URL}/admin/zones`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ name, description, max_capacity: Number(maxCapacity) || 0, requires_exit_scan: requiresExitScan }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || 'Unable to create zone');
      setZones(z => [...z, data]);
      setName(''); setDescription(''); setMaxCapacity(''); setRequiresExitScan(false);
      setShowForm(false);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Unable to create zone');
    } finally {
      setSaving(false);
    }
  };

  const filtered = zones.filter(z =>
    z.name.toLowerCase().includes(search.toLowerCase()) ||
    z.description?.toLowerCase().includes(search.toLowerCase())
  );

  if (loading) return <div className="text-secondary">Loading zones...</div>;

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1>Zones</h1>
          <p className="text-secondary mt-1">Configure monitored areas and their capacity rules.</p>
        </div>
        <button
          className="btn btn-primary"
          onClick={() => setShowForm(v => !v)}
        >
          <Plus size={16} className="mr-2" />
          {showForm ? 'Cancel' : 'New Zone'}
        </button>
      </div>

      {error && <div className="p-4 mb-6 bg-red-500/10 border border-red-500 rounded-lg text-red-500">{error}</div>}

      {/* Inline create form */}
      {showForm && (
        <div className="card mb-6">
          <h3 className="flex items-center gap-2 mb-4">
            <Plus size={16} className="text-accent" /> Create Zone
          </h3>
          <form onSubmit={createZone}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div className="form-group">
                <label className="form-label" htmlFor="zone-name">Name</label>
                <input id="zone-name" className="input" value={name} onChange={e => setName(e.target.value)} required disabled={saving} pattern=".*\S+.*" title="Zone name cannot be empty" />
              </div>
              <div className="form-group">
                <label className="form-label" htmlFor="zone-capacity">Max Capacity (0 = unlimited)</label>
                <input id="zone-capacity" className="input" type="number" min="0" value={maxCapacity} onChange={e => setMaxCapacity(e.target.value)} disabled={saving} />
              </div>
              <div className="form-group">
                <label className="form-label" htmlFor="zone-description">Description</label>
                <input id="zone-description" className="input" value={description} onChange={e => setDescription(e.target.value)} disabled={saving} />
              </div>
              <label className="flex items-center gap-2 text-sm text-secondary self-center" style={saving ? { opacity: 0.6, pointerEvents: 'none' } : {}}>
                <input type="checkbox" checked={requiresExitScan} onChange={e => setRequiresExitScan(e.target.checked)} disabled={saving} />
                Require explicit exit scan
              </label>
            </div>
            <button className="btn btn-primary" type="submit" disabled={saving}>
              {saving ? <><span className="spinner"></span> Creating...</> : 'Create Zone'}
            </button>
          </form>
        </div>
      )}

      {/* Search bar */}
      <div className="relative mb-4 flex items-center gap-3">
        <div className="relative">
          <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-secondary pointer-events-none" />
          <input
            className="input pl-9 h-9 text-sm w-64"
            placeholder="Search zones..."
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </div>
        <span className="text-xs text-secondary">{filtered.length} of {zones.length} zones</span>
      </div>

      {/* Zone cards — scrollable */}
      {zones.length === 0 ? (
        <div className="card text-center py-10">
          <Map size={28} className="text-accent mx-auto mb-3" />
          <h3>No zones yet</h3>
          <p className="text-secondary text-sm mt-2">Create the first zone above, then grant user access before running entry simulations.</p>
        </div>
      ) : (
        <div
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 overflow-y-auto pr-1"
          style={{ maxHeight: 'calc(100vh - 300px)' }}
        >
          {filtered.map((zone) => (
            <Link
              href={`/admin/zones/${zone.id}`}
              key={zone.id}
              className="card group hover:border-accent/40 transition-colors"
              style={{ textDecoration: 'none', color: 'inherit' }}
            >
              <div className="flex items-start justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Map size={16} className="text-accent flex-shrink-0" />
                  <span className="font-semibold text-sm">{zone.name}</span>
                </div>
                <span className="text-[10px] text-secondary mono px-1.5 py-0.5 rounded" style={{ background: 'rgba(255,255,255,0.05)' }}>
                  ZONE {zone.id}
                </span>
              </div>

              {zone.description && (
                <p className="text-secondary text-xs mb-2 line-clamp-1">{zone.description}</p>
              )}

              <div className="flex items-center gap-3 mt-1 text-xs text-secondary">
                <span className="flex items-center gap-1">
                  <Users size={11} />
                  {zone.max_capacity > 0 ? `Cap: ${zone.max_capacity}` : 'Unlimited'}
                </span>
                <span className="flex items-center gap-1">
                  {zone.requires_exit_scan
                    ? <><Lock size={11} /> Exit required</>
                    : <><Unlock size={11} /> Auto-exit</>
                  }
                </span>
              </div>
            </Link>
          ))}

          {filtered.length === 0 && search && (
            <div className="col-span-full p-8 text-center text-secondary border border-dashed border-[var(--border-color)] rounded-xl">
              No zones matching &quot;{search}&quot;
            </div>
          )}
        </div>
      )}
    </div>
  );
}
