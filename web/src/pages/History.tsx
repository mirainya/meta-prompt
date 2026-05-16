import { useState, useEffect } from 'react';
import { api, type HistoryItem, type ReviewedPrompt } from '../lib/api';

export default function History() {
  const [items, setItems] = useState<HistoryItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedId, setExpandedId] = useState<number | null>(null);

  useEffect(() => {
    api.histories(50, 0).then(setItems).finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="max-w-4xl mx-auto flex items-center justify-center py-20">
        <div className="text-sm" style={{ color: 'var(--mp-text-secondary)' }}>加载中...</div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <h2 className="text-base font-semibold" style={{ color: 'var(--mp-text-primary)' }}>
            历史记录
          </h2>
          <span
            className="text-xs px-2 py-0.5 rounded-full font-medium"
            style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-primary)' }}
          >
            {items.length} 条
          </span>
        </div>
      </div>

      {items.length === 0 ? (
        <div
          className="rounded-2xl py-16 text-center"
          style={{
            background: 'var(--mp-card-bg)',
            border: '1px solid var(--mp-card-border)',
            boxShadow: 'var(--mp-card-shadow)',
          }}
        >
          <svg className="w-12 h-12 mx-auto mb-3" style={{ color: 'var(--mp-border-color)' }} fill="none" viewBox="0 0 24 24" strokeWidth={1} stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="text-sm" style={{ color: 'var(--mp-text-secondary)' }}>暂无记录，去推演一下吧</p>
        </div>
      ) : (
        <div className="space-y-3">
          {items.map((item) => (
            <div
              key={item.id}
              className="rounded-2xl overflow-hidden transition-all"
              style={{
                background: 'var(--mp-card-bg)',
                border: '1px solid var(--mp-card-border)',
                boxShadow: 'var(--mp-card-shadow)',
              }}
            >
              <button
                onClick={() => setExpandedId(expandedId === item.id ? null : item.id)}
                className="w-full px-5 py-4 text-left transition-all flex items-center justify-between"
              >
                <div className="flex-1 min-w-0">
                  <p className="text-sm truncate" style={{ color: 'var(--mp-text-primary)' }}>{item.input}</p>
                  <div className="flex items-center gap-2 mt-1.5">
                    <span
                      className="text-xs px-2 py-0.5 rounded-full"
                      style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-primary)' }}
                    >
                      {item.llm_provider}
                    </span>
                    {item.status === 'running' && (
                      <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: 'rgba(59, 130, 246, 0.1)', color: '#3b82f6' }}>
                        推演中...
                      </span>
                    )}
                    {item.status === 'failed' && (
                      <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: 'rgba(220, 100, 100, 0.1)', color: '#dc6464' }}>
                        失败
                      </span>
                    )}
                    <span className="text-xs" style={{ color: 'var(--mp-text-secondary)' }}>
                      {new Date(item.created_at).toLocaleString('zh-CN')}
                    </span>
                    {item.status === 'done' && (
                      <span className="text-xs" style={{ color: 'var(--mp-text-secondary)' }}>
                        {(item.duration_ms / 1000).toFixed(1)}s
                      </span>
                    )}
                  </div>
                </div>
                <svg
                  className="w-4 h-4 shrink-0 ml-3 transition-transform"
                  style={{
                    color: 'var(--mp-text-secondary)',
                    transform: expandedId === item.id ? 'rotate(180deg)' : 'rotate(0)',
                  }}
                  fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                </svg>
              </button>
              {expandedId === item.id && (
                <HistoryDetail id={item.id} />
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function HistoryDetail({ id }: { id: number }) {
  const [detail, setDetail] = useState<HistoryItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [expandedIdx, setExpandedIdx] = useState<number | null>(null);
  const [copiedIdx, setCopiedIdx] = useState<number | null>(null);

  useEffect(() => {
    api.history(id).then(setDetail).finally(() => setLoading(false));
  }, [id]);

  const copyText = (text: string, idx: number) => {
    navigator.clipboard.writeText(text);
    setCopiedIdx(idx);
    setTimeout(() => setCopiedIdx(null), 1500);
  };

  if (loading) {
    return <div className="px-5 pb-4 text-sm" style={{ color: 'var(--mp-text-secondary)' }}>加载中...</div>;
  }
  if (!detail) {
    return <div className="px-5 pb-4 text-sm" style={{ color: 'var(--mp-danger)' }}>加载失败</div>;
  }

  const reviewer = detail.reviewer_output as { prompts?: ReviewedPrompt[] } | null;
  const prompts = reviewer?.prompts;

  if (!prompts?.length) {
    return (
      <div className="px-5 pb-5">
        <pre
          className="text-xs rounded-xl p-4 overflow-auto max-h-96 m-0"
          style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-regular)' }}
        >
          {JSON.stringify(detail.generator_output || detail.reviewer_output, null, 2)}
        </pre>
      </div>
    );
  }

  return (
    <div className="px-5 pb-5 space-y-2">
      {prompts.map((p, i) => (
        <div
          key={i}
          className="rounded-xl overflow-hidden transition-all"
          style={{ background: 'var(--mp-primary-lighter)', border: '1px solid var(--mp-border-color)' }}
        >
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
            <span className="text-sm font-medium flex-1 min-w-0 truncate" style={{ color: 'var(--mp-text-primary)' }}>
              {p.name}
            </span>
            <button
              onClick={(e) => { e.stopPropagation(); copyText(p.prompt_text, i); }}
              className="w-7 h-7 rounded-md flex items-center justify-center transition-colors"
              style={{ color: copiedIdx === i ? 'var(--mp-success)' : 'var(--mp-text-secondary)' }}
              title="复制"
            >
              {copiedIdx === i ? (
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><polyline points="20 6 9 17 4 12"/></svg>
              ) : (
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
              )}
            </button>
            <svg
              width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
              style={{ transition: 'transform 0.2s', transform: expandedIdx === i ? 'rotate(180deg)' : '', color: 'var(--mp-text-secondary)' }}
            >
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </div>
          {expandedIdx === i && (
            <div className="px-4 pb-4">
              <pre
                className="text-sm whitespace-pre-wrap rounded-lg p-3 m-0 leading-relaxed overflow-auto max-h-80"
                style={{ background: 'var(--mp-card-bg)', color: 'var(--mp-text-regular)' }}
              >
                {p.prompt_text}
              </pre>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
