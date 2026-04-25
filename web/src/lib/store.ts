import { create } from 'zustand';

interface AuthState {
  token: string | null;
  username: string | null;
  role: string | null;
  credits: number;
  setAuth: (token: string, username: string, credits: number, role?: string) => void;
  setCredits: (credits: number) => void;
  logout: () => void;
}

export const useAuth = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  username: localStorage.getItem('username'),
  role: localStorage.getItem('role'),
  credits: Number(localStorage.getItem('credits') || 0),

  setAuth: (token, username, credits, role = 'user') => {
    localStorage.setItem('token', token);
    localStorage.setItem('username', username);
    localStorage.setItem('credits', String(credits));
    localStorage.setItem('role', role);
    set({ token, username, credits, role });
  },

  setCredits: (credits) => {
    localStorage.setItem('credits', String(credits));
    set({ credits });
  },

  logout: () => {
    localStorage.removeItem('token');
    localStorage.removeItem('username');
    localStorage.removeItem('credits');
    localStorage.removeItem('role');
    set({ token: null, username: null, credits: 0, role: null });
  },
}));
