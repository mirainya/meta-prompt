import { useState, useEffect } from 'react';
import { api, type ChannelSource, type ChannelModel } from '../../lib/api';
import { useToast } from '../../components/Toast';

export default function LLMConfig() {
  const [sources, setSources] = useState<ChannelSource[]>([]);
  const [models, setModels] = useState<ChannelModel[]>([]);
  const [tab, setTab] = useState<'sources' | 'models'>('sources');
  const [showAdd, setShowAdd] = useState(false);
  const [syncing, setSyncing] = useState<number | null>(null);
  const [addForm, setAddForm] = useState({ name: '', base_url: '', api_key: '', proxy_url: '' });
  const [saving, setSaving] = useState(false);
  const toast = useToast((s) => s.add);

  const loadSources = () => api.adminChannelSources().then(setSources);
  const loadModels = () => api.adminChannelModels().then(setModels);
  useEffect(() => { loadSources(); loadModels(); }, []);

  const handleAdd = async () => {
    if (!addForm.name.trim() || !addForm.base_url.trim()) { toast('名称和地址不能为空', 'error'); return; }
    setSaving(true);
    try {
      await api.adminCreateSource(addForm);
      toast('添加成功', 'success');
      setShowAdd(false);
      setAddForm({ name: '', base_url: '', api_key: '', proxy_url: '' });
      loadSources();
    } catch (err: unknown) {
      toast(err instanceof Error ? err.message : '添加失败', 'error');
    } finally { setSaving(false); }
  };

  const handleSync = async (id: number) => {
    setSyncing(id);
    try {
      const res = await api.adminSyncSource(id);
      toast(`同步成功，${res.synced} 个 chat 模型（共 ${res.total} 个）`, 'success');
      loadModels();
    } catch (err: unknown) {
      toast(err instanceof Error ? err.message : '同步失败', 'error');
    } finally { setSyncing(null); }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('确定删除此渠道源？关联的模型也会被删除。')) return;
    await api.adminDeleteSource(id);
    loadSources();
    loadModels();
  };

  const handleToggleModel = async (m: ChannelModel) => {
    await api.adminUpdateModel(m.id, { enabled: !m.enabled });
    loadModels();
  };

  const handleBillingChange = async (m: ChannelModel, billing_type: string) => {
    await api.adminUpdateModel(m.id, { billing_type });
    loadModels();
  };

  const handleCreditsChange = async (m: ChannelModel, credits: number) => {
    await api.adminUpdateModel(m.id, { credits_per_call: credits });
    loadModels();
  };

  const handleTokenPriceChange = async (m: ChannelModel, field: 'input_token_price' | 'output_token_price', price: number) => {
    await api.adminUpdateModel(m.id, { [field]: price });
    loadModels();
  };

  const handleSetDefault = async (m: ChannelModel) => {
    await api.adminUpdateModel(m.id, { is_default: true });
    loadModels();
  };

  const inputStyle = { background: 'var(--mp-primary-lighter)', border: '1px solid var(--mp-border-color)', color: 'var(--mp-text-primary)' };

  return (
    <div className="max-w-5xl mx-auto space-y-5">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h2 className="text-lg font-semibold" style={{ color: 'var(--mp-text-primary)' }}>渠道管理</h2>
          <div className="flex rounded-lg overflow-hidden" style={{ border: '1px solid var(--mp-border-color)' }}>
            <button onClick={() => setTab('sources')} className="px-3 py-1.5 text-xs font-medium" style={{ background: tab === 'sources' ? 'var(--mp-primary)' : 'transparent', color: tab === 'sources' ? '#fff' : 'var(--mp-text-secondary)' }}>渠道源</button>
            <button onClick={() => setTab('models')} className="px-3 py-1.5 text-xs font-medium" style={{ background: tab === 'models' ? 'var(--mp-primary)' : 'transparent', color: tab === 'models' ? '#fff' : 'var(--mp-text-secondary)' }}>模型列表</button>
          </div>
        </div>
        {tab === 'sources' && (
          <button onClick={() => setShowAdd(true)} className="h-9 px-4 rounded-xl text-sm font-medium text-white" style={{ background: 'var(--mp-primary)' }}>+ 添加渠道</button>
        )}
      </div>

      {tab === 'sources' && (
        <div className="space-y-3">
          {sources.length === 0 && (
            <div className="rounded-2xl p-8 text-center text-sm" style={{ background: 'var(--mp-card-bg)', border: '1px solid var(--mp-card-border)', color: 'var(--mp-text-secondary)' }}>暂无渠道源，点击右上角添加</div>
          )}
          {sources.map((s) => (
            <div key={s.id} className="rounded-xl p-4" style={{ background: 'var(--mp-card-bg)', border: '1px solid var(--mp-card-border)' }}>
              <div className="flex items-center justify-between">
                <div>
                  <span className="text-sm font-medium" style={{ color: 'var(--mp-text-primary)' }}>{s.name}</span>
                </div>
                <div className="flex items-center gap-2">
                  <button onClick={() => handleSync(s.id)} disabled={syncing === s.id} className="h-7 px-3 rounded-lg text-xs font-medium text-white disabled:opacity-50" style={{ background: '#10b981' }}>
                    {syncing === s.id ? '同步中...' : '同步模型'}
                  </button>
                  <button onClick={() => handleDelete(s.id)} className="h-7 px-3 rounded-lg text-xs font-medium" style={{ background: '#fee2e2', color: '#dc2626' }}>删除</button>
                </div>
              </div>
              <div className="mt-2 text-xs" style={{ color: 'var(--mp-text-secondary)' }}>{s.base_url}</div>
            </div>
          ))}
        </div>
      )}

      {tab === 'models' && (
        <div>
          {models.length === 0 ? (
            <div className="rounded-2xl p-8 text-center text-sm" style={{ background: 'var(--mp-card-bg)', border: '1px solid var(--mp-card-border)', color: 'var(--mp-text-secondary)' }}>暂无模型，请先添加渠道源并点击「同步模型」</div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {models.map((m) => (
                <div key={m.id} className="rounded-2xl p-4 space-y-3 relative" style={{ background: 'var(--mp-card-bg)', border: m.is_default ? '2px solid var(--mp-primary)' : '1px solid var(--mp-card-border)' }}>
                  {m.is_default && (
                    <span className="absolute top-2 right-2 text-xs px-2 py-0.5 rounded-full font-medium" style={{ background: 'var(--mp-primary)', color: '#fff' }}>默认</span>
                  )}
                  <div>
                    <div className="text-sm font-medium truncate" style={{ color: 'var(--mp-text-primary)' }}>{m.code}</div>
                    <div className="text-xs mt-1" style={{ color: 'var(--mp-text-secondary)' }}>{m.source_name}</div>
                  </div>

                  {/* 计费模式 */}
                  <div>
                    <div className="text-xs mb-1.5" style={{ color: 'var(--mp-text-secondary)' }}>计费模式</div>
                    <div className="flex rounded-lg overflow-hidden" style={{ border: '1px solid var(--mp-border-color)' }}>
                      <button onClick={() => handleBillingChange(m, 'per_call')} className="flex-1 px-2 py-1 text-xs" style={{ background: m.billing_type === 'per_call' ? 'var(--mp-primary)' : 'transparent', color: m.billing_type === 'per_call' ? '#fff' : 'var(--mp-text-secondary)' }}>按次</button>
                      <button onClick={() => handleBillingChange(m, 'per_token')} className="flex-1 px-2 py-1 text-xs" style={{ background: m.billing_type === 'per_token' ? 'var(--mp-primary)' : 'transparent', color: m.billing_type === 'per_token' ? '#fff' : 'var(--mp-text-secondary)' }}>按Token</button>
                    </div>
                  </div>

                  {/* 价格 */}
                  <div>
                    {m.billing_type === 'per_token' ? (
                      <div className="space-y-2">
                        <div>
                          <div className="text-xs mb-1" style={{ color: 'var(--mp-text-secondary)' }}>输入积分/千Token</div>
                          <input type="number" className="w-full h-7 px-2 rounded-lg text-xs outline-none" style={inputStyle} defaultValue={m.input_token_price} key={`${m.id}-in`} onBlur={(e) => { const v = Number(e.target.value); if (v >= 0) handleTokenPriceChange(m, 'input_token_price', v); }} />
                        </div>
                        <div>
                          <div className="text-xs mb-1" style={{ color: 'var(--mp-text-secondary)' }}>输出积分/千Token</div>
                          <input type="number" className="w-full h-7 px-2 rounded-lg text-xs outline-none" style={inputStyle} defaultValue={m.output_token_price} key={`${m.id}-out`} onBlur={(e) => { const v = Number(e.target.value); if (v >= 0) handleTokenPriceChange(m, 'output_token_price', v); }} />
                        </div>
                      </div>
                    ) : (
                      <div>
                        <div className="text-xs mb-1.5" style={{ color: 'var(--mp-text-secondary)' }}>积分/次</div>
                        <input type="number" className="w-full h-7 px-2 rounded-lg text-xs outline-none" style={inputStyle} defaultValue={m.credits_per_call} key={`${m.id}-call`} onBlur={(e) => { const v = Number(e.target.value); if (v >= 0) handleCreditsChange(m, v); }} />
                      </div>
                    )}
                  </div>

                  {/* 操作 */}
                  <div className="flex items-center justify-between pt-1">
                    <button
                      onClick={() => handleToggleModel(m)}
                      className="text-xs px-2.5 py-1 rounded-lg"
                      style={{ background: m.enabled ? '#dcfce7' : '#f3f4f6', color: m.enabled ? '#166534' : '#6b7280' }}
                    >
                      {m.enabled ? '已启用' : '已禁用'}
                    </button>
                    {!m.is_default && (
                      <button
                        onClick={() => handleSetDefault(m)}
                        className="text-xs px-2.5 py-1 rounded-lg"
                        style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-primary)' }}
                      >
                        设为默认
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* 添加渠道弹窗 */}
      {showAdd && (
        <div className="fixed inset-0 flex items-center justify-center z-50">
          <div className="absolute inset-0 bg-black/30" onClick={() => setShowAdd(false)} />
          <div className="relative w-[420px] rounded-2xl p-6 space-y-4" style={{ background: 'var(--mp-card-bg)', border: '1px solid var(--mp-card-border)' }}>
            <h3 className="text-base font-semibold" style={{ color: 'var(--mp-text-primary)' }}>添加渠道源</h3>
            <div>
              <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>名称</label>
              <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} value={addForm.name} onChange={(e) => setAddForm({ ...addForm, name: e.target.value })} placeholder="例如：Prism 主渠道" />
            </div>
            <div>
              <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>Base URL</label>
              <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} value={addForm.base_url} onChange={(e) => setAddForm({ ...addForm, base_url: e.target.value })} placeholder="https://api.example.com" />
            </div>
            <div>
              <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>API Key</label>
              <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} type="password" value={addForm.api_key} onChange={(e) => setAddForm({ ...addForm, api_key: e.target.value })} placeholder="sk-..." />
            </div>
            <div>
              <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>代理地址（可选）</label>
              <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} value={addForm.proxy_url} onChange={(e) => setAddForm({ ...addForm, proxy_url: e.target.value })} placeholder="http://proxy:port" />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button onClick={() => setShowAdd(false)} className="h-9 px-4 rounded-xl text-sm" style={{ color: 'var(--mp-text-secondary)' }}>取消</button>
              <button onClick={handleAdd} disabled={saving} className="h-9 px-5 rounded-xl text-sm font-medium text-white disabled:opacity-50" style={{ background: 'var(--mp-primary)' }}>
                {saving ? '添加中...' : '添加'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
