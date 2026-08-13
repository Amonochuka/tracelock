"use client";

import { useEffect, useState } from 'react';
import { API_URL } from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import { Activity } from 'lucide-react';

interface ZoneAnalytics {
  day_of_week: number;
  hour: number;
  entry_count: number;
}

export default function ZoneAnalyticsChart({ zoneId }: { zoneId: number }) {
  const { token } = useAuth();
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!token || !zoneId) return;

    const fetchAnalytics = async () => {
      setLoading(true);
      setError('');
      try {
        const res = await fetch(`${API_URL}/admin/zones/${zoneId}/analytics`, {
          headers: { Authorization: `Bearer ${token}` }
        });
        if (!res.ok) throw new Error('Failed to fetch analytics');
        const rawData: ZoneAnalytics[] = await res.json();
        
        // Group by hour (0-23)
        const hourlyData = new Array(24).fill(0).map((_, i) => ({
          hour: `${i}:00`,
          entries: 0
        }));

        (rawData || []).forEach(record => {
          if (record.hour >= 0 && record.hour < 24) {
            hourlyData[record.hour].entries += record.entry_count;
          }
        });

        setData(hourlyData);
      } catch (err: unknown) {
        if (err instanceof Error) setError(err.message);
        else setError('An error occurred');
      } finally {
        setLoading(false);
      }
    };

    fetchAnalytics();
  }, [token, zoneId]);

  if (loading) {
    return <div className="h-64 flex items-center justify-center text-secondary text-sm border border-dashed border-[var(--border-color)] rounded-xl">Loading analytics...</div>;
  }

  if (error) {
    return <div className="h-64 flex items-center justify-center text-danger text-sm border border-dashed border-[var(--danger-primary)] rounded-xl">{error}</div>;
  }

  return (
    <div className="card w-full h-full flex flex-col" style={{ minHeight: '300px' }}>
      <div className="flex items-center gap-2 mb-6">
        <Activity size={18} className="text-accent" />
        <h3 className="m-0 text-sm font-semibold">Peak Entry Times</h3>
      </div>
      
      {data.every(d => d.entries === 0) ? (
        <div className="flex-1 flex items-center justify-center text-secondary text-sm">
          No access data recorded for this zone yet.
        </div>
      ) : (
        <div className="flex-1 w-full min-h-[200px]">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" vertical={false} />
              <XAxis 
                dataKey="hour" 
                tick={{ fill: '#8b949e', fontSize: 10 }}
                axisLine={{ stroke: 'rgba(255,255,255,0.1)' }}
                tickLine={false}
                interval={2}
              />
              <YAxis 
                tick={{ fill: '#8b949e', fontSize: 10 }}
                axisLine={false}
                tickLine={false}
                allowDecimals={false}
              />
              <Tooltip 
                cursor={{ fill: 'rgba(255,255,255,0.05)' }}
                contentStyle={{ 
                  backgroundColor: 'var(--bg-main)', 
                  border: '1px solid var(--border-color)',
                  borderRadius: '8px',
                  fontSize: '12px'
                }}
                itemStyle={{ color: 'var(--accent-primary)' }}
              />
              <Bar 
                dataKey="entries" 
                fill="var(--accent-primary)" 
                radius={[4, 4, 0, 0]} 
                name="Entries"
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}
