import { useState, useEffect } from 'react';
import { create } from 'zustand';

interface ToastItem {
  id: number;
  message: string;
  type: 'success' | 'error' | 'info';
}

interface ToastState {
  toasts: ToastItem[];
  add: (message: string, type?: ToastItem['type']) => void;
  remove: (id: number) => void;
}

let nextId = 0;

export const useToast = create<ToastState>((set) => ({
  toasts: [],
  add: (message, type = 'info') => {
    const id = ++nextId;
    set((s) => ({ toasts: [...s.toasts, { id, message, type }] }));
    setTimeout(() => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })), 3000);
  },
  remove: (id) => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })),
}));

export function ToastContainer() {
  const toasts = useToast((s) => s.toasts);
  const remove = useToast((s) => s.remove);

  if (!toasts.length) return null;

  return (
    <div className="fixed top-5 right-5 z-[99999] space-y-2" style={{ maxWidth: 360 }}>
      {toasts.map((t) => (
        <Toast key={t.id} toast={t} onClose={() => remove(t.id)} />
      ))}
    </div>
  );
}

function Toast({ toast, onClose }: { toast: ToastItem; onClose: () => void }) {
  const [visible, setVisible] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => setVisible(false), 2700);
    return () => clearTimeout(timer);
  }, []);

  const colors = {
    success: { bg: 'var(--mp-success)', icon: 'M20 6L9 17l-5-5' },
    error: { bg: 'var(--mp-danger)', icon: 'M6 18L18 6M6 6l12 12' },
    info: { bg: 'var(--mp-primary)', icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
  };
  const c = colors[toast.type];

  return (
    <div
      onClick={onClose}
      className="flex items-center gap-3 px-4 py-3 rounded-xl cursor-pointer text-sm shadow-lg transition-all duration-300"
      style={{
        background: 'var(--mp-card-bg)',
        border: '1px solid var(--mp-card-border)',
        backdropFilter: 'blur(12px)',
        color: 'var(--mp-text-primary)',
        opacity: visible ? 1 : 0,
        transform: visible ? 'translateX(0)' : 'translateX(20px)',
      }}
    >
      <svg className="w-5 h-5 shrink-0" style={{ color: c.bg }} fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" d={c.icon} />
      </svg>
      <span className="flex-1">{toast.message}</span>
    </div>
  );
}
