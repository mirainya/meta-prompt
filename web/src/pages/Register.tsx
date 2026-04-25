import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { api } from '../lib/api';

export default function Register() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await api.register(username, password);
      navigate('/login');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '注册失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="min-h-screen flex items-center justify-center relative overflow-hidden"
      style={{ background: 'var(--mp-login-bg)' }}
    >
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute w-[400px] h-[400px] rounded-full bg-white/15 -top-[100px] -right-[80px]" />
        <div className="absolute w-[300px] h-[300px] rounded-full bg-white/15 -bottom-[60px] -left-[60px]" />
        <div className="absolute w-[200px] h-[200px] rounded-full bg-white/10 top-[40%] left-[60%]" />
      </div>

      <div
        className="w-[400px] p-10 relative z-10"
        style={{
          background: 'rgba(255, 255, 255, 0.92)',
          backdropFilter: 'blur(20px)',
          borderRadius: '20px',
          boxShadow: '0 20px 60px rgba(0, 0, 0, 0.08)',
          border: '1px solid rgba(255, 255, 255, 0.6)',
        }}
      >
        <div className="text-center mb-8">
          <div
            className="w-14 h-14 rounded-2xl mx-auto mb-3 flex items-center justify-center text-white text-2xl font-bold"
            style={{ background: 'var(--mp-primary)' }}
          >
            M
          </div>
          <h2 className="text-[22px] font-bold" style={{ color: 'var(--mp-text-primary)' }}>
            创建账号
          </h2>
          <p className="text-sm mt-1" style={{ color: 'var(--mp-text-secondary)' }}>
            注册后即可使用提示词推演
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <input
            type="text"
            placeholder="用户名（2-50字符）"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="w-full h-11 px-4 rounded-xl text-sm outline-none transition-all"
            style={{
              border: '1px solid var(--mp-border-color)',
              background: 'var(--mp-primary-lighter)',
              color: 'var(--mp-text-primary)',
            }}
            onFocus={(e) => {
              e.target.style.borderColor = 'var(--mp-primary)';
              e.target.style.boxShadow = '0 0 0 3px rgba(154, 138, 189, 0.15)';
            }}
            onBlur={(e) => {
              e.target.style.borderColor = 'var(--mp-border-color)';
              e.target.style.boxShadow = 'none';
            }}
            required
            minLength={2}
            maxLength={50}
          />
          <input
            type="password"
            placeholder="密码（至少6位）"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full h-11 px-4 rounded-xl text-sm outline-none transition-all"
            style={{
              border: '1px solid var(--mp-border-color)',
              background: 'var(--mp-primary-lighter)',
              color: 'var(--mp-text-primary)',
            }}
            onFocus={(e) => {
              e.target.style.borderColor = 'var(--mp-primary)';
              e.target.style.boxShadow = '0 0 0 3px rgba(154, 138, 189, 0.15)';
            }}
            onBlur={(e) => {
              e.target.style.borderColor = 'var(--mp-border-color)';
              e.target.style.boxShadow = 'none';
            }}
            required
            minLength={6}
            maxLength={50}
          />

          {error && (
            <p className="text-sm px-3 py-2 rounded-lg" style={{ color: 'var(--mp-danger)', background: 'rgba(212, 122, 122, 0.08)' }}>
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full h-11 rounded-xl text-white font-semibold text-base tracking-widest transition-all disabled:opacity-50"
            style={{ background: 'var(--mp-primary)' }}
            onMouseEnter={(e) => { if (!loading) (e.target as HTMLElement).style.background = 'var(--mp-accent)'; }}
            onMouseLeave={(e) => { (e.target as HTMLElement).style.background = 'var(--mp-primary)'; }}
          >
            {loading ? '注册中...' : '注 册'}
          </button>
        </form>

        <p className="text-center text-sm mt-5" style={{ color: 'var(--mp-text-secondary)' }}>
          已有账号？
          <Link to="/login" className="font-medium hover:underline" style={{ color: 'var(--mp-primary)' }}>
            登录
          </Link>
        </p>
      </div>
    </div>
  );
}
