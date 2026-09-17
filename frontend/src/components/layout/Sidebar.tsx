import React from 'react';
import { Folder, ChevronRight, HardDrive } from 'lucide-react';
import { Project } from '../../types/project';
import { useUIStore } from '../../store/useUIStore';

interface SidebarProps {
  projects: Project[];
  isLoading: boolean;
  onSelectInstance: (instanceId: string, instanceName: string) => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ projects, isLoading, onSelectInstance }) => {
  const { activeProjectId, activeInstanceId, setActiveProject } = useUIStore();

  return (
    <aside className="w-72 bg-[#18181b] border-r border-zinc-800 flex flex-col h-full select-none text-xs">
      <div className="p-3 border-b border-zinc-800/80 font-medium text-zinc-400 uppercase tracking-wider flex items-center justify-between">
        <span>Workspace Projects</span>
        <span className="text-[10px] bg-zinc-800 px-1.5 py-0.5 rounded text-zinc-400">
          {projects.length}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto p-2 space-y-1">
        {isLoading ? (
          <div className="text-zinc-500 p-4 text-center">Loading projects...</div>
        ) : projects.length === 0 ? (
          <div className="text-zinc-500 p-4 text-center">
            No projects yet.<br />Create a project to get started.
          </div>
        ) : (
          projects.map((project) => {
            const isSelected = activeProjectId === project.id;
            return (
              <div key={project.id} className="rounded overflow-hidden">
                <button
                  onClick={() => setActiveProject(isSelected ? null : project.id)}
                  className={`w-full flex items-center justify-between px-2.5 py-1.5 rounded transition-colors text-left ${
                    isSelected ? 'bg-zinc-800 text-zinc-100 font-medium' : 'text-zinc-300 hover:bg-zinc-800/50'
                  }`}
                >
                  <div className="flex items-center space-x-2 truncate">
                    <ChevronRight className={`w-3.5 h-3.5 transition-transform ${isSelected ? 'rotate-90 text-blue-400' : 'text-zinc-500'}`} />
                    <Folder className="w-3.5 h-3.5 text-blue-400 shrink-0" />
                    <span className="truncate">{project.name}</span>
                  </div>
                  <span className="text-[10px] text-zinc-500 font-mono">
                    {project.instances_count || project.instances?.length || 0}
                  </span>
                </button>

                {isSelected && (
                  <div className="ml-4 pl-2 border-l border-zinc-800 my-1 space-y-0.5">
                    {project.instances && project.instances.length > 0 ? (
                      project.instances.map((instance) => {
                        const isInstActive = activeInstanceId === instance.id;
                        return (
                          <button
                            key={instance.id}
                            onClick={() => onSelectInstance(instance.id, instance.name)}
                            className={`w-full flex items-center space-x-2 px-2 py-1 rounded text-left transition-colors ${
                              isInstActive
                                ? 'bg-blue-600/20 text-blue-300 border border-blue-500/30'
                                : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40'
                            }`}
                          >
                            <HardDrive className="w-3.5 h-3.5 text-zinc-400 shrink-0" />
                            <span className="truncate">{instance.name}</span>
                            <span className="text-[10px] px-1 rounded bg-zinc-800 text-zinc-500 ml-auto font-mono">
                              {instance.driver_type}
                            </span>
                          </button>
                        );
                      })
                    ) : (
                      <div className="text-[11px] text-zinc-600 py-1 px-2 italic">
                        No database instances
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </aside>
  );
};
