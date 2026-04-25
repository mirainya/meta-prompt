import { BrowserRouter, Routes, Route, Navigate, NavLink, useNavigate, useLocation } from 'react-router-dom';
import { useEffect } from 'react';
import { useAuth } from './lib/store';
import { api } from './lib/api';
import Login from './pages/Login';
import Register from './pages/Register';
import Generate from './pages/Generate';
import History from './pages/History';
import AdminDashboard from './pages/admin/Dashboard';
import LLMConfig from './pages/admin/LLMConfig';
import Users from './pages/admin/Users';
import Templates from './pages/admin/Templates';

const userMenuItems = [
  { path: '/', label: '提示词推演', icon: 'M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.455 2.456L21.75 6l-1.036.259a3.375 3.375 0 00-2.455 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z' },
  { path: '/history', label: '历史记录', icon: 'M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z' },
];

const adminMenuItems = [
  { path: '/admin', label: '仪表盘', icon: 'M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z' },
  { path: '/admin/llm', label: '模型配置', icon: 'M4.5 12a7.5 7.5 0 0015 0m-15 0a7.5 7.5 0 1115 0m-15 0H3m16.5 0H21m-1.5 0H12m-8.457 3.077l1.41-.513m14.095-5.13l1.41-.513M5.106 17.785l1.15-.964m11.49-9.642l1.149-.964M7.501 19.795l.75-1.3m7.5-12.99l.75-1.3m-6.063 16.658l.26-1.477m2.605-14.772l.26-1.477m0 17.726l-.26-1.477M10.698 4.614l-.26-1.477M16.5 19.794l-.75-1.299M7.5 4.205L12 12m6.894 5.785l-1.149-.964M6.256 7.178l-1.15-.964m15.352 8.864l-1.41-.513M4.954 9.435l-1.41-.514M12.002 12l-3.75 6.495' },
  { path: '/admin/users', label: '用户管理', icon: 'M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z' },
  { path: '/admin/templates', label: '元提示词', icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z' },
];

function Layout({ children }: { children: React.ReactNode }) {
  const username = useAuth((s) => s.username);
  const credits = useAuth((s) => s.credits);
  const role = useAuth((s) => s.role);
  const logout = useAuth((s) => s.logout);
  const setCredits = useAuth((s) => s.setCredits);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    api.me().then((me) => setCredits(me.credits)).catch(() => {});
  }, [setCredits]);

  const isAdmin = role === 'admin';
  const allMenuItems = [...userMenuItems, ...(isAdmin ? adminMenuItems : [])];

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="h-screen flex overflow-hidden">
      {/* 侧边栏 */}
      <aside
        className="w-[220px] flex flex-col shrink-0 transition-all"
        style={{
          background: 'var(--mp-sidebar)',
          borderRight: '1px solid var(--mp-card-border)',
        }}
      >
        <div
          className="h-[60px] flex items-center justify-center gap-2.5"
          style={{ borderBottom: '1px solid var(--mp-border-color)' }}
        >
          <div
            className="w-8 h-8 rounded-lg flex items-center justify-center text-white text-sm font-bold"
            style={{ background: 'var(--mp-primary)' }}
          >
            M
          </div>
          <span className="text-[15px] font-bold tracking-wide" style={{ color: 'var(--mp-sidebar-text)' }}>
            Meta Prompt
          </span>
        </div>

        <nav className="flex-1 p-2 space-y-1 overflow-y-auto">
          {userMenuItems.map((item) => {
            const isActive = location.pathname === item.path;
            return (
              <NavLink
                key={item.path}
                to={item.path}
                className="flex items-center gap-3 h-11 px-3 rounded-[10px] text-sm font-medium transition-all"
                style={{
                  color: isActive ? 'var(--mp-sidebar-active-text)' : 'var(--mp-sidebar-text)',
                  background: isActive ? 'var(--mp-sidebar-active-bg)' : 'transparent',
                  fontWeight: isActive ? 600 : 500,
                }}
              >
                <svg className="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" d={item.icon} />
                </svg>
                {item.label}
              </NavLink>
            );
          })}

          {isAdmin && (
            <>
              <div className="pt-3 pb-1 px-3">
                <div className="text-[10px] font-semibold tracking-widest uppercase" style={{ color: 'var(--mp-text-secondary)' }}>
                  管理
                </div>
              </div>
              {adminMenuItems.map((item) => {
                const isActive = location.pathname === item.path;
                return (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    className="flex items-center gap-3 h-11 px-3 rounded-[10px] text-sm font-medium transition-all"
                    style={{
                      color: isActive ? 'var(--mp-sidebar-active-text)' : 'var(--mp-sidebar-text)',
                      background: isActive ? 'var(--mp-sidebar-active-bg)' : 'transparent',
                      fontWeight: isActive ? 600 : 500,
                    }}
                  >
                    <svg className="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" d={item.icon} />
                    </svg>
                    {item.label}
                  </NavLink>
                );
              })}
            </>
          )}
        </nav>

        {/* 底部积分 */}
        <div className="p-3">
          <div
            className="rounded-xl p-3 text-center"
            style={{ background: 'var(--mp-primary-lighter)', border: '1px solid var(--mp-border-color)' }}
          >
            <div className="text-2xl font-bold" style={{ color: 'var(--mp-primary)' }}>{credits}</div>
            <div className="text-xs mt-0.5" style={{ color: 'var(--mp-text-secondary)' }}>剩余积分</div>
          </div>
        </div>
      </aside>

      {/* 右侧 */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* 顶栏 */}
        <header
          className="h-[60px] flex items-center justify-between px-6 shrink-0 z-10"
          style={{
            background: 'var(--mp-header-bg)',
            backdropFilter: 'blur(12px)',
            boxShadow: '0 1px 8px rgba(0, 0, 0, 0.04)',
            borderBottom: '1px solid var(--mp-card-border)',
          }}
        >
          <div className="text-sm font-medium" style={{ color: 'var(--mp-text-regular)' }}>
            {allMenuItems.find((m) => m.path === location.pathname)?.label || ''}
          </div>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <div
                className="w-8 h-8 rounded-full flex items-center justify-center text-white text-sm font-semibold"
                style={{ background: 'var(--mp-primary)' }}
              >
                {username?.charAt(0).toUpperCase()}
              </div>
              <span className="text-sm font-medium" style={{ color: 'var(--mp-text-regular)' }}>
                {username}
              </span>
              {isAdmin && (
                <span className="text-[10px] px-1.5 py-0.5 rounded" style={{ background: 'rgba(154, 138, 189, 0.12)', color: 'var(--mp-accent)' }}>
                  Admin
                </span>
              )}
            </div>
            <button
              onClick={handleLogout}
              className="p-1.5 rounded-lg transition-colors"
              style={{ color: 'var(--mp-text-secondary)' }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.color = 'var(--mp-danger)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.color = 'var(--mp-text-secondary)'; }}
              title="退出登录"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15m3 0l3-3m0 0l-3-3m3 3H9" />
              </svg>
            </button>
          </div>
        </header>

        {/* 内容区 */}
        <main className="flex-1 overflow-y-auto p-5" style={{ background: 'var(--mp-body-bg)' }}>
          {children}
        </main>
      </div>
    </div>
  );
}

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const token = useAuth((s) => s.token);
  if (!token) return <Navigate to="/login" replace />;
  return <Layout>{children}</Layout>;
}

function AdminRoute({ children }: { children: React.ReactNode }) {
  const token = useAuth((s) => s.token);
  const role = useAuth((s) => s.role);
  if (!token) return <Navigate to="/login" replace />;
  if (role !== 'admin') return <Navigate to="/" replace />;
  return <Layout>{children}</Layout>;
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/" element={<PrivateRoute><Generate /></PrivateRoute>} />
        <Route path="/history" element={<PrivateRoute><History /></PrivateRoute>} />
        <Route path="/admin" element={<AdminRoute><AdminDashboard /></AdminRoute>} />
        <Route path="/admin/llm" element={<AdminRoute><LLMConfig /></AdminRoute>} />
        <Route path="/admin/users" element={<AdminRoute><Users /></AdminRoute>} />
        <Route path="/admin/templates" element={<AdminRoute><Templates /></AdminRoute>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
