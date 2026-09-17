import { create } from 'zustand';
import { EditorTab, QueryResponse } from '../types/query';

interface UIState {
  activeProjectId: string | null;
  activeInstanceId: string | null;
  isSidebarCollapsed: boolean;
  tabs: EditorTab[];
  activeTabId: string | null;

  // Actions
  setActiveProject: (projectId: string | null) => void;
  setActiveInstance: (instanceId: string | null) => void;
  toggleSidebar: () => void;
  openTab: (instanceId: string, instanceName: string) => void;
  closeTab: (tabId: string) => void;
  setActiveTab: (tabId: string) => void;
  updateTabSql: (tabId: string, sql: string) => void;
  setTabResult: (tabId: string, result: QueryResponse | null, error?: string | null) => void;
  setTabExecuting: (tabId: string, isExecuting: boolean) => void;
}

export const useUIStore = create<UIState>((set) => ({
  activeProjectId: null,
  activeInstanceId: null,
  isSidebarCollapsed: false,
  tabs: [],
  activeTabId: null,

  setActiveProject: (projectId) => set({ activeProjectId: projectId }),
  setActiveInstance: (instanceId) => set({ activeInstanceId: instanceId }),
  toggleSidebar: () => set((state) => ({ isSidebarCollapsed: !state.isSidebarCollapsed })),

  openTab: (instanceId, instanceName) => {
    const newTabId = `tab-${Date.now()}`;
    const newTab: EditorTab = {
      id: newTabId,
      instanceId,
      title: `${instanceName} (Query)`,
      sql: `-- SQL Workspace for ${instanceName}\nSELECT 1 as connection_test;\n`,
      isExecuting: false,
      result: null,
    };
    set((state) => ({
      tabs: [...state.tabs, newTab],
      activeTabId: newTabId,
      activeInstanceId: instanceId,
    }));
  },

  closeTab: (tabId) => {
    set((state) => {
      const nextTabs = state.tabs.filter((t) => t.id !== tabId);
      const nextActiveId =
        state.activeTabId === tabId
          ? nextTabs.length > 0
            ? nextTabs[nextTabs.length - 1]!.id
            : null
          : state.activeTabId;
      return { tabs: nextTabs, activeTabId: nextActiveId };
    });
  },

  setActiveTab: (tabId) => set({ activeTabId: tabId }),

  updateTabSql: (tabId, sql) =>
    set((state) => ({
      tabs: state.tabs.map((t) => (t.id === tabId ? { ...t, sql } : t)),
    })),

  setTabResult: (tabId, result, error = null) =>
    set((state) => ({
      tabs: state.tabs.map((t) =>
        t.id === tabId ? { ...t, result, error, isExecuting: false } : t
      ),
    })),

  setTabExecuting: (tabId, isExecuting) =>
    set((state) => ({
      tabs: state.tabs.map((t) => (t.id === tabId ? { ...t, isExecuting } : t)),
    })),
}));
