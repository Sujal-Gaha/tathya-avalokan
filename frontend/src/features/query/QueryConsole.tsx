import React from 'react';
import { X } from 'lucide-react';
import { useUIStore } from '../../store/useUIStore';
import { DataGrid } from './DataGrid';

interface QueryConsoleProps {
  onExecute: () => void;
}

export const QueryConsole: React.FC<QueryConsoleProps> = ({ onExecute }) => {
  const { tabs, activeTabId, setActiveTab, closeTab, updateTabSql } = useUIStore();
  const activeTab = tabs.find((t) => t.id === activeTabId);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault();
      onExecute();
    }
  };

  if (tabs.length === 0 || !activeTab) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center text-zinc-500 p-8">
        <div className="text-center space-y-2">
          <p className="text-lg text-zinc-300 font-medium">No Database Query Session Active</p>
          <p className="text-xs max-w-md">
            Select a database instance from the workspace sidebar or add a new instance to open a query editor.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col h-full bg-[#09090b] overflow-hidden">
      {/* Editor Tabs */}
      <div className="h-9 bg-[#121215] border-b border-zinc-800 flex items-center px-2 space-x-1 overflow-x-auto select-none">
        {tabs.map((tab) => {
          const isActive = tab.id === activeTabId;
          return (
            <div
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center space-x-2 px-3 py-1 text-xs rounded-t border-t-2 cursor-pointer transition-colors ${
                isActive
                  ? 'bg-[#09090b] text-zinc-100 border-blue-500'
                  : 'text-zinc-400 border-transparent hover:bg-zinc-800/40 hover:text-zinc-200'
              }`}
            >
              <span className="truncate max-w-[140px]">{tab.title}</span>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  closeTab(tab.id);
                }}
                className="hover:text-zinc-100 text-zinc-500 rounded p-0.5"
              >
                <X className="w-3 h-3" />
              </button>
            </div>
          );
        })}
      </div>

      {/* SQL Input Area */}
      <div className="h-44 border-b border-zinc-800 relative bg-[#09090b]">
        <textarea
          value={activeTab.sql}
          onChange={(e) => updateTabSql(activeTab.id, e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Write SQL query (e.g. SELECT * FROM users;)..."
          className="w-full h-full bg-transparent p-3 font-mono text-xs text-zinc-200 focus:outline-none resize-none selection:bg-blue-600"
          spellCheck={false}
        />
        <div className="absolute right-3 bottom-3 text-[10px] text-zinc-500 font-mono">
          Press Execute or Ctrl+Enter
        </div>
      </div>

      {/* Results Grid Area */}
      <div className="flex-1 overflow-hidden flex flex-col bg-[#121215]">
        {activeTab.error ? (
          <div className="p-4 text-xs text-rose-400 font-mono bg-rose-950/20 border-b border-rose-900/30">
            <strong>Error:</strong> {activeTab.error}
          </div>
        ) : activeTab.result ? (
          <DataGrid
            columns={activeTab.result.columns}
            rows={activeTab.result.rows}
            rowsAffected={activeTab.result.rows_affected}
            executionTimeMs={activeTab.result.execution_time_ms}
          />
        ) : (
          <div className="flex-1 flex items-center justify-center text-xs text-zinc-600">
            Execute a query to view tabular results.
          </div>
        )}
      </div>
    </div>
  );
};
