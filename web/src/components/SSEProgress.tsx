import { useState, useEffect } from 'react';

const STEPS = [
  { name: 'Analyzer', desc: '需求分析' },
  { name: 'Architect', desc: '结构规划' },
  { name: 'Writer', desc: '内容生成' },
  { name: 'Reviewer', desc: '质量审查' },
];

interface SSEEvent {
  step: number;
  name: string;
  status: string;
  progress?: string;
  error?: string;
}

export default function SSEProgress({ historyId, onDone }: { historyId: number; onDone: () => void }) {
  const [events, setEvents] = useState<SSEEvent[]>([]);
  const [elapsed, setElapsed] = useState(0);

  useEffect(() => {
    const t = setInterval(() => setElapsed((s) => s + 1), 1000);
    return () => clearInterval(t);
  }, []);

  useEffect(() => {
    const token = localStorage.getItem('token');
    const es = new EventSource(`/api/v1/histories/${historyId}/stream?token=${token}`);

    es.addEventListener('done', () => { es.close(); onDone(); });
    es.addEventListener('failed', (e) => {
      setEvents((prev) => [...prev, JSON.parse((e as MessageEvent).data)]);
      es.close();
    });
    es.addEventListener('running', (e) => {
      const data = JSON.parse((e as MessageEvent).data);
      setEvents((prev) => [...prev.filter((ev) => ev.step !== data.step), data]);
    });
    es.onmessage = (e) => {
      const data = JSON.parse(e.data);
      setEvents((prev) => [...prev, data]);
      if (data.name === 'done') { es.close(); onDone(); }
    };
    es.onerror = () => es.close();
    return () => es.close();
  }, [historyId, onDone]);

  const getStatus = (step: number) => events.find((e) => e.step === step)?.status || 'pending';
  const getProgress = (step: number) => events.find((e) => e.step === step)?.progress;
  const failed = events.find((e) => e.status === 'failed');

  return (
    <div className="py-2">
      <div className="flex items-center justify-between mb-4">
        <span className="text-xs font-medium" style={{ color: 'var(--mp-text-secondary)' }}>
          推演进行中
        </span>
        <span className="text-xs tabular-nums" style={{ color: 'var(--mp-text-secondary)' }}>
          {elapsed}s
        </span>
      </div>

      <div className="relative">
        {STEPS.map((s, i) => {
          const status = getStatus(i + 1);
          const progress = getProgress(i + 1);
          const isLast = i === STEPS.length - 1;

          return (
            <div key={i} className="flex gap-3" style={{ paddingBottom: isLast ? 0 : 20 }}>
              {/* 竖线 + 圆点 */}
              <div className="flex flex-col items-center">
                <div
                  className="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0 transition-all duration-300"
                  style={{
                    background: status === 'done' ? 'var(--mp-success)'
                      : status === 'running' ? 'var(--mp-primary)'
                      : status === 'failed' ? 'var(--mp-danger)'
                      : 'var(--mp-border-color)',
                    color: status === 'pending' ? 'var(--mp-text-secondary)' : '#fff',
                    boxShadow: status === 'running' ? '0 0 0 4px var(--mp-primary-lighter)' : 'none',
                  }}
                >
                  {status === 'done' ? '✓' : status === 'failed' ? '!' : i + 1}
                </div>
                {!isLast && (
                  <div
                    className="w-0.5 flex-1 mt-1 rounded transition-all duration-300"
                    style={{
                      background: status === 'done' ? 'var(--mp-success)' : 'var(--mp-border-color)',
                    }}
                  />
                )}
              </div>

              {/* 内容 */}
              <div className="pt-0.5 min-w-0">
                <div className="flex items-center gap-2">
                  <span
                    className="text-sm font-medium transition-colors"
                    style={{
                      color: status === 'running' ? 'var(--mp-primary)'
                        : status === 'done' ? 'var(--mp-text-primary)'
                        : 'var(--mp-text-secondary)',
                    }}
                  >
                    {s.name}
                  </span>
                  <span className="text-xs" style={{ color: 'var(--mp-text-secondary)' }}>
                    {s.desc}
                  </span>
                  {status === 'running' && (
                    <div className="w-3 h-3 border-2 border-current border-t-transparent rounded-full animate-spin" style={{ color: 'var(--mp-primary)' }} />
                  )}
                </div>
                {progress && (
                  <p className="text-xs mt-0.5" style={{ color: 'var(--mp-text-secondary)' }}>{progress}</p>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {failed && (
        <div className="mt-3 px-3 py-2 rounded-lg text-xs" style={{ background: 'rgba(212, 122, 122, 0.08)', color: 'var(--mp-danger)' }}>
          {failed.error}
        </div>
      )}
    </div>
  );
}
