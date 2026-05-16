import { useState } from 'react';

const endpoints = [
  {
    method: 'POST', path: '/open/v1/generate', title: '创建推演任务',
    desc: '提交需求文本，异步生成提示词。消耗 1 积分。',
    params: [
      { name: 'input', type: 'string', required: true, desc: '需求描述' },
      { name: 'llm_provider', type: 'string', required: false, desc: 'LLM 提供商（可选）' },
      { name: 'mode', type: 'string', required: false, desc: '"sync" 同步 / "async" 异步（默认）' },
      { name: 'webhook_url', type: 'string', required: false, desc: '完成后回调地址' },
    ],
    response: '{ "id": 42, "status": "running" }',
    example: `curl -X POST {BASE}/open/v1/generate \\
  -H "Content-Type: application/json" \\
  -H "X-API-Key: mp_your_key" \\
  -d '{"input": "写一个代码审查提示词"}'`,
  },
  {
    method: 'GET', path: '/open/v1/tasks/{id}', title: '查询任务状态',
    desc: '根据 task_id 查询推演进度和结果。完成后包含 reviewer_output。',
    params: [{ name: 'id', type: 'integer', required: true, desc: '任务 ID（路径参数）' }],
    response: '{ "id": 42, "status": "done", "current_step": 4, "reviewer_output": {...}, "duration_ms": 12345 }',
    example: `curl {BASE}/open/v1/tasks/42 \\
  -H "X-API-Key: mp_your_key"`,
  },
  {
    method: 'GET', path: '/open/v1/tasks/{id}/stream', title: 'SSE 实时进度流',
    desc: '通过 Server-Sent Events 实时接收推演进度。事件：step / done / failed。',
    params: [{ name: 'id', type: 'integer', required: true, desc: '任务 ID（路径参数）' }],
    response: 'event: step\\ndata: {"current_step":2}\\n\\nevent: done\\ndata: {"reviewer_output":{...}}',
    example: `curl -N {BASE}/open/v1/tasks/42/stream \\
  -H "X-API-Key: mp_your_key"`,
  },
  {
    method: 'GET', path: '/open/v1/tasks/{id}/export', title: '导出结果',
    desc: '导出推演结果为 Markdown 或 JSON。',
    params: [
      { name: 'id', type: 'integer', required: true, desc: '任务 ID（路径参数）' },
      { name: 'format', type: 'string', required: false, desc: '"markdown" 或 "json"（默认）' },
    ],
    response: '{ "prompts": [{ "name": "...", "prompt_text": "..." }] }',
    example: `curl {BASE}/open/v1/tasks/42/export?format=json \\
  -H "X-API-Key: mp_your_key"`,
  },
  {
    method: 'POST', path: '/open/v1/tasks/{id}/cancel', title: '取消任务',
    desc: '取消正在运行的推演任务，退还积分。',
    params: [{ name: 'id', type: 'integer', required: true, desc: '任务 ID（路径参数）' }],
    response: '{ "message": "cancelled" }',
    example: `curl -X POST {BASE}/open/v1/tasks/42/cancel \\
  -H "X-API-Key: mp_your_key"`,
  },
  {
    method: 'GET', path: '/open/v1/providers', title: '获取可用模型列表',
    desc: '返回当前已启用的 LLM 提供商及其模型。',
    params: [],
    response: '[{ "provider": "openai", "model": "gpt-4o" }]',
    example: `curl {BASE}/open/v1/providers \\
  -H "X-API-Key: mp_your_key"`,
  },
];

const MC: Record<string, string> = { GET: '#4caf50', POST: '#2196f3' };

