import React, { useState } from 'react';

const providers = [
  { key: 'google', name: 'Google' },
  { key: 'slack', name: 'Slack' },
  { key: 'linkedin', name: 'LinkedIn' }
];

const Login: React.FC = () => {
  const [loadingProvider, setLoadingProvider] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleLogin = async (provider: string) => {
    setError(null);
    setLoadingProvider(provider);
    try {
      const resp = await fetch(`/api/v1/auth/login?provider=${provider}`);
      if (!resp.ok) {
        const text = await resp.text();
        throw new Error(text || 'Failed to get login URL');
      }
      const data: { url: string } = await resp.json();
      window.location.href = data.url;
    } catch (err: any) {
      setError(err.message || 'Unexpected error');
    } finally {
      setLoadingProvider(null);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center h-screen p-4">
      <h1 className="text-3xl font-bold mb-6">Sign In</h1>
      {error && <div className="text-red-600 mb-4">{error}</div>}
      <div className="space-y-3 w-full max-w-xs">
        {providers.map(p => (
          <button
            key={p.key}
            disabled={loadingProvider === p.key}
            onClick={() => handleLogin(p.key)}
            className="w-full py-2 px-4 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {loadingProvider === p.key ? `Redirecting to ${p.name}...` : `Sign in with ${p.name}`}
          </button>
        ))}
      </div>
    </div>
  );
};

export default Login;