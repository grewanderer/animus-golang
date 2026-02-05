'use client';

import type { ReactNode } from 'react';
import { createContext, useCallback, useContext, useEffect, useState } from 'react';

const PROJECT_STORAGE_KEY = 'animus.console.project';

type ProjectContextValue = {
  projectId: string;
  setProjectId: (value: string) => void;
};

const ProjectContext = createContext<ProjectContextValue | undefined>(undefined);

export function ProjectProvider({ children }: { children: ReactNode }) {
  const [projectId, setProjectIdState] = useState('');

  useEffect(() => {
    const stored = window.localStorage.getItem(PROJECT_STORAGE_KEY) ?? '';
    if (stored) {
      setProjectIdState(stored);
    }
  }, []);

  const setProjectId = useCallback((value: string) => {
    const trimmed = value.trim();
    setProjectIdState(trimmed);
    if (trimmed) {
      window.localStorage.setItem(PROJECT_STORAGE_KEY, trimmed);
    } else {
      window.localStorage.removeItem(PROJECT_STORAGE_KEY);
    }
  }, []);

  return <ProjectContext.Provider value={{ projectId, setProjectId }}>{children}</ProjectContext.Provider>;
}

export function useProjectContext() {
  const context = useContext(ProjectContext);
  if (!context) {
    throw new Error('ProjectContext is not configured');
  }
  return context;
}
