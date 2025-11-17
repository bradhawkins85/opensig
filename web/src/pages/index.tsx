import Link from 'next/link';

export default function Home() {
  return (
    <main style={{padding:'2rem',fontFamily:'system-ui'}}>
      <h1>OpenSig (skeleton)</h1>
      <p>Welcome. This is a minimal admin UI shell.</p>
      <ul>
        <li><a href="/api/health">Local API health proxy</a> (if configured)</li>
        <li><Link href="https://github.com/">GitHub</Link></li>
      </ul>
    </main>
  );
}
