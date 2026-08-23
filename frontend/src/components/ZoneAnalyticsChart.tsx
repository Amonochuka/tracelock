"use client";

import { useEffect, useState } from 'react';
import { API_URL } from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import { Activity } from 'lucide-react';

interface ZoneAnalytics {
  day_of_week: number; // 0 = Sunday … 6 = Saturday (UTC)
  hour: number;        // 0–23 (UTC)
  entry_count: number;
}

const DAYS_MON_FIRST = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

// Convert a UTC (day, hour) pair into the browser's local (day, hour),
// carrying over into the previous or next day when the offset crosses
// midnight — e.g. 23:00 UTC Sunday becomes 02:00 Monday for UTC+3.
function toLocalDayHour(day: number, hour: number): { day: number; hour: number } {
  const shift = -new Date().getTimezoneOffset() / 60;
  const weekHour = (((day * 24 + hour + shift) % 168) + 168) % 168;
  return { day: Math.floor(weekHour / 24), hour: weekHour % 24 };
}

export default function ZoneAnalyticsChart({ zoneId }: { zoneId: number }) {
  const { token } = useAuth();
  const [rawData, setRawData] = useState<ZoneAnalytics[]>([]);
  const [view, setView] = useState<'bars' | 'heat'>('bars');
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
        setRawData((await res.json()) || []);
      } catch (err: unknown) {
        if (err instanceof Error) setError(err.message);
        else setError('An error occurred');
      } finally {
        setLoading(false);
      }
    };

    fetchAnalytics();
  }, [token, zoneId]);

  // Bars view data: entries grouped by local hour across all days.
  const hourlyData = new Array(24).fill(0).map((_, i) => ({ hour: `${i}:00`, entries: 0 }));
  // Heat view data: 7×24 matrix indexed [monFirstRow][localHour].
  const heat: number[][] = Array.from({ length: 7 }, () => new Array(24).fill(0));
  let maxCount = 0;

  rawData.forEach(record => {
    const { day, hour } = toLocalDayHour(record.day_of_week, record.hour);
    hourlyData[hour].entries += record.entry_count;
    const row = (day + 6) % 7; // Sunday(0) → last row, Monday(1) → first row
    heat[row][hour] += record.entry_count;
    if (heat[row][hour] > maxCount) maxCount = heat[row][hour];
  });

  const isEmpty = rawData.length === 0;

  const cellColor = (count: number): string => {
    if (count === 0) return 'rgba(255,255,255,0.04)';
    const t = count / maxCount;
    return `rgba(0, 212, 170, ${0.15 + 0.85 * t})`;
  };

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
        <h3 className="m-0 text-sm font-semibold">Zone Activity</h3>
        <div className="chart-toggle" role="tablist" aria-label="Analytics view">
          <button role="tab" aria-selected={view === 'bars'} className={view === 'bars' ? 'active' : ''} onClick={() => setView('bars')}>
            Peak Hours
          </button>
          <button role="tab" aria-selected={view === 'heat'} className={view === 'heat' ? 'active' : ''} onClick={() => setView('heat')}>
            Heatmap
          </button>
        </div>
      </div>

      {isEmpty ? (
        <div className="flex-1 flex items-center justify-center text-secondary text-sm">
          No access data recorded for this zone yet.
        </div>
      ) : view === 'bars' ? (
        <div className="flex-1 w-full min-h-[200px]">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={hourlyData} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
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
      ) : (
        <div className="heatmap-wrap">
          <div className="heatmap-grid">
            <div className="heatmap-corner"></div>
            {Array.from({ length: 24 }, (_, h) => (
              <div key={`h${h}`} className="heatmap-hour-label">{h % 3 === 0 ? h : ''}</div>
            ))}
            {DAYS_MON_FIRST.map((label, row) => (
              <HeatRow key={label} label={label} counts={heat[row]} maxCount={maxCount} cellColor={cellColor} />
            ))}
          </div>
          <div className="heatmap-legend">
            <span>Quiet</span>
            <div className="heatmap-legend-scale">
              {[0, 1, 2, 3, 4].map(step => (
                <span key={step} style={{ backgroundColor: cellColor(Math.round((step / 4) * maxCount)) }}></span>
              ))}
            </div>
            <span>Busy · peak {maxCount} entries</span>
          </div>
        </div>
      )}
    </div>
  );
}

function HeatRow({ label, counts, cellColor }: {
  label: string;
  counts: number[];
  maxCount: number;
  cellColor: (n: number) => string;
}) {
  return (
    <>
      <div className="heatmap-day-label">{label}</div>
      {counts.map((count, hour) => (
        <div
          key={`${label}-${hour}`}
          className="heatmap-cell"
          style={{ backgroundColor: cellColor(count) }}
          title={`${label} ${String(hour).padStart(2, '0')}:00 — ${count} ${count === 1 ? 'entry' : 'entries'}`}
        ></div>
      ))}
    </>
  );
}
