import Link from 'next/link';

export default function Home() {
  return (
    <main style={{padding:'2rem',fontFamily:'system-ui'}}>
      <h1>OpenSig Admin</h1>
      <p>Welcome to the OpenSig email signature management system.</p>
      
      <h2>Management</h2>
      <ul>
        <li><Link href="/schedules">📅 Schedules</Link> - Manage time windows and recurrence patterns</li>
        <li><Link href="/rules">📋 Rules</Link> - Define signature rules with conditions</li>
      </ul>

      <h2>System</h2>
      <ul>
        <li><a href="/api/health">API Health</a> (if configured)</li>
      </ul>

      <div style={{ marginTop: '2rem', padding: '1rem', background: '#f0f0f0', borderRadius: '4px' }}>
        <h3>M4: Rules Engine & Schedules</h3>
        <p>This implementation includes:</p>
        <ul>
          <li>Rules with conditions (sender/recipient/message type)</li>
          <li>Schedules with time ranges and recurrence patterns</li>
          <li>Priority and exclusivity support</li>
          <li>Rule evaluation engine for test messages</li>
        </ul>
      </div>
    </main>
  );
}
