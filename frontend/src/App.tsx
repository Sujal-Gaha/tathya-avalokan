import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Header } from './components/layout/Header';
import { Sidebar } from './components/layout/Sidebar';
import { StatusBar } from './components/layout/StatusBar';
import { QueryConsole } from './features/query/QueryConsole';
import { ProjectModal } from './features/projects/ProjectModal';
import { InstanceModal } from './features/instances/InstanceModal';
import { ToastContainer } from './components/ui/Toast';
import { projectService } from './services/project-service';
import { instanceService } from './services/instance-service';
import { queryService } from './services/query-service';
import { useUIStore } from './store/useUIStore';
import { useToastStore } from './store/useToastStore';

export const App: React.FC = () => {
  const queryClient = useQueryClient();
  const [isProjectModalOpen, setIsProjectModalOpen] = useState(false);
  const [isInstanceModalOpen, setIsInstanceModalOpen] = useState(false);

  const {
    activeProjectId,
    activeTabId,
    tabs,
    openTab,
    setTabExecuting,
    setTabResult,
  } = useUIStore();

  const addToast = useToastStore((s) => s.addToast);

  // Fetch Projects List
  const { data: projects = [], isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectService.list(),
  });

  // Create Project Mutation
  const createProjectMutation = useMutation({
    mutationFn: (data: { name: string; description: string }) => projectService.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      addToast('Project created successfully', 'success');
    },
    onError: (err: Error) => {
      addToast(err.message || 'Failed to create project', 'error');
    },
  });

  // Create Instance Mutation
  const createInstanceMutation = useMutation({
    mutationFn: ({ projectId, input }: { projectId: string; input: any }) =>
      instanceService.create(projectId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      addToast('Database instance added successfully', 'success');
    },
    onError: (err: Error) => {
      addToast(err.message || 'Failed to create instance', 'error');
    },
  });

  // Query Execution Handler
  const handleExecuteQuery = async () => {
    const currentTab = tabs.find((t) => t.id === activeTabId);
    if (!currentTab || !currentTab.sql.trim()) return;
    if (!currentTab.instanceId) {
      addToast('No database instance selected. Open a connection from the sidebar first.', 'warning');
      return;
    }

    setTabExecuting(currentTab.id, true);
    try {
      const response = await queryService.executeQuery(currentTab.instanceId, {
        sql: currentTab.sql,
      });
      setTabResult(currentTab.id, response);
    } catch (err: any) {
      setTabResult(currentTab.id, null, err.message || 'Execution error');
      addToast(err.message || 'Query execution failed', 'error');
    }
  };

  // Guard: warn user if they try to add an instance with no active project selected
  const handleOpenInstanceModal = () => {
    if (!activeProjectId) {
      addToast('Select a project from the sidebar before adding an instance.', 'warning');
      return;
    }
    setIsInstanceModalOpen(true);
  };

  return (
    <div className="h-screen w-screen flex flex-col bg-[#09090b] text-zinc-100 overflow-hidden font-sans">
      <Header
        onNewProject={() => setIsProjectModalOpen(true)}
        onNewInstance={handleOpenInstanceModal}
        onExecuteQuery={handleExecuteQuery}
      />

      <div className="flex-1 flex overflow-hidden">
        <Sidebar
          projects={projects}
          isLoading={isLoading}
          onSelectInstance={(instanceId, instanceName) => openTab(instanceId, instanceName)}
        />
        <main className="flex-1 flex flex-col overflow-hidden">
          <QueryConsole onExecute={handleExecuteQuery} />
        </main>
      </div>

      <StatusBar />

      <ProjectModal
        isOpen={isProjectModalOpen}
        onClose={() => setIsProjectModalOpen(false)}
        onCreate={async (name, description) => {
          await createProjectMutation.mutateAsync({ name, description });
        }}
      />

      <InstanceModal
        isOpen={isInstanceModalOpen}
        activeProjectId={activeProjectId}
        onClose={() => setIsInstanceModalOpen(false)}
        onCreate={async (projectId, input) => {
          await createInstanceMutation.mutateAsync({ projectId, input });
        }}
      />

      {/* Global toast notification container */}
      <ToastContainer />
    </div>
  );
};
