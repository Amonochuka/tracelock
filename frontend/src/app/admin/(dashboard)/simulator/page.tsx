"use client";

import { useState, useEffect } from 'react';
import { useAuth } from '@/context/AuthContext';
import { API_URL } from '@/lib/api';
import { Cpu, Fingerprint, Activity, TerminalSquare, Plus, Save } from 'lucide-react';

export default function SimulatorPage() {
  const { token } = useAuth();
  
  // Data for selectors
  const [zones, setZones] = useState<{ id: number, name: string }[]>([]);
  const [users, setUsers] = useState<{ id: number, name: string, email: string }[]>([]);
  
  // Shared state between steps
  const [activeDeviceId, setActiveDeviceId] = useState('');
  const [activeCredentialHash, setActiveCredentialHash] = useState('');

  // Step 1: Device Registration
  const [devZoneId, setDevZoneId] = useState('');
  const [devName, setDevName] = useState('');
  const [devType, setDevType] = useState('biometric_scanner');
  const [devSerial, setDevSerial] = useState('');
  const [devLoading, setDevLoading] = useState(false);
  const [devFeedback, setDevFeedback] = useState({ type: '', msg: '' });

  // Step 2: Credential Enrollment
  const [credUserId, setCredUserId] = useState('');
  const [credHash, setCredHash] = useState('');
  const [credLoading, setCredLoading] = useState(false);
  const [credFeedback, setCredFeedback] = useState({ type: '', msg: '' });

  // Step 3: Simulation
  const [simAction, setSimAction] = useState('enter');
  const [simLoading, setSimLoading] = useState(false);
  const [simOutput, setSimOutput] = useState('');

  useEffect(() => {
    if (!token) return;
    
    // Fetch zones and users for the setup forms
    Promise.all([
      fetch(`${API_URL}/zones`, { headers: { Authorization: `Bearer ${token}` } }),
      fetch(`${API_URL}/admin/users`, { headers: { Authorization: `Bearer ${token}` } })
    ]).then(async ([zRes, uRes]) => {
      if (zRes.ok) setZones(await zRes.json() || []);
      if (uRes.ok) setUsers(await uRes.json() || []);
    }).catch(console.error);
  }, [token]);

  const handleCreateDevice = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!devZoneId) return setDevFeedback({ type: 'error', msg: 'Please select a zone' });
    
    setDevLoading(true);
    setDevFeedback({ type: '', msg: '' });
    
    try {
      const res = await fetch(`${API_URL}/admin/zones/${devZoneId}/devices`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ name: devName, type: devType, serial: devSerial })
      });
      const data = await res.json();
      
      if (!res.ok) throw new Error(data.error || 'Failed to create device');
      
      setActiveDeviceId(data.id.toString());
      setDevFeedback({ type: 'success', msg: `Device created successfully (ID: ${data.id})` });
      setDevName('');
      setDevSerial('');
    } catch (err: any) {
      setDevFeedback({ type: 'error', msg: err.message });
    } finally {
      setDevLoading(false);
    }
  };

  const handleEnrolCredential = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!credUserId) return setCredFeedback({ type: 'error', msg: 'Please select a user' });
    
    setCredLoading(true);
    setCredFeedback({ type: '', msg: '' });
    
    try {
      const res = await fetch(`${API_URL}/admin/users/${credUserId}/credentials`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ entry_method: 'biometric', credential_hash: credHash })
      });
      const data = await res.json();
      
      if (!res.ok) throw new Error(data.error || 'Failed to enrol credential');
      
      setActiveCredentialHash(credHash);
      setCredFeedback({ type: 'success', msg: 'Credential enrolled successfully' });
      setCredHash('');
    } catch (err: any) {
      setCredFeedback({ type: 'error', msg: err.message });
    } finally {
      setCredLoading(false);
    }
  };

  const handleSimulate = async (e: React.FormEvent) => {
    e.preventDefault();
    setSimLoading(true);
    setSimOutput('Sending request to backend...');
    
    try {
      const payload = {
        device_id: parseInt(activeDeviceId, 10),
        credential_hash: activeCredentialHash,
        action: simAction
      };

      const res = await fetch(`${API_URL}/admin/simulate-device`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify(payload)
      });
      
      const data = await res.json();
      
      // Pretty print the JSON output
      setSimOutput(`STATUS: ${res.status} ${res.statusText}\n\nPAYLOAD SENT:\n${JSON.stringify(payload, null, 2)}\n\nRESPONSE:\n${JSON.stringify(data, null, 2)}`);
    } catch (err: any) {
      setSimOutput(`ERROR:\n${err.message}`);
    } finally {
      setSimLoading(false);
    }
  };

  return (
    <div>
      <div className="mb-8">
        <h1>Hardware Simulator</h1>
        <p className="text-secondary mt-2">Set up mock hardware and test backend device integration securely.</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        
        {/* Step 1: Device Registration */}
        <div className="card">
          <div className="flex items-center gap-2 mb-4">
            <Cpu size={20} className="text-accent" />
            <h3 className="m-0">1. Register Device</h3>
          </div>
          
          {devFeedback.msg && (
            <div className={`p-3 mb-4 text-sm rounded border ${devFeedback.type === 'error' ? 'bg-[rgba(255,77,106,0.1)] border-[var(--danger-primary)] text-danger' : 'bg-[rgba(0,212,170,0.1)] border-accent text-accent'}`}>
              {devFeedback.msg}
            </div>
          )}

          <form onSubmit={handleCreateDevice} className="space-y-4">
            <div className="form-group">
              <label className="form-label">Assign to Zone</label>
              <select className="input" value={devZoneId} onChange={(e) => setDevZoneId(e.target.value)} required>
                <option value="">Select a zone...</option>
                {zones.map(z => <option key={z.id} value={z.id}>{z.name} (ID: {z.id})</option>)}
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">Device Name</label>
              <input type="text" className="input" value={devName} onChange={(e) => setDevName(e.target.value)} placeholder="Main Entrance Scanner" required />
            </div>
            <div className="form-group">
              <label className="form-label">Serial Number</label>
              <input type="text" className="input mono" value={devSerial} onChange={(e) => setDevSerial(e.target.value)} placeholder="SN-12345" required />
            </div>
            <button type="submit" className="btn btn-primary w-full" disabled={devLoading}>
              <Plus size={16} className="mr-2" />
              {devLoading ? 'Registering...' : 'Register Device'}
            </button>
          </form>
        </div>

        {/* Step 2: Credential Enrollment */}
        <div className="card">
          <div className="flex items-center gap-2 mb-4">
            <Fingerprint size={20} className="text-accent" />
            <h3 className="m-0">2. Enrol Credential</h3>
          </div>

          {credFeedback.msg && (
            <div className={`p-3 mb-4 text-sm rounded border ${credFeedback.type === 'error' ? 'bg-[rgba(255,77,106,0.1)] border-[var(--danger-primary)] text-danger' : 'bg-[rgba(0,212,170,0.1)] border-accent text-accent'}`}>
              {credFeedback.msg}
            </div>
          )}

          <form onSubmit={handleEnrolCredential} className="space-y-4">
            <div className="form-group">
              <label className="form-label">Target Personnel</label>
              <select className="input" value={credUserId} onChange={(e) => setCredUserId(e.target.value)} required>
                <option value="">Select a user...</option>
                {users.map(u => <option key={u.id} value={u.id}>{u.name} ({u.email})</option>)}
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">Raw Credential Hash</label>
              <input type="text" className="input mono" value={credHash} onChange={(e) => setCredHash(e.target.value)} placeholder="test_fingerprint_001" required />
            </div>
            <div className="text-secondary text-xs mb-2">Simulates linking a physical biometric template to a user ID.</div>
            <button type="submit" className="btn btn-primary w-full" disabled={credLoading}>
              <Save size={16} className="mr-2" />
              {credLoading ? 'Enrolling...' : 'Enrol Credential'}
            </button>
          </form>
        </div>

      </div>

      {/* Step 3: Simulation execution */}
      <div className="card">
        <div className="flex items-center gap-2 mb-4">
          <Activity size={20} className="text-accent" />
          <h3 className="m-0">3. Hardware Authentication Payload Test</h3>
        </div>
        <p className="text-secondary text-sm mb-6">
          This secure proxy tests the backend `POST /devices/authenticate` logic without exposing the server`&apos;`s `DEVICE_API_KEY` to the browser.
        </p>

        <form onSubmit={handleSimulate} className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end mb-6">
          <div className="form-group">
            <label className="form-label">Device ID</label>
            <input type="number" className="input mono" value={activeDeviceId} onChange={(e) => setActiveDeviceId(e.target.value)} required />
          </div>
          <div className="form-group">
            <label className="form-label">Credential Hash</label>
            <input type="text" className="input mono" value={activeCredentialHash} onChange={(e) => setActiveCredentialHash(e.target.value)} required />
          </div>
          <div className="form-group">
            <label className="form-label">Action</label>
            <select className="input" value={simAction} onChange={(e) => setSimAction(e.target.value)}>
              <option value="enter">Enter Zone</option>
              <option value="exit">Exit Zone</option>
            </select>
          </div>
          <button type="submit" className="btn btn-primary h-11" disabled={simLoading || !activeDeviceId || !activeCredentialHash}>
            <TerminalSquare size={16} className="mr-2" />
            Simulate Scan
          </button>
        </form>

        <div className="bg-black border border-[rgba(255,255,255,0.1)] rounded-lg p-4 font-mono text-sm overflow-x-auto">
          {simOutput ? (
            <pre className={simOutput.includes('ERROR:') ? 'text-danger' : 'text-[#00d4aa]'}>
              {simOutput}
            </pre>
          ) : (
            <div className="text-[rgba(255,255,255,0.3)]">Ready for payload test...</div>
          )}
        </div>

      </div>
    </div>
  );
}
