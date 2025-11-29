import React, { useState } from 'react';
import { useRouter } from 'next/router';
import Link from 'next/link';

export default function Login() {
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      if (!res.ok) {
        throw new Error('Login failed');
      }

      const data = await res.json();
      localStorage.setItem('token', data.token);
      
      // Decode token to get user ID (simple decode, not verification)
      const payload = JSON.parse(atob(data.token.split('.')[1]));
      localStorage.setItem('userId', payload.userId);
      localStorage.setItem('username', username);

      router.push('/');
    } catch (err) {
      setError('Invalid username or password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--bg-primary)] text-[var(--text-primary)]">
      <div className="max-w-md w-full p-8 bg-[var(--bg-secondary)] rounded-xl border border-[var(--border-primary)] shadow-lg">
        <h2 className="text-2xl font-bold mb-6 text-center">Login to GoMessenger</h2>
        
        {error && (
          <div className="mb-4 p-3 bg-[var(--accent-danger)]/10 text-[var(--accent-danger)] rounded-lg text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-4 py-2 bg-[var(--bg-tertiary)] border border-[var(--border-primary)] rounded-lg focus:outline-none focus:border-[var(--accent-primary)]"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2 bg-[var(--bg-tertiary)] border border-[var(--border-primary)] rounded-lg focus:outline-none focus:border-[var(--accent-primary)]"
              required
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full py-2 bg-[var(--accent-primary)] text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50"
          >
            {loading ? 'Logging in...' : 'Login'}
          </button>
        </form>

        <div className="mt-4 text-center text-sm text-[var(--text-secondary)]">
          Don't have an account?{' '}
          <Link href="/register" className="text-[var(--accent-primary)] hover:underline">
            Register
          </Link>
        </div>
      </div>
    </div>
  );
}
