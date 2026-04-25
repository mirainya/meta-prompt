import { useState, useEffect } from 'react';
import { api, type AdminDashboard as DashboardData } from '../../lib/api';

function StatCard({ title, value, color }: { title: string; value: number | string; color: string }) {
  return (
    <div
      className="rounded-2xl p-5 transition-all hover:-translate-y-0.5"
      style={{
        background: 'var(--mp-card-bg)',
        border: '1px solid var(--mp-card-border)',
        boxShadow: 'var(--mp-card-shadow)',
      }}
    >
      <div className="text-sm font-medium mb-2" style={{ color: 'var(--mp-text-secondary)' }}>{title}</div>
      <div className="text-3xl font-bold" style={{ color }}>{value}</div>
    </div>
  );
}

export default function AdminDashboard() {
  const [data, setData] = useState<DashboardData | null>(null);

  useEffect(() => {
    api.adminDashboard().then(setData);
  }, []);

  if (!data) {
    return <div className="text-sm" style={{ color: 'var(--mp-text-secondary)' }}>加载中...</div>;
  }

  return (
    <div className="max-w-4xl mx-auto">
      <h2 className="text-base font-semibold mb-4" style={{ color: 'var(--mp-text-primary)' }}>管理仪表盘</h2>
      <div className="grid grid-cols-3 gap-5">
        <StatCard title="注册用户" value={data.user_count} color="var(--mp-primary)" />
        <StatCard title="总生成次数" value={data.total_generations} color="var(--mp-success)" />
        <StatCard title="今日生成" value={data.today_generations} color="var(--mp-warning)" />
      </div>
    </div>
  );
}