export default function APIDocs() {
  const [open, setOpen] = useState<number | null>(0);
  const base = window.location.origin;

  const exportJSON = () => {
    window.open(`${base}/open/v1/docs`, '_blank');
  };

  return (
    <div className="max-w-4xl mx-auto space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold" style={{ color: 'var(--mp-text-primary)' }}>Open API 文档</h2>
          <p className="text-xs mt-1" style={{ color: 'var(--mp-text-secondary)' }}>
            所有接口需在 Header 中携带 <code className="px-1.5 py-0.5 rounded" style={{ background: 'var(--mp-primary-lighter)' }}>X-API-Key</code> 鉴权。
            前往 <a href="/api-keys" className="underline" style={{ color: 'var(--mp-primary)' }}>API Key 管理</a> 创建密钥。
          </p>
        </div>
        <button
          onClick={exportJSON}
          className="h-9 px-4 rounded-xl text-xs font-medium"
          style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-primary)' }}
        >
          导出 OpenAPI JSON
        </button>
      </div>

      <div className="space-y-3">
        {endpoints.map((ep, i) => (
          <div key={i} className="rounded-2xl overflow-hidden" style={{ background: 'var(--mp-card-bg)', border: '1px solid var(--mp-card-border)', boxShadow: 'var(--mp-card-shadow)' }}>
            <button onClick={() => setOpen(open === i ? null : i)} className="w-full px-5 py-4 flex items-center gap-3 text-left">
              <span className="shrink-0 text-xs font-bold px-2 py-0.5 rounded" style={{ background: MC[ep.method] + '18', color: MC[ep.method] }}>{ep.method}</span>
              <code className="text-sm font-mono" style={{ color: 'var(--mp-text-primary)' }}>{ep.path}</code>
              <span className="ml-auto text-xs" style={{ color: 'var(--mp-text-secondary)' }}>{ep.title}</span>
              <svg className={`w-4 h-4 shrink-0 transition-transform ${open === i ? 'rotate-180' : ''}`} fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" style={{ color: 'var(--mp-text-secondary)' }}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
              </svg>
            </button>

            {open === i && (
              <div className="px-5 pb-5 space-y-4" style={{ borderTop: '1px solid var(--mp-border-color)' }}>
                <p className="text-sm pt-3" style={{ color: 'var(--mp-text-regular)' }}>{ep.desc}</p>

                {ep.params.length > 0 && (
                  <div>
                    <h4 className="text-xs font-semibold mb-2" style={{ color: 'var(--mp-text-secondary)' }}>参数</h4>
                    <div className="rounded-xl overflow-hidden" style={{ border: '1px solid var(--mp-border-color)' }}>
                      <table className="w-full text-xs">
                        <thead>
                          <tr style={{ background: 'var(--mp-primary-lighter)' }}>
                            <th className="text-left px-3 py-2 font-medium" style={{ color: 'var(--mp-text-secondary)' }}>字段</th>
                            <th className="text-left px-3 py-2 font-medium" style={{ color: 'var(--mp-text-secondary)' }}>类型</th>
                            <th className="text-left px-3 py-2 font-medium" style={{ color: 'var(--mp-text-secondary)' }}>必填</th>
                            <th className="text-left px-3 py-2 font-medium" style={{ color: 'var(--mp-text-secondary)' }}>说明</th>
                          </tr>
                        </thead>
                        <tbody>
                          {ep.params.map((p) => (
                            <tr key={p.name} style={{ borderTop: '1px solid var(--mp-border-color)' }}>
                              <td className="px-3 py-2 font-mono" style={{ color: 'var(--mp-text-primary)' }}>{p.name}</td>
                              <td className="px-3 py-2" style={{ color: 'var(--mp-text-secondary)' }}>{p.type}</td>
                              <td className="px-3 py-2">{p.required ? <span style={{ color: MC.POST }}>是</span> : <span style={{ color: 'var(--mp-text-secondary)' }}>否</span>}</td>
                              <td className="px-3 py-2" style={{ color: 'var(--mp-text-regular)' }}>{p.desc}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}

                <div>
                  <h4 className="text-xs font-semibold mb-2" style={{ color: 'var(--mp-text-secondary)' }}>响应示例</h4>
                  <pre className="text-xs p-3 rounded-xl overflow-x-auto font-mono" style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-regular)' }}>{ep.response}</pre>
                </div>

                <div>
                  <h4 className="text-xs font-semibold mb-2" style={{ color: 'var(--mp-text-secondary)' }}>调用示例</h4>
                  <pre className="text-xs p-3 rounded-xl overflow-x-auto font-mono" style={{ background: 'var(--mp-primary-lighter)', color: 'var(--mp-text-regular)' }}>{ep.example.replace(/\{BASE\}/g, base)}</pre>
                </div>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
