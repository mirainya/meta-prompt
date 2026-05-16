import { useState, useEffect, useRef, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { api, type HistoryItem, type ReviewedPrompt, type ProviderItem } from '../lib/api';
import { useAuth } from '../lib/store';
import { useToast } from '../components/Toast';

function ModelSelect({ providers, value, onChange, disabled }: {
  providers: ProviderItem[];
  value: string;
  onChange: (v: string) => void;
  disabled: boolean;
}) {
  const [open, setOpen] = useState(false);
  const btnRef = useRef<HTMLButtonElement>(null);
  const [pos, setPos] = useState({ top: 0, left: 0, width: 0 });

  useEffect(() => {
    if (!open) return;
    const onClick = (e: MouseEvent) => {
      if (btnRef.current?.contains(e.target as Node)) return;
      setOpen(false);
    };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, [open]);

  const toggle = () => {
    if (disabled) return;
    if (!open && btnRef.current) {
      const r = btnRef.current.getBoundingClientRect();
      setPos({ top: r.bottom + 4, left: r.left, width: Math.max(r.width, 220) });
    }
    setOpen(!open);
  };

  const current = providers.find((p) => p.provider === value);

  return (
    <>
      <button
        ref={btnRef}
        onClick={toggle}
        disabled={disabled}
        className="h-9 px-3 rounded-lg text-sm flex items-center gap-2 transition-all"
        style={{
          background: 'var(--mp-primary-lighter)',
          border: `1px solid ${open ? 'var(--mp-primary-light)' : 'var(--mp-border-color)'}`,
          color: 'var(--mp-text-regular)',
          cursor: disabled ? 'not-allowed' : 'pointer',
          opacity: disabled ? 0.6 : 1,
        }}
      >
        <span>{current ? `${current.provider} (${current.model})` : '选择模型'}</span>
        <svg className="w-3.5 h-3.5 shrink-0 transition-transform" style={{ transform: open ? 'rotate(180deg)' : '' }} fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
        </svg>
      </button>
      {open && createPortal(
        <div
          className="rounded-xl py-1 text-sm shadow-lg"
          style={{
            position: 'fixed',
            top: pos.top,
            left: pos.left,
            minWidth: pos.width,
            zIndex: 9999,
            background: 'var(--mp-card-bg)',
            border: '1px solid var(--mp-border-color)',
            backdropFilter: 'blur(12px)',
          }}
        >
          {providers.map((p) => (
            <button
              key={p.provider}
              onClick={() => { onChange(p.provider); setOpen(false); }}
              className="w-full px-3 py-2 text-left flex items-center gap-2 transition-colors"
              style={{
                color: p.provider === value ? 'var(--mp-primary)' : 'var(--mp-text-regular)',
                background: p.provider === value ? 'var(--mp-primary-lighter)' : 'transparent',
              }}
              onMouseEnter={(e) => { if (p.provider !== value) e.currentTarget.style.background = 'var(--mp-primary-lighter)'; }}
              onMouseLeave={(e) => { if (p.provider !== value) e.currentTarget.style.background = 'transparent'; }}
            >
              <span className="font-medium">{p.provider}</span>
              <span style={{ color: 'var(--mp-text-secondary)', fontSize: '12px' }}>{p.model}</span>
            </button>
          ))}
        </div>,
        document.body,
      )}
    </>
  );
}

const STEPS = ['分析需求', '设计架构', '编写提示词', '审核优化'];

export default function Generate() {
  const [input, setInput] = useState('');
  const [providers, setProviders] = useState<ProviderItem[]>([]);
  const [provider, setProvider] = useState('');
  const toast = useToast((s) => s.add);

  useEffect(() => {
    api.providers().then((list) => {
      setProviders(list);
      if (list.length > 0) setProvider(list[0].provider);
    });
  }, []);

  const [loading, setLoading] = useState(false);
  const [currentStep, setCurrentStep] = useState(-1);
  const [result, setResult] = useState<HistoryItem | null>(null);
  const [error, setError] = useState('');
  const [expandedIdx, setExpandedIdx] = useState<number | null>(0);
  const [copiedIdx, setCopiedIdx] = useState<number | null>(null);
  const [copiedAll, setCopiedAll] = useState(false);
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const pollingIdRef = useRef<number | null>(null);

  const credits = useAuth((s) => s.credits);
  const setCredits = useAuth((s) => s.setCredits);

  const stopPolling = useCallback(() => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current);
      pollingRef.current = null;
    }
    pollingIdRef.current = null;
  }, []);

  const onComplete = useCallback((history: HistoryItem) => {
    stopPolling();
    setResult(history);
    setCurrentStep(4);
    setLoading(false);
    api.me().then((me) => setCredits(me.credits)).catch(() => {});
  }, [stopPolling, setCredits]);

  const onFailed = useCallback((history: HistoryItem) => {
    stopPolling();
    setError(history.error_msg || '推演失败');
    setLoading(false);
  }, [stopPolling]);

  const startPolling = useCallback((historyId: number) => {
    pollingIdRef.current = historyId;
    setLoading(true);
    setError('');
    setResult(null);
    setExpandedIdx(null);

    const poll = async () => {
      try {
        const h = await api.history(historyId);
        setCurrentStep(h.current_step);
        if (h.status === 'done') {
          onComplete(h);
        } else if (h.status === 'failed' || h.status === 'cancelled') {
          onFailed(h);
        }
      } catch {
        // 网络错误，继续轮询
      }
    };

    poll(); // 立即查一次
    pollingRef.current = setInterval(poll, 2000);
  }, [onComplete, onFailed]);

  // 页面 mount 时检查是否有正在进行的推演
  useEffect(() => {
    api.historyRunning().then((status) => {
      if (status.running && status.id) {
        setInput(status.input || '');
        startPolling(status.id);
      }
    });
    return () => stopPolling();
  }, [startPolling, stopPolling]);

  const handleGenerate = async () => {
    if (!input.trim() || loading) return;
    setError('');
    setResult(null);
    setLoading(true);
    setCurrentStep(0);
    try {
      const res = await api.generate(input, provider);
      startPolling(res.id);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : '生成失败';
      setError(msg);
      toast(msg, 'error');
      setLoading(false);
    }
  };

  const handleCancel = async () => {
    const id = pollingIdRef.current;
    if (!id) return;
    try {
      await api.cancelHistory(id);
    } catch {
      // ignore
    }
    stopPolling();
    setLoading(false);
    setError('已取消');
    api.me().then((me) => setCredits(me.credits)).catch(() => {});
  };

  const copyToClipboard = (text: string, idx: number) => {
    navigator.clipboard.writeText(text);
    setCopiedIdx(idx);
    setTimeout(() => setCopiedIdx(null), 2000);
  };

  const reviewerOutput = result?.reviewer_output as { prompts?: ReviewedPrompt[] } | undefined;
  const prompts: ReviewedPrompt[] = reviewerOutput?.prompts || [];

  return (
    <div className="max-w-4xl mx-auto">
      {/* 输入区 */}
      <div
        className="rounded-2xl p-6 mb-5 transition-all"
        style={{
          background: 'var(--mp-card-bg)',
          border: '1px solid var(--mp-card-border)',
          boxShadow: 'var(--mp-card-shadow)',
        }}
      >
        <div className="flex items-center gap-2 mb-4">
          <div className="text-base font-semibold" style={{ color: 'var(--mp-text-primary)' }}>
            描述你的需求
          </div>
          <div className="text-xs px-2 py-0.5 rounded-full" style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-primary)' }}>
            AI 将自动推演最优提示词
          </div>
        </div>
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="例如：我需要为一款新上市的运动鞋生成15张电商主图..."
          className="w-full h-28 px-4 py-3 rounded-xl text-sm resize-none outline-none transition-all"
          style={{
            background: 'var(--mp-primary-lighter)',
            border: '1px solid var(--mp-border-color)',
            color: 'var(--mp-text-primary)',
          }}
          disabled={loading}
        />
        <div className="flex items-center justify-between mt-4">
          <div className="flex items-center gap-3">
            <ModelSelect providers={providers} value={provider} onChange={setProvider} disabled={loading} />
            <span className="text-xs" style={{ color: 'var(--mp-text-secondary)' }}>
              剩余 {credits} 积分
            </span>
          </div>
          {loading ? (
            <button
              onClick={handleCancel}
              className="h-10 px-6 rounded-xl text-white text-sm font-medium transition-all"
              style={{ background: '#dc6464' }}
            >
              取消推演
            </button>
          ) : (
            <button
              onClick={handleGenerate}
              disabled={!input.trim()}
              className="h-10 px-6 rounded-xl text-white text-sm font-medium transition-all disabled:opacity-50"
              style={{ background: 'var(--mp-primary)' }}
            >
              开始推演
            </button>
          )}
        </div>
      </div>

      {/* 进度条 */}
      {loading && (
        <div
          className="rounded-2xl p-5 mb-5"
          style={{
            background: 'var(--mp-card-bg)',
            border: '1px solid var(--mp-card-border)',
            boxShadow: 'var(--mp-card-shadow)',
          }}
        >
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-medium" style={{ color: 'var(--mp-text-primary)' }}>
              推演进度
            </span>
            <span className="text-xs" style={{ color: 'var(--mp-text-secondary)' }}>
              {currentStep >= 0 && currentStep < 4 ? STEPS[currentStep] + '...' : '准备中...'}
            </span>
          </div>
          <div className="flex gap-2">
            {STEPS.map((step, i) => (
              <div key={step} className="flex-1">
                <div
                  className="h-2 rounded-full transition-all duration-500"
                  style={{
                    background: i < currentStep
                      ? 'var(--mp-primary)'
                      : i === currentStep
                        ? 'var(--mp-primary)'
                        : 'var(--mp-border-color)',
                    opacity: i === currentStep ? 0.6 : 1,
                    animation: i === currentStep ? 'pulse 1.5s ease-in-out infinite' : 'none',
                  }}
                />
                <div className="text-xs mt-1 text-center" style={{ color: i <= currentStep ? 'var(--mp-primary)' : 'var(--mp-text-secondary)' }}>
                  {step}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 错误 */}
      {error && (
        <div className="rounded-2xl px-5 py-4 mb-5 text-sm" style={{ background: 'rgba(220, 100, 100, 0.1)', color: '#dc6464' }}>
          {error}
        </div>
      )}

      {/* 结果 */}
      {result && prompts.length > 0 && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-base font-semibold" style={{ color: 'var(--mp-text-primary)' }}>
              推演结果
            </h3>
            <div className="flex items-center gap-2">
              <button
                onClick={() => {
                  const allText = prompts.map((p, i) => `## ${i + 1}. ${p.name}\n\n${p.prompt_text}`).join('\n\n---\n\n');
                  navigator.clipboard.writeText(allText);
                  setCopiedAll(true);
                  toast('已复制全部提示词', 'success');
                  setTimeout(() => setCopiedAll(false), 1500);
                }}
                className="text-xs px-2.5 py-1.5 rounded-lg transition-colors"
                style={{ background: copiedAll ? 'rgba(107,177,107,0.15)' : 'var(--mp-primary-lighter)', color: copiedAll ? 'var(--mp-success)' : 'var(--mp-primary)' }}
              >
                {copiedAll ? '已复制 ✓' : '复制全部'}
              </button>
              <button
                onClick={async () => {
                  if (!result?.id) return;
                  const md = await api.exportMarkdown(result.id);
                  const blob = new Blob([md], { type: 'text/markdown' });
                  const url = URL.createObjectURL(blob);
                  const a = document.createElement('a');
                  a.href = url;
                  a.download = `prompts_${result.id}.md`;
                  a.click();
                  URL.revokeObjectURL(url);
                }}
                className="text-xs px-2.5 py-1.5 rounded-lg transition-colors"
                style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-primary)' }}
              >
                导出 MD
              </button>
              <button
                onClick={async () => {
                  if (!result?.id) return;
                  const json = await api.exportJSON(result.id);
                  const blob = new Blob([json], { type: 'application/json' });
                  const url = URL.createObjectURL(blob);
                  const a = document.createElement('a');
                  a.href = url;
                  a.download = `prompts_${result.id}.json`;
                  a.click();
                  URL.revokeObjectURL(url);
                }}
                className="text-xs px-2.5 py-1.5 rounded-lg transition-colors"
                style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-primary)' }}
              >
                导出 JSON
              </button>
              <span className="text-xs" style={{ color: 'var(--mp-text-secondary)' }}>
                {prompts.length} 组 · {((result.duration_ms || 0) / 1000).toFixed(1)}s
              </span>
            </div>
          </div>

          <div
            className="rounded-2xl overflow-hidden divide-y"
            style={{
              background: 'var(--mp-card-bg)',
              border: '1px solid var(--mp-card-border)',
              boxShadow: 'var(--mp-card-shadow)',
              '--tw-divide-color': 'var(--mp-card-border)',
            } as React.CSSProperties}
          >
            {prompts.map((p, i) => (
              <div key={i}>
                <div
                  className="flex items-center gap-3 px-4 py-3 cursor-pointer select-none"
                  onClick={() => setExpandedIdx(expandedIdx === i ? null : i)}
                >
                  <span
                    className="shrink-0 w-5 h-5 rounded-full flex items-center justify-center text-[11px] font-semibold"
                    style={{ background: 'var(--mp-primary)', color: '#fff' }}
                  >
                    {p.order}
                  </span>
                  <div className="min-w-0 flex-1">
                    <span className="text-sm font-medium" style={{ color: 'var(--mp-text-primary)' }}>{p.name}</span>
                    {p.user_instruction && (
                      <span className="text-xs ml-2" style={{ color: 'var(--mp-text-secondary)' }}>{p.user_instruction}</span>
                    )}
                  </div>
                  <div className="shrink-0 flex items-center gap-1">
                    <button
                      onClick={(e) => { e.stopPropagation(); copyToClipboard(p.prompt_text, i); }}
                      className="w-7 h-7 rounded-md flex items-center justify-center transition-colors hover:opacity-80"
                      style={{
                        background: copiedIdx === i ? 'var(--mp-success)' : 'transparent',
                        color: copiedIdx === i ? '#fff' : 'var(--mp-text-secondary)',
                      }}
                      title="复制"
                    >
                      {copiedIdx === i ? (
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                      ) : (
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
                      )}
                    </button>
                    <svg
                      width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                      style={{ transition: 'transform 0.2s', transform: expandedIdx === i ? 'rotate(180deg)' : 'rotate(0deg)', color: 'var(--mp-text-secondary)' }}
                    >
                      <polyline points="6 9 12 15 18 9"/>
                    </svg>
                  </div>
                </div>
                {expandedIdx === i && (
                  <div className="px-4 pb-4">
                    <pre
                      className="text-sm whitespace-pre-wrap rounded-xl p-4 m-0 leading-relaxed"
                      style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-regular)' }}
                    >
                      {p.prompt_text}
                    </pre>
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* 推演过程 */}
          <div
            className="rounded-2xl overflow-hidden"
            style={{
              background: 'var(--mp-card-bg)',
              border: '1px solid var(--mp-card-border)',
              boxShadow: 'var(--mp-card-shadow)',
            }}
          >
            <button
              onClick={() => setExpandedIdx(expandedIdx === -1 ? null : -1)}
              className="w-full px-4 py-3 text-left text-sm transition-all flex items-center justify-between"
              style={{ color: 'var(--mp-text-secondary)' }}
            >
              <span>推演过程（Analyzer → Architect → Writer）</span>
              <svg
                className="w-4 h-4 transition-transform shrink-0"
                style={{ transform: expandedIdx === -1 ? 'rotate(180deg)' : 'rotate(0)' }}
                fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor"
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
              </svg>
            </button>
            {expandedIdx === -1 && (
              <div className="px-4 pb-4 space-y-3">
                {[
                  { label: 'Analyzer', data: result.reasoner_output },
                  { label: 'Architect', data: result.architect_output },
                  { label: 'Writer', data: result.generator_output },
                ].map((section) => (
                  <div key={section.label}>
                    <h4 className="text-xs font-medium mb-1.5" style={{ color: 'var(--mp-text-secondary)' }}>
                      {section.label}
                    </h4>
                    <pre
                      className="text-xs rounded-lg p-3 overflow-auto max-h-60 m-0"
                      style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-regular)' }}
                    >
                      {JSON.stringify(section.data, null, 2)}
                    </pre>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
