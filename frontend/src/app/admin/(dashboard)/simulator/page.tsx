"use client";

import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/context/AuthContext';
import { API_URL } from '@/lib/api';
import { Cpu, Fingerprint, Activity, TerminalSquare, Plus, Save, ShieldCheck, Trash2, RefreshCw, Key, Link, Settings, Database } from 'lucide-react';

type Zone = { id: number; name: string };
type User = { id: number; name: string; email: string };
type Device = { id: number; zone_id: number; name: string; type: string; serial: string; active: boolean };
type Credential = { id: number; user_id: number; entry_method: string; credential_hash: string; enrolled_at: string; revoked: boolean };
type AccessEvent = { id: number; user_id: number; zone_id: number; action: string; status: string; reason?: string; timestamp: string; entry_method: string; hash: string };

export default function SimulatorPage() {
  const { token } = useAuth();

  const [activeTab, setActiveTab] = useState<'setup' | 'records'>('setup');
  const [zones, setZones] = useState<Zone[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [allDevices, setAllDevices] = useState<Device[]>([]);
  const [allCredentials, setAllCredentials] = useState<(Credential & { user_name: string })[]>([]);
  const [recentEvents, setRecentEvents] = useState<AccessEvent[]>([]);

  const [activeDeviceId, setActiveDeviceId] = useState('');
  const [activeDeviceType, setActiveDeviceType] = useState('');
  const [activeCredentialHash, setActiveCredentialHash] = useState('');
  const [activeCredentialMethod, setActiveCredentialMethod] = useState('');

  const [devZoneId, setDevZoneId] = useState('');
  const [devName, setDevName] = useState('');
  const [devType, setDevType] = useState('fingerprint');
  const [devSerial, setDevSerial] = useState('');
  const [devLoading, setDevLoading] = useState(false);
  const [devFeedback, setDevFeedback] = useState({ type: '', msg: '' });

  const [credUserId, setCredUserId] = useState('');
  const [credMethod, setCredMethod] = useState('fingerprint');
  const [credHash, setCredHash] = useState('');
  const [credLoading, setCredLoading] = useState(false);
  const [credFeedback, setCredFeedback] = useState({ type: '', msg: '' });

  const [accessUserId, setAccessUserId] = useState('');
  const [accessZoneId, setAccessZoneId] = useState('');
  const [accessLoading, setAccessLoading] = useState(false);
  const [accessFeedback, setAccessFeedback] = useState({ type: '', msg: '' });

  const [simAction, setSimAction] = useState('enter');
  const [simLoading, setSimLoading] = useState(false);
  const [simOutput, setSimOutput] = useState('');

  const [eventsZoneId, setEventsZoneId] = useState('');

  const headers = { Authorization: `Bearer ${token}` };

  const fetchDevices = useCallback(async (zoneList: Zone[]) => {
    if (!token || zoneList.length === 0) return;
    const results = await Promise.all(
      zoneList.map(z =>
        fetch(`${API_URL}/admin/zones/${z.id}/devices`, { headers: { Authorization: `Bearer ${token}` } })
          .then(r => r.ok ? r.json() : [])
          .then((devs: Device[]) => (devs || []).map(d => ({ ...d, zone_id: z.id })))
      )
    );
    setAllDevices(results.flat());
  }, [token]);

  const fetchCredentials = useCallback(async (userList: User[]) => {
    if (!token || userList.length === 0) return;
    const results = await Promise.all(
      userList.map(u =>
        fetch(`${API_URL}/admin/users/${u.id}/credentials`, { headers: { Authorization: `Bearer ${token}` } })
          .then(r => r.ok ? r.json() : [])
          .then((creds: Credential[]) => (creds || []).map(c => ({ ...c, user_name: u.name })))
      )
    );
    setAllCredentials(results.flat());
  }, [token]);

  const fetchEvents = useCallback(async (zoneId: string) => {
    if (!token || !zoneId) return;
    const res = await fetch(`${API_URL}/zones/${zoneId}/events`, { headers: { Authorization: `Bearer ${token}` } });
    if (res.ok) {
      const data = await res.json();
      setRecentEvents((data.events || []).slice(0, 15));
    }
  }, [token]);

  useEffect(() => {
    if (!token) return;
    Promise.all([
      fetch(`${API_URL}/zones`, { headers }).then(r => r.ok ? r.json() : []),
      fetch(`${API_URL}/admin/users`, { headers }).then(r => r.ok ? r.json() : []),
    ]).then(([zList, uList]) => {
      const zl = zList || [];
      const ul = uList || [];
      setZones(zl);
      setUsers(ul);
      fetchDevices(zl);
      fetchCredentials(ul);
      if (zl.length > 0) {
        setEventsZoneId(zl[0].id.toString());
        fetchEvents(zl[0].id.toString());
      }
    });
  }, [token]);

  useEffect(() => {
    if (eventsZoneId) fetchEvents(eventsZoneId);
  }, [eventsZoneId]);

  const feedback = (type: string, msg: string) => ({ type, msg });

  const handleCreateDevice = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!devZoneId) return setDevFeedback(feedback('error', 'Please select a zone'));
    setDevLoading(true); setDevFeedback({ type: '', msg: '' });
    try {
      const res = await fetch(`${API_URL}/admin/zones/${devZoneId}/devices`, {
        method: 'POST', headers: { 'Content-Type': 'application/json', ...headers },
        body: JSON.stringify({ name: devName, type: devType, serial: devSerial })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to create device');
      setActiveDeviceId(data.id.toString());
      setActiveDeviceType(devType);
      setDevFeedback(feedback('success', `Device created (ID: ${data.id}) — loaded into Step 4`));
      setDevName(''); setDevSerial('');
      fetchDevices(zones);
    } catch (err: any) { setDevFeedback(feedback('error', err.message)); }
    finally { setDevLoading(false); }
  };

  const handleEnrolCredential = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!credUserId) return setCredFeedback(feedback('error', 'Please select a user'));
    setCredLoading(true); setCredFeedback({ type: '', msg: '' });
    try {
      const res = await fetch(`${API_URL}/admin/users/${credUserId}/credentials`, {
        method: 'POST', headers: { 'Content-Type': 'application/json', ...headers },
        body: JSON.stringify({ entry_method: credMethod, credential_hash: credHash })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to enrol credential');
      setActiveCredentialHash(credHash);
      setActiveCredentialMethod(credMethod);
      setCredFeedback(feedback('success', 'Credential enrolled — loaded into Step 4'));
      setCredHash('');
      fetchCredentials(users);
    } catch (err: any) { setCredFeedback(feedback('error', err.message)); }
    finally { setCredLoading(false); }
  };

  const handleGrantAccess = async (e: React.FormEvent) => {
    e.preventDefault();
    setAccessLoading(true); setAccessFeedback({ type: '', msg: '' });
    try {
      const res = await fetch(`${API_URL}/admin/access`, {
        method: 'POST', headers: { 'Content-Type': 'application/json', ...headers },
        body: JSON.stringify({ user_id: parseInt(accessUserId, 10), zone_id: parseInt(accessZoneId, 10) })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to grant access');
      const userName = users.find(u => u.id === parseInt(accessUserId, 10))?.name || accessUserId;
      const zoneName = zones.find(z => z.id === parseInt(accessZoneId, 10))?.name || accessZoneId;
      setAccessFeedback(feedback('success', `${userName} → ${zoneName} access granted`));
    } catch (err: any) { setAccessFeedback(feedback('error', err.message)); }
    finally { setAccessLoading(false); }
  };

  const handleSimulate = async (e: React.FormEvent) => {
    e.preventDefault();
    setSimLoading(true);
    if (activeDeviceType && activeCredentialMethod && activeDeviceType !== activeCredentialMethod) {
      setSimOutput(`ERROR: Type Mismatch\n\nDevice is '${activeDeviceType}' but credential is '${activeCredentialMethod}'.`);
      setSimLoading(false); return;
    }
    setSimOutput('Sending request...');
    try {
      const payload = { device_id: parseInt(activeDeviceId, 10), credential_hash: activeCredentialHash, action: simAction };
      const res = await fetch(`${API_URL}/admin/simulate-device`, {
        method: 'POST', headers: { 'Content-Type': 'application/json', ...headers },
        body: JSON.stringify(payload)
      });
      let data: any;
      try { data = await res.json(); } catch { data = { error: res.statusText }; }
      if (res.status === 401) {
        setSimOutput(`ERROR: Session Expired\n\nYour login session has expired. Please log out and log back in.`);
        setSimLoading(false); return;
      }
      setSimOutput(`STATUS: ${res.status} ${res.statusText}\n\nPAYLOAD:\n${JSON.stringify(payload, null, 2)}\n\nRESPONSE:\n${JSON.stringify(data, null, 2)}`);
      if (res.ok && eventsZoneId) setTimeout(() => fetchEvents(eventsZoneId), 800);
    } catch (err: any) { setSimOutput(`ERROR:\n${err.message}`); }
    finally { setSimLoading(false); }
  };

  const handleDeleteDevice = async (deviceId: number) => {
    if (!confirm(`Delete device ID ${deviceId}?`)) return;
    await fetch(`${API_URL}/admin/devices/${deviceId}`, { method: 'DELETE', headers });
    if (activeDeviceId === deviceId.toString()) setActiveDeviceId('');
    fetchDevices(zones);
  };

  const alertClass = (type: string) =>
    `p-3 mb-4 text-sm rounded border ${type === 'error' ? 'bg-[rgba(255,77,106,0.1)] border-[var(--danger-primary)] text-danger' : 'bg-[rgba(0,212,170,0.1)] border-accent text-accent'}`;

  const zoneName = (id: number) => zones.find(z => z.id === id)?.name || `Zone ${id}`;
  const userName = (id: number) => users.find(u => u.id === id)?.name || `User ${id}`;

  return (
    <div className="space-y-6">
      {/* Description Text */}
      <p className="text-secondary text-sm">
        Set up mock hardware and test backend device integration securely.
      </p>

      {/* Tabs */}
      <div className="flex gap-6 mb-6 border-b border-[var(--border-color)]">
        <button
          onClick={() => setActiveTab('setup')}
          type="button"
          className={`flex items-center gap-2 px-1 pb-3 text-sm font-medium border-b-2 transition-all ${
            activeTab === 'setup'
              ? 'border-[var(--accent-primary)] text-white'
              : 'border-transparent text-gray-500 hover:text-gray-300'
          }`}
        >
          <Settings size={14} />
          Setup
        </button>
        <button
          onClick={() => setActiveTab('records')}
          type="button"
          className={`flex items-center gap-2 px-1 pb-3 text-sm font-medium border-b-2 transition-all ${
            activeTab === 'records'
              ? 'border-[var(--accent-primary)] text-white'
              : 'border-transparent text-gray-500 hover:text-gray-300'
          }`}
        >
          <Database size={14} />
          Records
        </button>
      </div>


      {activeTab === 'setup' && <>
        {/* Setup grid */}
        <div className="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-xl p-6">

          {/* Step 1 */}
          <div className="card">
            <div className="flex items-center gap-2 mb-4">
              <Cpu size={20} className="text-accent" />
              <h3 className="m-0">1. Register Device</h3>
            </div>
            {devFeedback.msg && <div className={alertClass(devFeedback.type)}>{devFeedback.msg}</div>}
            <form onSubmit={handleCreateDevice} className="space-y-4">
              <div className="form-group">
                <label className="form-label">Zone</label>
                <select className="input" value={devZoneId} onChange={e => setDevZoneId(e.target.value)} required>
                  <option value="">Select a zone...</option>
                  {zones.map(z => <option key={z.id} value={z.id}>{z.name} (ID: {z.id})</option>)}
                </select>
              </div>
              <div className="form-group">
                <label className="form-label">Device Name</label>
                <input className="input" value={devName} onChange={e => setDevName(e.target.value)} placeholder="Main Entrance Scanner" required />
              </div>
              <div className="form-group">
                <label className="form-label">Device Type</label>
                <select className="input" value={devType} onChange={e => setDevType(e.target.value)}>
                  <option value="fingerprint">Fingerprint</option>
                  <option value="face">Face Recognition</option>
                  <option value="iris">Iris Scanner</option>
                  <option value="card">Card Reader</option>
                  <option value="pin">PIN Pad</option>
                </select>
              </div>
              <div className="form-group">
                <label className="form-label">Serial Number</label>
                <input className="input mono" value={devSerial} onChange={e => setDevSerial(e.target.value)} placeholder="SN-12345" required />
              </div>
              <button type="submit" className="btn btn-primary w-full" disabled={devLoading}>
                <Plus size={16} className="mr-2" />{devLoading ? 'Registering...' : 'Register Device'}
              </button>
            </form>
          </div>

          {/* Step 2 */}
          <div className="card">
            <div className="flex items-center gap-2 mb-4">
              <Fingerprint size={20} className="text-accent" />
              <h3 className="m-0">2. Enrol Credential</h3>
            </div>
            {credFeedback.msg && <div className={alertClass(credFeedback.type)}>{credFeedback.msg}</div>}
            <form onSubmit={handleEnrolCredential} className="space-y-4">
              <div className="form-group">
                <label className="form-label">Personnel</label>
                <select className="input" value={credUserId} onChange={e => setCredUserId(e.target.value)} required>
                  <option value="">Select a user...</option>
                  {users.map(u => <option key={u.id} value={u.id}>{u.name} ({u.email})</option>)}
                </select>
              </div>
              <div className="form-group">
                <label className="form-label">Entry Method</label>
                <select className="input" value={credMethod} onChange={e => setCredMethod(e.target.value)}>
                  <option value="fingerprint">Fingerprint</option>
                  <option value="face">Face Recognition</option>
                  <option value="iris">Iris Scanner</option>
                  <option value="card">Card Reader</option>
                  <option value="pin">PIN Pad</option>
                </select>
              </div>
              <div className="form-group">
                <label className="form-label">Raw Credential Value</label>
                <input className="input mono" value={credHash} onChange={e => setCredHash(e.target.value)} placeholder="test_fingerprint_001" required />
              </div>
              <div className="text-secondary text-xs mb-2">Backend hashes this before storing.</div>
              <button type="submit" className="btn btn-primary w-full" disabled={credLoading}>
                <Save size={16} className="mr-2" />{credLoading ? 'Enrolling...' : 'Enrol Credential'}
              </button>
            </form>
          </div>
        </div>

        {/* Step 3 */}
        <div className="card mb-6">
          <div className="flex items-center gap-2 mb-4">
            <ShieldCheck size={20} className="text-accent" />
            <h3 className="m-0">3. Grant Zone Access</h3>
          </div>
          <p className="text-secondary text-sm mb-4">Without this, the scan returns &quot;access denied&quot;.</p>
          {accessFeedback.msg && <div className={alertClass(accessFeedback.type)}>{accessFeedback.msg}</div>}
          <form onSubmit={handleGrantAccess} className="grid grid-cols-1 md:grid-cols-3 gap-4 items-end">
            <div className="form-group">
              <label className="form-label">Personnel</label>
              <select className="input" value={accessUserId} onChange={e => setAccessUserId(e.target.value)} required>
                <option value="">Select a user...</option>
                {users.map(u => <option key={u.id} value={u.id}>{u.name}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">Zone</label>
              <select className="input" value={accessZoneId} onChange={e => setAccessZoneId(e.target.value)} required>
                <option value="">Select a zone...</option>
                {zones.map(z => <option key={z.id} value={z.id}>{z.name}</option>)}
              </select>
            </div>
            <button type="submit" className="btn btn-primary h-11" disabled={accessLoading}>
              <ShieldCheck size={16} className="mr-2" />{accessLoading ? 'Granting...' : 'Grant Access'}
            </button>
          </form>
        </div>

        {/* Step 4: Simulate */}
        <div className="card mb-6">
          <div className="flex items-center gap-2 mb-4">
            <Activity size={20} className="text-accent" />
            <h3 className="m-0">4. Hardware Authentication Test</h3>
          </div>
          <p className="text-secondary text-sm mb-6">
            Fields auto-fill from Steps 1 &amp; 2. You can also type any Device ID and credential manually.
          </p>
          <form onSubmit={handleSimulate} className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end mb-6">
            <div className="form-group">
              <label className="form-label">Device ID</label>
              <input type="number" className="input mono" value={activeDeviceId} onChange={e => setActiveDeviceId(e.target.value)} required />
            </div>
            <div className="form-group">
              <label className="form-label">Credential Hash</label>
              <input type="text" className="input mono" value={activeCredentialHash} onChange={e => setActiveCredentialHash(e.target.value)} required />
            </div>
            <div className="form-group">
              <label className="form-label">Action</label>
              <select className="input" value={simAction} onChange={e => setSimAction(e.target.value)}>
                <option value="enter">Enter Zone</option>
                <option value="exit">Exit Zone</option>
              </select>
            </div>
            <button type="submit" className="btn btn-primary h-11" disabled={simLoading || !activeDeviceId || !activeCredentialHash}>
              <TerminalSquare size={16} className="mr-2" />Simulate Scan
            </button>
          </form>
          <div className="bg-black border border-[rgba(255,255,255,0.1)] rounded-lg p-4 font-mono text-sm overflow-x-auto min-h-[100px]">
            {simOutput ? (
              <pre className={simOutput.includes('ERROR') ? 'text-danger' : 'text-[#00d4aa]'}>{simOutput}</pre>
            ) : (
              <div className="text-[rgba(255,255,255,0.3)]">Ready for payload test...</div>
            )}
          </div>
        </div>

      </> /* end setup tab */}

      {activeTab === 'records' && <>
        {/* Data panels */}
        <div className="bg-[var(--bg-card)] border border-[var(--border-color)] rounded-xl p-6">

          {/* Registered Devices */}
          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <Cpu size={20} className="text-accent" />
                <h3 className="m-0">Registered Devices</h3>
              </div>
              <button onClick={() => fetchDevices(zones)} className="btn text-xs px-3 py-1.5 h-auto">
                <RefreshCw size={13} className="mr-1" />Refresh
              </button>
            </div>
            {allDevices.length === 0 ? (
              <p className="text-secondary text-sm">No devices registered yet.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="table w-full text-sm">
                  <thead>
                    <tr>
                      <th>ID</th><th>Name</th><th>Type</th><th>Zone</th><th>Serial</th><th>Status</th><th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {allDevices.map(d => (
                      <tr key={d.id} className={activeDeviceId === d.id.toString() ? 'bg-[rgba(0,212,170,0.05)]' : ''}>
                        <td className="mono font-bold text-accent">{d.id}</td>
                        <td>{d.name}</td>
                        <td><span className="badge">{d.type}</span></td>
                        <td>{zoneName(d.zone_id)}</td>
                        <td className="mono text-xs text-secondary">{d.serial}</td>
                        <td>
                          <span className={`badge ${d.active ? 'text-accent' : 'text-danger'}`}>
                            {d.active ? 'Active' : 'Inactive'}
                          </span>
                        </td>
                        <td>
                          <div className="flex gap-2">
                            <button
                              onClick={() => { setActiveDeviceId(d.id.toString()); setActiveDeviceType(d.type); setActiveTab('setup'); }}
                              className="btn text-xs px-2 py-1 h-auto text-accent"
                              title="Load into Setup"
                            >Use</button>
                            <button
                              onClick={() => handleDeleteDevice(d.id)}
                              className="btn text-xs px-2 py-1 h-auto text-danger"
                              title="Delete device"
                            ><Trash2 size={13} /></button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          {/* Enrolled Credentials */}
          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <Key size={20} className="text-accent" />
                <h3 className="m-0">Enrolled Credentials</h3>
              </div>
              <button onClick={() => fetchCredentials(users)} className="btn text-xs px-3 py-1.5 h-auto">
                <RefreshCw size={13} className="mr-1" />Refresh
              </button>
            </div>
            {allCredentials.length === 0 ? (
              <p className="text-secondary text-sm">No credentials enrolled yet.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="table w-full text-sm">
                  <thead>
                    <tr><th>User</th><th>Method</th><th>Stored Hash</th><th>Status</th><th></th></tr>
                  </thead>
                  <tbody>
                    {allCredentials.map(c => (
                      <tr key={c.id} className={activeCredentialHash === c.credential_hash ? 'bg-[rgba(0,212,170,0.05)]' : ''}>
                        <td>{c.user_name}</td>
                        <td><span className="badge">{c.entry_method}</span></td>
                        <td className="mono text-xs text-secondary max-w-[140px] truncate" title={c.credential_hash}>
                          {c.credential_hash.slice(0, 12)}…
                        </td>
                        <td>
                          <span className={`badge ${c.revoked ? 'text-danger' : 'text-accent'}`}>
                            {c.revoked ? 'Revoked' : 'Active'}
                          </span>
                        </td>
                        <td>
                          <button
                            onClick={() => { setActiveCredentialHash(c.credential_hash); setActiveCredentialMethod(c.entry_method); setActiveTab('setup'); }}
                            className="btn text-xs px-2 py-1 h-auto text-accent"
                            title="Load into Setup"
                          >Use</button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>

        {/* Hash Chain Events */}
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <Link size={20} className="text-accent" />
              <h3 className="m-0">Hash Chain Events</h3>
            </div>
            <div className="flex items-center gap-3">
              <select className="input text-sm h-9 py-0" value={eventsZoneId} onChange={e => setEventsZoneId(e.target.value)}>
                {zones.map(z => <option key={z.id} value={z.id}>{z.name}</option>)}
              </select>
              <button onClick={() => fetchEvents(eventsZoneId)} className="btn text-xs px-3 py-1.5 h-auto">
                <RefreshCw size={13} className="mr-1" />Refresh
              </button>
            </div>
          </div>
          {recentEvents.length === 0 ? (
            <p className="text-secondary text-sm">No events recorded for this zone yet.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="table w-full text-sm">
                <thead>
                  <tr><th>Time</th><th>User</th><th>Action</th><th>Method</th><th>Status</th><th>Reason</th><th>Hash</th></tr>
                </thead>
                <tbody>
                  {recentEvents.map(ev => (
                    <tr key={ev.id}>
                      <td className="text-xs text-secondary whitespace-nowrap">
                        {new Date(ev.timestamp).toLocaleString()}
                      </td>
                      <td>{userName(ev.user_id)}</td>
                      <td><span className="badge">{ev.action}</span></td>
                      <td className="text-xs">{ev.entry_method || '—'}</td>
                      <td>
                        <span className={`badge ${ev.status === 'allowed' ? 'text-accent' : 'text-danger'}`}>
                          {ev.status}
                        </span>
                      </td>
                      <td className="text-xs text-secondary">{ev.reason || '—'}</td>
                      <td className="mono text-xs text-secondary" title={ev.hash}>{ev.hash.slice(0, 10)}…</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </> /* end records tab */}
    </div>
  );
}
