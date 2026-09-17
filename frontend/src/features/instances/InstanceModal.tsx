import React, { useState } from 'react';
import { X } from 'lucide-react';
import { CreateInstanceInput, DriverType } from '../../types/instance';

interface InstanceModalProps {
  isOpen: boolean;
  activeProjectId: string | null;
  onClose: () => void;
  onCreate: (projectId: string, input: CreateInstanceInput) => Promise<void>;
}

export const InstanceModal: React.FC<InstanceModalProps> = ({
  isOpen,
  activeProjectId,
  onClose,
  onCreate,
}) => {
  const [name, setName] = useState('');
  const [driverType, setDriverType] = useState<DriverType>('postgresql');
  const [host, setHost] = useState('localhost');
  const [port, setPort] = useState(5432);
  const [databaseName, setDatabaseName] = useState('');
  const [username, setUsername] = useState('postgres');
  const [password, setPassword] = useState('');
  const [isReadOnly, setIsReadOnly] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeProjectId || !name.trim() || !databaseName.trim()) return;
    setIsSubmitting(true);
    try {
      await onCreate(activeProjectId, {
        name: name.trim(),
        driver_type: driverType,
        host: host.trim(),
        port: Number(port),
        database_name: databaseName.trim(),
        username: username.trim(),
        password: password,
        is_read_only: isReadOnly,
      });
      onClose();
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/70 flex items-center justify-center p-4">
      <div className="bg-[#18181b] border border-zinc-800 rounded-lg w-full max-w-lg overflow-hidden shadow-2xl">
        <div className="flex items-center justify-between px-4 py-3 border-b border-zinc-800">
          <h3 className="text-sm font-semibold text-zinc-100">Add Database Instance</h3>
          <button onClick={onClose} className="text-zinc-400 hover:text-zinc-200">
            <X className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 space-y-3 text-xs">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-zinc-300 mb-1 font-medium">Instance Name</label>
              <input
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Primary DB"
                className="w-full bg-[#09090b] border border-zinc-800 rounded px-2.5 py-1.5 text-zinc-100 focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-zinc-300 mb-1 font-medium">Database Driver</label>
              <select
                value={driverType}
                onChange={(e) => {
                  const driver = e.target.value as DriverType;
                  setDriverType(driver);
                  if (driver === 'postgresql') setPort(5432);
                  if (driver === 'mysql') setPort(3306);
                }}
                className="w-full bg-[#09090b] border border-zinc-800 rounded px-2.5 py-1.5 text-zinc-100 focus:outline-none focus:border-blue-500"
              >
                <option value="postgresql">PostgreSQL</option>
                <option value="mysql">MySQL</option>
                <option value="sqlite">SQLite</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div className="col-span-2">
              <label className="block text-zinc-300 mb-1 font-medium">Host</label>
              <input
                type="text"
                value={host}
                onChange={(e) => setHost(e.target.value)}
                placeholder="localhost or IP"
                className="w-full bg-[#09090b] border border-zinc-800 rounded px-2.5 py-1.5 text-zinc-100 focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-zinc-300 mb-1 font-medium">Port</label>
              <input
                type="number"
                value={port}
                onChange={(e) => setPort(Number(e.target.value))}
                className="w-full bg-[#09090b] border border-zinc-800 rounded px-2.5 py-1.5 text-zinc-100 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-zinc-300 mb-1 font-medium">Database Name</label>
              <input
                type="text"
                required
                value={databaseName}
                onChange={(e) => setDatabaseName(e.target.value)}
                placeholder="Database name"
                className="w-full bg-[#09090b] border border-zinc-800 rounded px-2.5 py-1.5 text-zinc-100 focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-zinc-300 mb-1 font-medium">Username</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="DB user"
                className="w-full bg-[#09090b] border border-zinc-800 rounded px-2.5 py-1.5 text-zinc-100 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div>
            <label className="block text-zinc-300 mb-1 font-medium">Password (Encrypted at Rest)</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••••••"
              className="w-full bg-[#09090b] border border-zinc-800 rounded px-2.5 py-1.5 text-zinc-100 focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>

          <div className="flex items-center space-x-2 pt-1">
            <input
              type="checkbox"
              id="readOnly"
              checked={isReadOnly}
              onChange={(e) => setIsReadOnly(e.target.checked)}
              className="rounded border-zinc-800 bg-[#09090b] text-blue-600 focus:ring-0"
            />
            <label htmlFor="readOnly" className="text-zinc-300">
              Read-Only Mode (Prohibit mutating DML/DDL queries)
            </label>
          </div>

          <div className="flex items-center justify-end space-x-2 pt-3">
            <button
              type="button"
              onClick={onClose}
              className="px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 rounded transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting || !name.trim() || !databaseName.trim()}
              className="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded font-medium transition-colors"
            >
              {isSubmitting ? 'Saving...' : 'Save Instance'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
