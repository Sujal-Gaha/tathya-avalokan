import React from 'react';
import { Database, FolderKanban, Play, Plus, RefreshCw } from 'lucide-react';
import { useUIStore } from '../../store/useUIStore';

interface HeaderProps {
  onNewProject: () => void;
  onNewInstance: () => void;
  onExecuteQuery: () => void;
}

export const Header: React.FC<HeaderProps> = ({ onNewProject, onNewInstance, onExecuteQuery }) => {
  const { tabs, activeTabId } = useUIStore();
  const currentTab = tabs.find((t) => t.id === activeTabId);

  return (
    <header className="h-14 border-b border-zinc-800 bg-[#121215] flex items-center justify-between px-4 select-none">
      <div className="flex items-center space-x-3">
        <div className="flex items-center space-x-2 text-blue-500 font-semibold tracking-wide">
          <Database className="w-5 h-5" />
          <span className="text-zinc-100">Tathya<span className="text-blue-500">-Avalokan</span></span>
        </div>
        <span className="text-xs px-2 py-0.5 rounded bg-zinc-800 text-zinc-400 font-mono">v0.1.0</span>
      </div>

      <div className="flex items-center space-x-2">
        {currentTab && (
          <button
            onClick={onExecuteQuery}
            disabled={currentTab.isExecuting}
            className="flex items-center space-x-1.5 px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white rounded text-xs font-medium transition-colors shadow-sm"
          >
            {currentTab.isExecuting ? (
              <RefreshCw className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Play className="w-3.5 h-3.5 fill-current" />
            )}
            <span>{currentTab.isExecuting ? 'Running...' : 'Execute Query'}</span>
          </button>
        )}

        <button
          onClick={onNewProject}
          className="flex items-center space-x-1 px-2.5 py-1.5 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 rounded text-xs transition-colors"
        >
          <FolderKanban className="w-3.5 h-3.5" />
          <span>New Project</span>
        </button>

        <button
          onClick={onNewInstance}
          className="flex items-center space-x-1 px-2.5 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded text-xs transition-colors font-medium"
        >
          <Plus className="w-3.5 h-3.5" />
          <span>Add Instance</span>
        </button>
      </div>
    </header>
  );
};
