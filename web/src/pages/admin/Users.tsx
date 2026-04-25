import { useState, useEffect } from 'react';
import { api, type AdminUser } from '../../lib/api';

export default function Users() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [editCredits, setEditCredits] = useState<{ id: number; value: string } | null>(null);
  const [resetPwd, setResetPwd] = useState<{ id: number; value: string } | null>(null);
  const [msg, setMsg] = useState('');

  const load = () => {
    setLoading(true);
    api.adminUsers(100, 0).then((res) => {
      setUsers(res.users);
      setTotal(res.total);
    }).finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, []);

  const handleSetCredits = async (id: number, credits: number) => {
    try {
      await api.adminSetCredits(id, credits);
      setEditCredits(null);
      setMsg('积分已更新');
      load();
    } catch (err: unknown) {
      setMsg(err instanceof Error ? err.message : '操作失败');
    }
  };

  const handleToggle = async (id: number) => {
    try {
      await api.adminToggleUser(id);
      setMsg('状态已更新');
      load();
    } catch (err: unknown) {
      setMsg(err instanceof Error ? err.message : '操作失败');
    }
  };

  const handleResetPwd = async (id: number, password: string) => {
    try {
      await api.adminResetPassword(id, password);
      setResetPwd(null);
      setMsg('密码已重置');
    } catch (err: unknown) {
      setMsg(err instanceof Error ? err.message : '操作失败');
    }
  };

  if (loading) {
    return <div className="text-sm" style={{ color: 'var(--mp-text-secondary)' }}>加载中...</div>;
  }

  return (
    <div className="max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <h2 className="text-base font-semibold" style={{ color: 'var(--mp-text-primary)' }}>用户管理</h2>
          <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-primary)' }}>
            {total} 人
          </span>
        </div>
      </div>

      {msg && (
        <div
          className="rounded-xl px-4 py-2 mb-4 text-sm cursor-pointer"
          style={{ background: 'rgba(139, 189, 139, 0.1)', color: 'var(--mp-success)' }}
          onClick={() => setMsg('')}
        >
          {msg}
        </div>
      )}

      <div
        className="rounded-2xl overflow-hidden"
        style={{
          background: 'var(--mp-card-bg)',
          border: '1px solid var(--mp-card-border)',
          boxShadow: 'var(--mp-card-shadow)',
        }}
      >
        <table className="w-full text-sm">
          <thead>
            <tr style={{ borderBottom: '1px solid var(--mp-border-color)' }}>
              {['ID', '用户名', '角色', '积分', '状态', '注册时间', '操作'].map((h) => (
                <th key={h} className="px-4 py-3 text-left font-medium" style={{ color: 'var(--mp-text-secondary)' }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id} style={{ borderBottom: '1px solid var(--mp-card-border)' }}>
                <td className="px-4 py-3" style={{ color: 'var(--mp-text-secondary)' }}>{u.id}</td>
                <td className="px-4 py-3 font-medium" style={{ color: 'var(--mp-text-primary)' }}>{u.username}</td>
                <td className="px-4 py-3">
                  <span
                    className="text-xs px-2 py-0.5 rounded-full"
                    style={{
                      background: u.role === 'admin' ? 'rgba(154, 138, 189, 0.12)' : 'var(--mp-primary-lighter)',
                      color: u.role === 'admin' ? 'var(--mp-accent)' : 'var(--mp-text-secondary)',
                    }}
                  >
                    {u.role || 'user'}
                  </span>
                </td>
                <td className="px-4 py-3">
                  {editCredits?.id === u.id ? (
                    <div className="flex items-center gap-1">
                      <input
                        type="number"
                        value={editCredits.value}
                        onChange={(e) => setEditCredits({ id: u.id, value: e.target.value })}
                        className="w-20 h-7 px-2 rounded text-xs outline-none"
                        style={{ border: '1px solid var(--mp-border-color)', background: 'var(--mp-primary-lighter)' }}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') handleSetCredits(u.id, Number(editCredits.value));
                          if (e.key === 'Escape') setEditCredits(null);
                        }}
                        autoFocus
                      />
                      <button
                        onClick={() => handleSetCredits(u.id, Number(editCredits.value))}
                        className="text-xs px-2 py-1 rounded"
                        style={{ color: 'var(--mp-primary)' }}
                      >
                        ✓
                      </button>
                    </div>
                  ) : (
                    <span
                      className="cursor-pointer hover:underline"
                      style={{ color: 'var(--mp-primary)' }}
                      onClick={() => setEditCredits({ id: u.id, value: String(u.credits) })}
                      title="点击修改"
                    >
                      {u.credits}
                    </span>
                  )}
                </td>
                <td className="px-4 py-3">
                  <span
                    className="text-xs px-2 py-0.5 rounded-full"
                    style={{
                      background: u.disabled ? 'rgba(212, 122, 122, 0.1)' : 'rgba(139, 189, 139, 0.1)',
                      color: u.disabled ? 'var(--mp-danger)' : 'var(--mp-success)',
                    }}
                  >
                    {u.disabled ? '已禁用' : '正常'}
                  </span>
                </td>
                <td className="px-4 py-3 text-xs" style={{ color: 'var(--mp-text-secondary)' }}>
                  {new Date(u.created_at).toLocaleDateString('zh-CN')}
                </td>
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => handleToggle(u.id)}
                      className="text-xs px-2 py-1 rounded transition-all"
                      style={{
                        background: u.disabled ? 'rgba(139, 189, 139, 0.1)' : 'rgba(212, 122, 122, 0.1)',
                        color: u.disabled ? 'var(--mp-success)' : 'var(--mp-danger)',
                      }}
                    >
                      {u.disabled ? '启用' : '禁用'}
                    </button>
                    {resetPwd?.id === u.id ? (
                      <div className="flex items-center gap-1">
                        <input
                          type="text"
                          value={resetPwd.value}
                          onChange={(e) => setResetPwd({ id: u.id, value: e.target.value })}
                          placeholder="新密码"
                          className="w-24 h-7 px-2 rounded text-xs outline-none"
                          style={{ border: '1px solid var(--mp-border-color)', background: 'var(--mp-primary-lighter)' }}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter' && resetPwd.value.length >= 6) handleResetPwd(u.id, resetPwd.value);
                            if (e.key === 'Escape') setResetPwd(null);
                          }}
                          autoFocus
                        />
                        <button
                          onClick={() => resetPwd.value.length >= 6 && handleResetPwd(u.id, resetPwd.value)}
                          className="text-xs px-2 py-1 rounded"
                          style={{ color: 'var(--mp-primary)' }}
                        >
                          ✓
                        </button>
                      </div>
                    ) : (
                      <button
                        onClick={() => setResetPwd({ id: u.id, value: '' })}
                        className="text-xs px-2 py-1 rounded transition-all"
                        style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-regular)' }}
                      >
                        重置密码
                      </button>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
