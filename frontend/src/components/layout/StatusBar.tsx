import React from 'react';
import { CheckCircle2, Clock, Layers } from 'lucide-react';
import { useUIStore } from '../../store/useUIStore';

export const StatusBar: React.FC = () => {
  const { tabs, activeTabId } = useUIStore();
  const currentTab = tabs.find((t) => t.id === activeTabId);

  return (
    <footer className="h-6 bg-[#18181b] border-t border-zinc-800 px-3 flex items-center justify-between text-[11px] text-zinc-400 select-none">
      <div className="flex items-center space-x-4">
        <div className="flex items-center space-x-1 text-emerald-400">
          <CheckCircle2 className="w-3 h-3" />
          <span>BFF Proxy Connected</span>
        </div>
        {currentTab?.result && (
          <div className="flex items-center space-x-1 text-zinc-400">
            <Clock className="w-3 h-3 text-zinc-500" />
            <span>{currentTab.result.execution_time_ms} ms</span>
          </div>
        )}
        {currentTab?.result && (
          <div className="flex items-center space-x-1 text-zinc-400">
            <Layers className="w-3 h-3 text-zinc-500" />
            <span>{currentTab.result.rows_affected} rows</span>
          </div>
        )}
      </div>

      <div className="flex items-center space-x-3 text-zinc-500">
        <span>SQLite Metadata Storage</span>
        <span>UTF-8</span>
      </div>
    </footer>
  );
};
