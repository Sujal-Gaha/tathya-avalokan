import React, { useEffect, useState } from 'react';
import { Toast, ToastType, useToastStore } from '../../store/useToastStore';

// ─── Icon mapping ────────────────────────────────────────────────────────────

const ICONS: Record<ToastType, string> = {
  success: '✓',
  error: '✕',
  warning: '⚠',
  info: 'ℹ',
};

const COLORS: Record<ToastType, { border: string; bg: string; icon: string; progress: string }> = {
  success: {
    border: 'border-emerald-500/40',
    bg: 'bg-emerald-500/10',
    icon: 'text-emerald-400',
    progress: 'bg-emerald-500',
  },
  error: {
    border: 'border-red-500/40',
    bg: 'bg-red-500/10',
    icon: 'text-red-400',
    progress: 'bg-red-500',
  },
  warning: {
    border: 'border-amber-500/40',
    bg: 'bg-amber-500/10',
    icon: 'text-amber-400',
    progress: 'bg-amber-500',
  },
  info: {
    border: 'border-sky-500/40',
    bg: 'bg-sky-500/10',
    icon: 'text-sky-400',
    progress: 'bg-sky-500',
  },
};

// ─── Single toast item ────────────────────────────────────────────────────────

const ToastItem: React.FC<{ toast: Toast }> = ({ toast }) => {
  const removeToast = useToastStore((s) => s.removeToast);
  const [visible, setVisible] = useState(false);

  // Animate in on mount
  useEffect(() => {
    const t = setTimeout(() => setVisible(true), 10);
    return () => clearTimeout(t);
  }, []);

  const colors = COLORS[toast.type];

  return (
    <div
      role="alert"
      aria-live="assertive"
      className={`
        relative flex items-start gap-3 w-80 px-4 py-3 rounded-lg border
        backdrop-blur-md bg-zinc-900/90 shadow-2xl
        transition-all duration-300 ease-out overflow-hidden
        ${colors.border}
        ${visible ? 'opacity-100 translate-x-0' : 'opacity-0 translate-x-8'}
      `}
    >
      {/* Type icon */}
      <span className={`mt-0.5 text-sm font-bold flex-shrink-0 ${colors.icon}`}>
        {ICONS[toast.type]}
      </span>

      {/* Message */}
      <p className="flex-1 text-sm text-zinc-200 leading-snug break-words">{toast.message}</p>

      {/* Dismiss button */}
      <button
        id={`toast-dismiss-${toast.id}`}
        onClick={() => removeToast(toast.id)}
        className="flex-shrink-0 mt-0.5 text-zinc-500 hover:text-zinc-200 transition-colors text-xs leading-none"
        aria-label="Dismiss notification"
      >
        ✕
      </button>

      {/* Auto-dismiss progress bar */}
      <div
        className={`absolute bottom-0 left-0 h-0.5 ${colors.progress} opacity-60`}
        style={{ animation: 'toast-progress 4s linear forwards' }}
      />
    </div>
  );
};

// ─── Toast container ──────────────────────────────────────────────────────────

export const ToastContainer: React.FC = () => {
  const toasts = useToastStore((s) => s.toasts);

  return (
    <>
      {/* Keyframe for progress bar */}
      <style>{`
        @keyframes toast-progress {
          from { width: 100%; }
          to   { width: 0%; }
        }
      `}</style>

      <div
        id="toast-container"
        aria-label="Notifications"
        className="fixed bottom-6 right-6 z-50 flex flex-col gap-2 items-end pointer-events-none"
      >
        {toasts.map((toast) => (
          <div key={toast.id} className="pointer-events-auto">
            <ToastItem toast={toast} />
          </div>
        ))}
      </div>
    </>
  );
};
