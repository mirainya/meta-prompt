import { useState, useEffect } from 'react';
import { api, type TemplateItem, type TemplateVersion } from '../../lib/api';
import { useToast } from '../../components/Toast';

const STAGE_LABELS: Record<string, string> = {
  analyzer: '分析师',
  architect: '架构师',
  writer: '撰写师',
  reviewer: '审稿人',
};

const STAGE_ORDER = ['analyzer', 'architect', 'writer', 'reviewer'];

export default function Templates() {
  const [templates, setTemplates] = useState<TemplateItem[]>([]);
  const [editing, setEditing] = useState<number | null>(null);
  const [form, setForm] = useState({ name: '', description: '', prompt: '' });
  const [saving, setSaving] = useState(false);
  const [msg, setMsg] = useState('');
  const [msgType, setMsgType] = useState<'ok' | 'err'>('ok');
  const [versionsId, setVersionsId] = useState<number | null>(null);
  const [versions, setVersions] = useState<TemplateVersion[]>([]);
  const toast = useToast((s) => s.add);

  const load = () => api.adminTemplates().then(setTemplates);
  useEffect(() => { load(); }, []);

  const flash = (text: string, type: 'ok' | 'err' = 'ok') => { setMsg(text); setMsgType(type); };

  const sorted = [...templates].sort((a, b) => {
    const ai = STAGE_ORDER.indexOf(a.stage);
    const bi = STAGE_ORDER.indexOf(b.stage);
    return (ai === -1 ? 99 : ai) - (bi === -1 ? 99 : bi);
  });

  const startEdit = (t: TemplateItem) => {
    setEditing(t.id);
    setForm({ name: t.name, description: t.description, prompt: t.prompt });
    setMsg('');
  };

  const showVersions = async (id: number) => {
    if (versionsId === id) { setVersionsId(null); return; }
    const list = await api.templateVersions(id);
    setVersions(list);
    setVersionsId(id);
  };

  const rollback = async (templateId: number, version: number) => {
    try {
      await api.templateRollback(templateId, version);
      toast('已回滚到 v' + version, 'success');
      setVersionsId(null);
      load();
    } catch {
      toast('回滚失败', 'error');
    }
  };

  const save = async () => {
    if (!editing) return;
    setSaving(true);
    try {
      await api.adminUpdateTemplate(editing, form);
      flash('保存成功');
      setEditing(null);
      load();
    } catch (err: unknown) {
      flash(err instanceof Error ? err.message : '保存失败', 'err');
    } finally {
      setSaving(false);
    }
  };

  const inputStyle = {
    border: '1px solid var(--mp-border-color)',
    background: 'var(--mp-primary-lighter)',
    color: 'var(--mp-text-primary)',
  };

  return (
    <div className="max-w-4xl mx-auto">
      <h2 className="text-base font-semibold mb-4" style={{ color: 'var(--mp-text-primary)' }}>元提示词管理</h2>
      <p className="text-xs mb-4" style={{ color: 'var(--mp-text-secondary)' }}>
        这些是驱动四层 Pipeline 的系统提示词。修改后立即生效，影响所有用户的生成结果。
      </p>

      {msg && (
        <div className="rounded-xl px-4 py-2 mb-4 text-sm" style={{
          background: msgType === 'ok' ? 'rgba(139, 189, 139, 0.1)' : 'rgba(220, 100, 100, 0.1)',
          color: msgType === 'ok' ? 'var(--mp-success)' : '#dc6464',
        }}>
          {msg}
        </div>
      )}

      <div className="space-y-4">
        {sorted.map((t) => {
          const isEditing = editing === t.id;
          const stageLabel = STAGE_LABELS[t.stage] || t.stage;
          const stageIdx = STAGE_ORDER.indexOf(t.stage);

          return (
            <div key={t.id} className="rounded-2xl overflow-hidden transition-all" style={{ background: 'var(--mp-card-bg)', border: '1px solid var(--mp-card-border)', boxShadow: 'var(--mp-card-shadow)' }}>
              <div className="px-5 py-4 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl flex items-center justify-center text-white text-sm font-bold" style={{ background: 'var(--mp-primary)' }}>
                    {stageIdx >= 0 ? stageIdx + 1 : '?'}
                  </div>
                  <div>
                    <div className="text-sm font-semibold" style={{ color: 'var(--mp-text-primary)' }}>
                      {stageLabel}
                      <span className="ml-2 text-xs font-normal" style={{ color: 'var(--mp-text-secondary)' }}>
                        {t.name} · v{t.version}
                      </span>
                    </div>
                    <div className="text-xs mt-0.5" style={{ color: 'var(--mp-text-secondary)' }}>
                      {t.description} · {t.prompt.length} 字符
                    </div>
                  </div>
                </div>
                {!isEditing && (
                  <div className="flex gap-2">
                    <button
                      onClick={() => showVersions(t.id)}
                      className="h-8 px-3 rounded-lg text-xs font-medium transition-all"
                      style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-secondary)' }}
                    >
                      版本
                    </button>
                    <button
                      onClick={() => startEdit(t)}
                      className="h-8 px-4 rounded-lg text-xs font-medium transition-all"
                      style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-primary)' }}
                    >
                      编辑
                    </button>
                  </div>
                )}
              </div>

              {isEditing && (
                <div className="px-5 pb-5 space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>名称</label>
                      <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
                    </div>
                    <div>
                      <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>描述</label>
                      <input className="w-full h-9 px-3 rounded-lg text-sm outline-none" style={inputStyle} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
                    </div>
                  </div>
                  <div>
                    <label className="text-xs mb-1 block" style={{ color: 'var(--mp-text-secondary)' }}>提示词内容</label>
                    <textarea
                      className="w-full px-3 py-2 rounded-lg text-sm outline-none resize-y font-mono leading-relaxed"
                      style={{ ...inputStyle, minHeight: '300px' }}
                      value={form.prompt}
                      onChange={(e) => setForm({ ...form, prompt: e.target.value })}
                    />
                  </div>
                  <div className="flex gap-2">
                    <button onClick={save} disabled={saving} className="h-8 px-5 rounded-lg text-white text-xs font-medium disabled:opacity-50" style={{ background: 'var(--mp-primary)' }}>
                      {saving ? '保存中...' : '保存'}
                    </button>
                    <button onClick={() => setEditing(null)} className="h-8 px-5 rounded-lg text-xs font-medium" style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-regular)' }}>
                      取消
                    </button>
                  </div>
                </div>
              )}

              {versionsId === t.id && !isEditing && (
                <div className="px-5 pb-5">
                  {versions.length === 0 ? (
                    <div className="text-xs py-3" style={{ color: 'var(--mp-text-secondary)' }}>暂无历史版本</div>
                  ) : (
                    <div className="space-y-2 max-h-60 overflow-y-auto">
                      {versions.map((v) => (
                        <div key={v.id} className="flex items-center justify-between px-3 py-2 rounded-lg" style={{ background: 'var(--mp-primary-lighter)' }}>
                          <div>
                            <span className="text-xs font-medium" style={{ color: 'var(--mp-text-primary)' }}>v{v.version}</span>
                            <span className="text-xs ml-2" style={{ color: 'var(--mp-text-secondary)' }}>
                              {new Date(v.created_at).toLocaleString()} · {v.prompt.length} 字符
                            </span>
                          </div>
                          <button
                            onClick={() => rollback(t.id, v.version)}
                            className="text-xs px-2 py-1 rounded"
                            style={{ color: 'var(--mp-primary)' }}
                          >
                            回滚
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
