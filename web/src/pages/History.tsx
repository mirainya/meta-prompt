import { useState, useEffect } from 'react';
import { api, type HistoryItem } from '../lib/api';

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

  useEffect(() => {
    api.history(id).then(setDetail).finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="px-5 pb-4 text-sm" style={{ color: 'var(--mp-text-secondary)' }}>加载中...</div>
    );
  }
  if (!detail) {
    return (
      <div className="px-5 pb-4 text-sm" style={{ color: 'var(--mp-danger)' }}>加载失败</div>
    );
  }

  return (
    <div className="px-5 pb-5">
      <pre
        className="text-xs rounded-xl p-4 overflow-auto max-h-96 m-0"
        style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-regular)' }}
      >
        {JSON.stringify(detail.generator_output, null, 2)}
      </pre>
    </div>
  );
}
