"use client";

import { useEffect, useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { Users, AlertTriangle } from 'lucide-react';

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

  // Fetch initial state
  useEffect(() => {
    if (!token) return;

    const fetchOccupancy = async () => {
      try {
        const res = await fetch(`http://${window.location.hostname}:8080/zones/occupancy`, {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!res.ok) throw new Error('Failed to fetch occupancy');
        const data = await res.json();
        
        const occMap: Record<number, ZoneOccupancy> = {};
        if (data && data.length > 0) {
          data.forEach((z: { id: number, name: string, max_capacity: number, active_count: number }) => {
            occMap[z.id] = {
              zone: {
                id: z.id,
                name: z.name,
                max_capacity: z.max_capacity
              },
              active_count: z.active_count || 0
            };
          });
        }
        setOccupancy(occMap);
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

    fetchOccupancy();
  }, [token]);

  // Connect to WebSocket for live updates
  useEffect(() => {
    if (!token) return;

    // Send JWT as a query param since JS WebSockets don't support custom headers easily
    // We would need the backend to support picking up token from query param, or use cookie.
    // Assuming backend ws endpoint might need auth. If it's authenticated via router, 
    // we must pass token in query param for standard WebSocket.
    const ws = new WebSocket(`ws://${window.location.hostname}:8080/ws/zones?token=${token}`);

    ws.onopen = () => {
      setWsConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const update: ZoneOccupancy = JSON.parse(event.data);
        if (update && update.zone && update.zone.id) {
          setOccupancy(prev => ({
            ...prev,
            [update.zone.id]: update
          }));
        }
      } catch (e) {
        console.error("Failed to parse WS message", e);
      }
    };

    ws.onclose = () => {
      setWsConnected(false);
    };

    return () => {
      ws.close();
    };
  }, [token]);

  if (loading) return <div className="text-secondary">Loading dashboard...</div>;

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1>Global Zone Occupancy</h1>
          <p className="text-secondary mt-2">Real-time facility monitoring</p>
        </div>
        <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-[var(--bg-card)] border border-[var(--border-color)]">
          <div className={`status-dot ${wsConnected ? 'status-active' : 'status-danger'}`}></div>
          <span className="text-xs font-medium text-secondary">
            {wsConnected ? 'LIVE FEED CONNECTED' : 'FEED DISCONNECTED'}
          </span>
        </div>
      </div>

      {error && (
        <div className="p-4 mb-8 bg-red-500/10 border border-red-500 rounded-lg text-red-500">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        {Object.values(occupancy).map((occ) => {
          const isAtCapacity = occ.zone.max_capacity > 0 && occ.active_count >= occ.zone.max_capacity;
          const isNearCapacity = occ.zone.max_capacity > 0 && (occ.active_count / occ.zone.max_capacity) > 0.8;
          
          let statusColor = 'var(--accent-primary)';
          if (isAtCapacity) {
            statusColor = 'var(--danger-primary)';
          } else if (isNearCapacity) {
            statusColor = 'var(--warning-primary)';
          }

          const percent = occ.zone.max_capacity > 0 
            ? Math.min(100, Math.round((occ.active_count / occ.zone.max_capacity) * 100))
            : 0;

          return (
            <div key={occ.zone.id} className="card flex flex-col gap-4 relative overflow-hidden">
              {isAtCapacity && (
                <div className="absolute top-0 right-0 w-16 h-16 pointer-events-none">
                  <div className="absolute top-4 right-4 animate-pulse">
                    <AlertTriangle size={20} className="text-danger" />
                  </div>
                </div>
              )}
              
              <div className="flex justify-between items-start">
                <div>
                  <h3 className="mb-1">{occ.zone.name}</h3>
                  <span className="text-xs text-secondary font-medium px-2 py-1 rounded" style={{ background: 'rgba(255,255,255,0.05)' }}>
                    ZONE {occ.zone.id}
                  </span>
                </div>
              </div>

              <div className="mt-4 flex items-end justify-between">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <Users size={16} className="text-secondary" />
                    <span className="text-sm text-secondary">Active Users</span>
                  </div>
                  <div className="flex items-baseline gap-2">
                    <span className="text-3xl font-bold mono" style={{ color: statusColor }}>
                      {occ.active_count}
                    </span>
                    {occ.zone.max_capacity > 0 && (
                      <span className="text-secondary mono text-sm">/ {occ.zone.max_capacity}</span>
                    )}
                  </div>
                </div>
              </div>

              {occ.zone.max_capacity > 0 && (
                <div className="progress-bg mt-2">
                  <div 
                    className="progress-fill" 
                    style={{ 
                      width: `${percent}%`,
                      backgroundColor: statusColor
                    }}
                  />
                </div>
              )}
            </div>
          );
        })}

        {Object.keys(occupancy).length === 0 && !loading && (
          <div className="col-span-full p-8 text-center text-secondary border border-dashed border-[var(--border-color)] rounded-xl">
            No zones configured in the system.
          </div>
        )}
      </div>
    </div>
  );
}
