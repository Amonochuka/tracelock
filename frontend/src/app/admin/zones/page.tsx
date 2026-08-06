"use client";

import { FormEvent, useCallback, useEffect, useState } from 'react';
import Link from 'next/link';
import { Map, Plus, Users } from 'lucide-react';
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

  const loadZones = useCallback(async () => {
    if (!token) return;
    const response = await fetch(`${API_URL}/zones`, { headers: { Authorization: `Bearer ${token}` } });
    if (!response.ok) throw new Error('Unable to load zones');
    setZones(await response.json());
  }, [token]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadZones().catch((caughtError: unknown) => setError(caughtError instanceof Error ? caughtError.message : 'Unable to load zones')).finally(() => setLoading(false));
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
      setZones((currentZones) => [...currentZones, data]);
      setName('');
      setDescription('');
      setMaxCapacity('');
      setRequiresExitScan(false);
    } catch (caughtError: unknown) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to create zone');
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <div className="text-secondary">Loading zones...</div>;

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div><h1>Zones</h1><p className="text-secondary mt-2">Configure monitored areas and their capacity rules.</p></div>
      </div>
      {error && <div className="p-4 mb-8 bg-red-500/10 border border-red-500 rounded-lg text-red-500">{error}</div>}
      <div className="grid grid-cols-1 md:grid-cols-2">
        {zones.map((zone) => <Link href={`/admin/zones/${zone.id}`} key={zone.id} className="card" style={{ textDecoration: 'none', color: 'inherit' }}><div className="flex items-center justify-between"><div><div className="flex items-center gap-2"><Map size={18} className="text-accent" /><h3>{zone.name}</h3></div><p className="text-secondary text-sm" style={{ marginTop: '0.75rem' }}>{zone.description || 'No description added.'}</p></div><span className="text-xs text-secondary mono">ZONE {zone.id}</span></div><div className="flex items-center gap-2 text-secondary text-sm" style={{ marginTop: '1.25rem' }}><Users size={16} />{zone.max_capacity > 0 ? `${zone.max_capacity} person capacity` : 'No capacity limit'} · {zone.requires_exit_scan ? 'Explicit exit required' : 'Auto-exit enabled'}</div></Link>)}
      </div>
      {zones.length === 0 && <div className="card text-center" style={{ marginBottom: '1.5rem' }}><Map size={28} className="text-accent" style={{ margin: '0 auto 0.75rem' }} /><h3>No zones yet</h3><p className="text-secondary text-sm" style={{ marginTop: '0.5rem' }}>Create the first zone below, then grant user access before running entry simulations.</p></div>}
      <form className="card" onSubmit={createZone}>
        <h3 className="flex items-center gap-2"><Plus size={18} className="text-accent" />Create zone</h3>
        <div className="grid grid-cols-1 md:grid-cols-2" style={{ marginTop: '1rem' }}>
          <div className="form-group"><label className="form-label" htmlFor="zone-name">Name</label><input id="zone-name" className="input" value={name} onChange={(event) => setName(event.target.value)} required /></div>
          <div className="form-group"><label className="form-label" htmlFor="zone-capacity">Maximum capacity (0 = unlimited)</label><input id="zone-capacity" className="input" type="number" min="0" value={maxCapacity} onChange={(event) => setMaxCapacity(event.target.value)} /></div>
          <div className="form-group"><label className="form-label" htmlFor="zone-description">Description</label><input id="zone-description" className="input" value={description} onChange={(event) => setDescription(event.target.value)} /></div>
          <label className="flex items-center gap-2 text-sm text-secondary" style={{ alignSelf: 'center' }}><input type="checkbox" checked={requiresExitScan} onChange={(event) => setRequiresExitScan(event.target.checked)} />Require explicit exit scan</label>
        </div>
        <button className="btn btn-primary" type="submit" disabled={saving}>{saving ? 'Creating...' : 'Create zone'}</button>
      </form>
    </div>
  );
}
