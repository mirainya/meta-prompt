import { useState, useEffect } from 'react';
import { api, type LLMConfigItem } from '../../lib/api';

const TYPE_OPTIONS = [
  { value: 'openai_compatible', label: 'OpenAI 兼容' },
  { value: 'claude', label: 'Claude' },
  { value: 'gemini', label: 'Gemini' },
];

export default function LLMConfig() {
  const [configs, setConfigs] = useState<LLMConfigItem[]>([]);
  const [editing, setEditing] = useState<string | null>(null);
  const [form, setForm] = useState<Partial<LLMConfigItem>>({});
  const [saving, setSaving] = useState(false);
  const [msg, setMsg] = useState('');
  const [msgType, setMsgType] = useState<'ok' | 'err'>('ok');
  const [testing, setTesting] = useState<string | null>(null);
  const [showAdd, setShowAdd] = useState(false);
  const [addForm, setAddForm] = useState({ provider: '', type: 'openai_compatible', api_key: '', base_url: '', model: '', max_tokens: 4096, enabled: true });

  const load = () => api.adminLLMConfigs().then(setConfigs);
  useEffect(() => { load(); }, []);

  const flash = (text: string, type: 'ok' | 'err' = 'ok') => { setMsg(text); setMsgType(type); };

  const startEdit = (cfg: LLMConfigItem) => {
    setEditing(cfg.provider);
    setForm({ type: cfg.type, api_key: '', base_url: cfg.base_url, model: cfg.model, max_tokens: cfg.max_tokens, enabled: cfg.enabled });
    setMsg('');
    setShowAdd(false);
  };

  const save = async () => {
    if (!editing) return;
    setSaving(true);
    setMsg('');
    try {
      const data: Record<string, unknown> = {};
      if (form.type !== undefined) data.type = form.type;
      if (form.api_key) data.api_key = form.api_key;
      if (form.base_url !== undefined) data.base_url = form.base_url;
      if (form.model !== undefined) data.model = form.model;
      if (form.max_tokens !== undefined) data.max_tokens = form.max_tokens;
      if (form.enabled !== undefined) data.enabled = form.enabled;
      await api.adminUpdateLLMConfig(editing, data as Partial<LLMConfigItem>);
      flash('保存成功');
      setEditing(null);
      load();
    } catch (err: unknown) {
      flash(err instanceof Error ? err.message : '保存失败', 'err');
    } finally {
      setSaving(false);
    }
  };

  const handleAdd = async () => {
    if (!addForm.provider.trim() || !addForm.model.trim()) { flash('名称和模型不能为空', 'err'); return; }
    setSaving(true);
    try {
      await api.adminCreateLLMConfig(addForm);
      flash('添加成功');
      setShowAdd(false);
      setAddForm({ provider: '', type: 'openai_compatible', api_key: '', base_url: '', model: '', max_tokens: 4096, enabled: true });
      load();
    } catch (err: unknown) {
      flash(err instanceof Error ? err.message : '添加失败', 'err');
    } finally {
      setSaving(false);
    }
  };

  const handleTest = async (provider: string) => {
    setTesting(provider);
    flash('');
    try {
      const res = await api.adminTestLLMConfig(provider);
      flash(res.success ? `测试成功：${res.reply}` : `测试失败：${res.error}`, res.success ? 'ok' : 'err');
    } catch (err: unknown) {
      flash(err instanceof Error ? err.message : '测试失败', 'err');
    } finally {
      setTesting(null);
    }
  };

  const handleDelete = async (provider: string) => {
    if (!confirm(`确定删除 ${provider}？`)) return;
    try {
      await api.adminDeleteLLMConfig(provider);
      flash('已删除');
      load();
    } catch (err: unknown) {
      flash(err instanceof Error ? err.message : '删除失败', 'err');
    }
  };

  const inputStyle = {
    border: '1px solid var(--mp-border-color)',
    background: 'var(--mp-primary-lighter)',
    color: 'var(--mp-text-primary)',
  };

  return (
    <div className="max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-base font-semibold" style={{ color: 'var(--mp-text-primary)' }}>模型配置</h2>
        <button
          onClick={() => { setShowAdd(!showAdd); setEditing(null); setMsg(''); }}
          className="h-8 px-4 rounded-lg text-white text-xs font-medium transition-all"
          style={{ background: 'var(--mp-primary)' }}
        >
          + 添加模型
        </button>
      </div>

      {msg && (
        <div className="rounded-xl px-4 py-2 mb-4 text-sm" style={{
          background: msgType === 'ok' ? 'rgba(139, 189, 139, 0.1)' : 'rgba(220, 100, 100, 0.1)',
          color: msgType === 'ok' ? 'var(--mp-success)' : '#dc6464',
        }}>
          {msg}
        </div>
      )}

      {showAdd && (
        <div className="rounded-2xl p-5 mb-4" style={{ background: 'var(--mp-card-bg)', border: '1px solid var(--mp-card-border)', boxShadow: 'var(--mp-card-shadow)' }}>
          <div className="text-sm font-semibold mb-3" style={{ color: 'var(--mp-text-primary)' }}>添加新模型</div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>名称（唯一标识）</label>
              <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} value={addForm.provider} onChange={(e) => setAddForm({ ...addForm, provider: e.target.value })} placeholder="如 deepseek" />
            </div>
            <div>
              <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>协议类型</label>
              <select className="w-full h-9 px-3 rounded-lg text-sm outline-none cursor-pointer" style={inputStyle} value={addForm.type} onChange={(e) => setAddForm({ ...addForm, type: e.target.value })}>
                {TYPE_OPTIONS.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
              </select>
            </div>
            <div>
              <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>API Key</label>
              <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} type="password" value={addForm.api_key} onChange={(e) => setAddForm({ ...addForm, api_key: e.target.value })} />
            </div>
            <div>
              <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>Base URL</label>
              <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} value={addForm.base_url} onChange={(e) => setAddForm({ ...addForm, base_url: e.target.value })} placeholder="留空使用默认" />
            </div>
            <div>
              <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>模型名</label>
              <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} value={addForm.model} onChange={(e) => setAddForm({ ...addForm, model: e.target.value })} placeholder="如 deepseek-chat" />
            </div>
            <div>
              <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>Max Tokens</label>
              <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} type="number" value={addForm.max_tokens} onChange={(e) => setAddForm({ ...addForm, max_tokens: Number(e.target.value) })} />
            </div>
          </div>
          <div className="flex items-center gap-3 mt-3">
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" checked={addForm.enabled} onChange={(e) => setAddForm({ ...addForm, enabled: e.target.checked })} className="accent-[var(--mp-primary)]" />
              <span className="text-sm" style={{ color: 'var(--mp-text-regular)' }}>启用</span>
            </label>
            <div className="flex-1" />
            <button onClick={handleAdd} disabled={saving} className="h-8 px-5 rounded-lg text-white text-xs font-medium disabled:opacity-50" style={{ background: 'var(--mp-primary)' }}>
              {saving ? '添加中...' : '添加'}
            </button>
            <button onClick={() => setShowAdd(false)} className="h-8 px-5 rounded-lg text-xs font-medium" style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-regular)' }}>
              取消
            </button>
          </div>
        </div>
      )}

      <div className="space-y-4">
        {configs.map((cfg) => {
          const isEditing = editing === cfg.provider;
          const typeLabel = TYPE_OPTIONS.find((t) => t.value === cfg.type)?.label || cfg.type;

          return (
            <div key={cfg.provider} className="rounded-2xl overflow-hidden transition-all" style={{ background: 'var(--mp-card-bg)', border: '1px solid var(--mp-card-border)', boxShadow: 'var(--mp-card-shadow)' }}>
              <div className="px-5 py-4 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl flex items-center justify-center text-white text-sm font-bold" style={{ background: cfg.enabled ? 'var(--mp-primary)' : 'var(--mp-text-secondary)' }}>
                    {cfg.provider.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <div className="text-sm font-semibold" style={{ color: 'var(--mp-text-primary)' }}>{cfg.provider}</div>
                    <div className="text-xs" style={{ color: 'var(--mp-text-secondary)' }}>
                      {typeLabel} · {cfg.model} · {cfg.enabled ? '已启用' : '已禁用'}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => handleTest(cfg.provider)}
                    disabled={testing === cfg.provider}
                    className="h-8 px-3 rounded-lg text-xs font-medium transition-all disabled:opacity-50"
                    style={{ border: '1px solid var(--mp-border-color)', color: 'var(--mp-text-regular)' }}
                  >
                    {testing === cfg.provider ? '测试中...' : '测试'}
                  </button>
                  <button
                    onClick={() => startEdit(cfg)}
                    className="h-8 px-3 rounded-lg text-xs font-medium transition-all"
                    style={{ border: '1px solid var(--mp-border-color)', color: 'var(--mp-text-regular)' }}
                  >
                    编辑
                  </button>
                  <button
                    onClick={() => handleDelete(cfg.provider)}
                    className="h-8 px-3 rounded-lg text-xs font-medium transition-all"
                    style={{ border: '1px solid #dc6464', color: '#dc6464' }}
                  >
                    删除
                  </button>
                </div>
              </div>

              {isEditing && (
                <div className="px-5 pb-5 space-y-3 border-t" style={{ borderColor: 'var(--mp-card-border)' }}>
                  <div className="grid grid-cols-2 gap-3 pt-3">
                    <div>
                      <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>协议类型</label>
                      <select className="w-full h-9 px-3 rounded-lg text-sm outline-none cursor-pointer" style={inputStyle} value={form.type || cfg.type} onChange={(e) => setForm({ ...form, type: e.target.value })}>
                        {TYPE_OPTIONS.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
                      </select>
                    </div>
                    <div>
                      <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>API Key</label>
                      <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} type="password" placeholder="留空不修改" value={form.api_key || ''} onChange={(e) => setForm({ ...form, api_key: e.target.value })} />
                    </div>
                    <div>
                      <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>Base URL</label>
                      <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} value={form.base_url || ''} onChange={(e) => setForm({ ...form, base_url: e.target.value })} />
                    </div>
                    <div>
                      <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>模型名</label>
                      <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} value={form.model || ''} onChange={(e) => setForm({ ...form, model: e.target.value })} />
                    </div>
                    <div>
                      <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>Max Tokens</label>
                      <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} type="number" value={form.max_tokens || 4096} onChange={(e) => setForm({ ...form, max_tokens: Number(e.target.value) })} />
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <label className="flex items-center gap-2 cursor-pointer">
                      <input type="checkbox" checked={form.enabled ?? true} onChange={(e) => setForm({ ...form, enabled: e.target.checked })} className="accent-[var(--mp-primary)]" />
                      <span className="text-sm" style={{ color: 'var(--mp-text-regular)' }}>启用</span>
                    </label>
                    <div className="flex-1" />
                    <button onClick={save} disabled={saving} className="h-8 px-5 rounded-lg text-white text-xs font-medium disabled:opacity-50" style={{ background: 'var(--mp-primary)' }}>
                      {saving ? '保存中...' : '保存'}
                    </button>
                    <button onClick={() => setEditing(null)} className="h-8 px-5 rounded-lg text-xs font-medium" style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-regular)' }}>
                      取消
                    </button>
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
