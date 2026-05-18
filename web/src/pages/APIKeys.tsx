import { useState, useEffect } from 'react';
import { api, type APIKeyItem } from '../lib/api';
import { useToast } from '../components/Toast';

export default function APIKeys() {
  const [keys, setKeys] = useState<APIKeyItem[]>([]);
  const [showDialog, setShowDialog] = useState(false);
  const [name, setName] = useState('');
  const [newKey, setNewKey] = useState('');
  const [loading, setLoading] = useState(false);
  const toast = useToast((s) => s.add);

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    toast('已复制到剪贴板', 'success');
  };

  const load = () => api.listAPIKeys().then(setKeys);
  useEffect(() => { load(); }, []);

  const handleCreate = async () => {
    if (!name.trim()) return;
    setLoading(true);
    try {
      const res = await api.createAPIKey(name.trim());
      setNewKey(res.key);
      setName('');
      load();
    } finally {
      setLoading(false);
    }
  };

  const closeDialog = () => { setShowDialog(false); setNewKey(''); setName(''); };

  const handleRevoke = async (id: number) => {
    if (!confirm('确定撤销此 Key？撤销后无法恢复。')) return;
    await api.revokeAPIKey(id);
    load();
  };

  return (
    <div className="max-w-5xl mx-auto space-y-5">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold" style={{ color: 'var(--mp-text-primary)' }}>API Key 管理</h2>
        <button
          onClick={() => setShowDialog(true)}
          className="h-9 px-4 rounded-xl text-sm font-medium text-white"
          style={{ background: 'var(--mp-primary)' }}
        >
          + 创建 Key
        </button>
      </div>

      {/* 列表 */}
      <div
        className="rounded-2xl overflow-hidden"
        style={{ background: 'var(--mp-card-bg)', border: '1px solid var(--mp-card-border)', boxShadow: 'var(--mp-card-shadow)' }}
      >
        {keys.length === 0 ? (
          <div className="p-8 text-center text-sm" style={{ color: 'var(--mp-text-secondary)' }}>
            暂无 API Key，点击右上角创建
          </div>
        ) : (
          <table className="w-full text-sm table-fixed">
            <thead>
              <tr style={{ borderBottom: '1px solid var(--mp-border-color)' }}>
                <th className="text-left px-4 py-3 font-medium w-[12%]" style={{ color: 'var(--mp-text-secondary)' }}>名称</th>
                <th className="text-left px-4 py-3 font-medium w-[46%]" style={{ color: 'var(--mp-text-secondary)' }}>Key</th>
                <th className="text-left px-4 py-3 font-medium w-[12%]" style={{ color: 'var(--mp-text-secondary)' }}>状态</th>
                <th className="text-left px-4 py-3 font-medium w-[18%]" style={{ color: 'var(--mp-text-secondary)' }}>最后使用</th>
                <th className="text-right px-4 py-3 font-medium w-[12%]" style={{ color: 'var(--mp-text-secondary)' }}>操作</th>
              </tr>
            </thead>
            <tbody>
              {keys.map((k) => (
                <tr key={k.id} style={{ borderBottom: '1px solid var(--mp-border-color)' }}>
                  <td className="px-4 py-3" style={{ color: 'var(--mp-text-primary)' }}>{k.name}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1">
                      <code className="text-xs px-2 py-0.5 rounded truncate block max-w-[360px]" style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-regular)' }} title={k.raw_key || `${k.prefix}...`}>
                        {k.raw_key || `${k.prefix}...`}
                      </code>
                      {k.raw_key && (
                        <button
                          onClick={() => copyKey(k.raw_key)}
                          className="text-xs px-1.5 py-0.5 rounded hover:opacity-80"
                          style={{ color: 'var(--mp-primary)' }}
                        >
                          复制
                        </button>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className="text-xs px-2 py-0.5 rounded-full"
                      style={{
                        background: k.is_active ? 'rgba(76, 175, 80, 0.1)' : 'rgba(220, 100, 100, 0.1)',
                        color: k.is_active ? 'var(--mp-success)' : 'var(--mp-danger)',
                      }}
                    >
                      {k.is_active ? '活跃' : '已撤销'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs" style={{ color: 'var(--mp-text-secondary)' }}>
                    {k.last_used_at ? new Date(k.last_used_at).toLocaleString() : '从未使用'}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {k.is_active && (
                      <button
                        onClick={() => handleRevoke(k.id)}
                        className="text-xs px-2.5 py-1 rounded-lg"
                        style={{ color: 'var(--mp-danger)', background: 'rgba(220, 100, 100, 0.08)' }}
                      >
                        撤销
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* 创建对话框 */}
      {showDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center" onClick={closeDialog}>
          <div className="absolute inset-0 bg-black/40" />
          <div
            className="relative w-full max-w-md rounded-2xl p-6 space-y-4"
            style={{ background: 'var(--mp-card-bg)', boxShadow: '0 20px 60px rgba(0,0,0,0.3)' }}
            onClick={(e) => e.stopPropagation()}
          >
            {!newKey ? (
              <>
                <h3 className="text-base font-semibold" style={{ color: 'var(--mp-text-primary)' }}>创建 API Key</h3>
                <input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Key 名称，如：我的应用"
                  className="w-full h-10 px-4 rounded-xl text-sm outline-none"
                  style={{ background: 'var(--mp-input-bg)', border: '1px solid var(--mp-border-color)', color: 'var(--mp-text-primary)' }}
                  onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
                  autoFocus
                />
                <div className="flex justify-end gap-2">
                  <button onClick={closeDialog} className="h-9 px-4 rounded-xl text-sm" style={{ color: 'var(--mp-text-secondary)' }}>取消</button>
                  <button
                    onClick={handleCreate}
                    disabled={loading || !name.trim()}
                    className="h-9 px-5 rounded-xl text-sm font-medium text-white disabled:opacity-50"
                    style={{ background: 'var(--mp-primary)' }}
                  >
                    {loading ? '创建中...' : '创建'}
                  </button>
                </div>
              </>
            ) : (
              <>
                <h3 className="text-base font-semibold" style={{ color: 'var(--mp-text-primary)' }}>Key 已创建</h3>
                <div className="flex items-center gap-2">
                  <code className="flex-1 text-xs break-all p-3 rounded-lg" style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-primary)' }}>
                    {newKey}
                  </code>
                  <button
                    onClick={() => copyKey(newKey)}
                    className="shrink-0 px-3 py-1.5 rounded-lg text-xs font-medium text-white"
                    style={{ background: 'var(--mp-primary)' }}
                  >
                    复制
                  </button>
                </div>
                <div className="flex justify-end">
                  <button onClick={closeDialog} className="h-9 px-5 rounded-xl text-sm font-medium" style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-primary)' }}>关闭</button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
