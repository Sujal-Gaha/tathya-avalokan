import React from 'react';
import { ColumnMeta } from '../../types/query';

interface DataGridProps {
  columns: ColumnMeta[];
  rows: Record<string, unknown>[];
  rowsAffected: number;
  executionTimeMs: number;
}

export const DataGrid: React.FC<DataGridProps> = ({ columns, rows }) => {
  if (columns.length === 0 && rows.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center text-xs text-zinc-500">
        Query completed with 0 rows returned.
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-auto">
      <table className="w-full text-left text-xs border-collapse">
        <thead className="bg-[#18181b] sticky top-0 border-b border-zinc-800 text-zinc-400 select-none z-10">
          <tr>
            <th className="px-3 py-2 text-[10px] font-mono text-zinc-500 w-12 border-r border-zinc-800">#</th>
            {columns.map((col) => (
              <th key={col.name} className="px-3 py-2 font-medium border-r border-zinc-800 last:border-r-0">
                <div className="flex items-center justify-between space-x-2">
                  <span className="text-zinc-200">{col.name}</span>
                  <span className="text-[10px] font-mono text-zinc-500 uppercase">{col.type}</span>
                </div>
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-zinc-800/60 font-mono text-[11px]">
          {rows.map((row, idx) => (
            <tr key={idx} className="hover:bg-zinc-800/40 transition-colors">
              <td className="px-3 py-1.5 text-zinc-600 border-r border-zinc-800/60 select-none">{idx + 1}</td>
              {columns.map((col) => {
                const val = row[col.name];
                const display = val === null || val === undefined ? (
                  <span className="text-zinc-600 italic">NULL</span>
                ) : typeof val === 'object' ? (
                  JSON.stringify(val)
                ) : (
                  String(val)
                );
                return (
                  <td key={col.name} className="px-3 py-1.5 text-zinc-300 border-r border-zinc-800/60 last:border-r-0 whitespace-nowrap">
                    {display}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
