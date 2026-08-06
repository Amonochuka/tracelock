"use client";

import { useEffect, useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { useParams } from 'next/navigation';
import { ArrowLeft, User, Activity, Clock } from 'lucide-react';
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

export default function ZoneDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const { token } = useAuth();
  const [zone, setZone] = useState<ZoneObj | null>(null);
  const [activeUsers, setActiveUsers] = useState<UserObj[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!token || !id) return;

    const fetchData = async () => {
      try {
        // Fetch zone details
        const zRes = await fetch(`${API_URL}/zones/${id}`, {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!zRes.ok) throw new Error('Failed to fetch zone details');
        const zData = await zRes.json();
        setZone(zData);

        // Fetch active users in this zone
        const uRes = await fetch(`${API_URL}/admin/zones/${id}/users`, {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!uRes.ok) throw new Error('Failed to fetch active users');
        const uData = await uRes.json();
        setActiveUsers(uData || []);
        
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

    fetchData();
  }, [token, id]);

  if (loading) return <div className="text-secondary">Loading zone telemetry...</div>;
  if (!zone) return <div className="text-danger">Zone not found</div>;

  return (
    <div>
      <div className="mb-8 flex items-center gap-4">
        <Link href="/admin" className="p-2 rounded hover:bg-[rgba(255,255,255,0.05)] text-secondary transition-colors">
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
        <div className="p-4 mb-8 bg-red-500/10 border border-red-500 rounded-lg text-red-500">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-8">
        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-2">
            <Activity size={16} />
            <span className="text-sm">Current Status</span>
          </div>
          <div className="text-2xl font-bold text-accent">SECURE</div>
        </div>
        
        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-2">
            <User size={16} />
            <span className="text-sm">Occupancy</span>
          </div>
          <div className="text-2xl font-bold mono">
            {activeUsers.length} <span className="text-secondary text-sm font-normal">/ {zone.max_capacity > 0 ? zone.max_capacity : '∞'}</span>
          </div>
        </div>
        
        <div className="card flex flex-col gap-2">
          <div className="flex items-center gap-2 text-secondary mb-2">
            <Clock size={16} />
            <span className="text-sm">Protocol</span>
          </div>
          <div className="text-sm font-medium">
            {zone.requires_exit_scan ? 'Strict (Entry/Exit Scan Required)' : 'Standard (Auto-Exit Enabled)'}
          </div>
        </div>
      </div>

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
                <td>
                  <span className="font-medium">{u.name}</span>
                </td>
                <td>
                  <span className="text-xs uppercase tracking-wider text-secondary">{u.role}</span>
                </td>
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
  );
}
