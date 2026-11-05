"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { listProjects, type ProjectSummary } from "@/api/projects";

const selectedProjectStorageKey = "chura.selectedProjectId";
const configuredDefaultProjectId =
  process.env.NEXT_PUBLIC_CHURA_PROJECT_ID ?? "";

type ProjectLoadState = "loading" | "error" | "ready";

interface ProjectContextValue {
  projects: ProjectSummary[];
  selectedProjectId: string | null;
  selectProject: (projectId: string) => void;
  loadState: ProjectLoadState;
  error: string | null;
  refresh: () => void;
}

const ProjectContext = createContext<ProjectContextValue | null>(null);

function readStoredProjectId(): string | null {
  try {
    return window.localStorage.getItem(selectedProjectStorageKey);
  } catch {
    return null;
  }
}

function persistSelectedProjectId(projectId: string | null) {
  try {
    if (projectId) {
      window.localStorage.setItem(selectedProjectStorageKey, projectId);
    } else {
      window.localStorage.removeItem(selectedProjectStorageKey);
    }
  } catch {
    // The in-memory selection remains usable when browser storage is unavailable.
  }
}

export function ProjectProvider({ children }: { children: ReactNode }) {
  const [projects, setProjects] = useState<ProjectSummary[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(
    null,
  );
  const [loadState, setLoadState] = useState<ProjectLoadState>("loading");
  const [error, setError] = useState<string | null>(null);
  const [requestNumber, setRequestNumber] = useState(0);

  useEffect(() => {
    let active = true;
    listProjects()
      .then((accessibleProjects) => {
        if (!active) return;
        setProjects(accessibleProjects);
        const storedProjectId = readStoredProjectId();
        const isAccessible = (id: string | null) =>
          id !== null && accessibleProjects.some((project) => project.id === id);
        const nextProjectId = isAccessible(storedProjectId)
          ? storedProjectId
          : isAccessible(configuredDefaultProjectId)
            ? configuredDefaultProjectId
            : (accessibleProjects[0]?.id ?? null);
        setSelectedProjectId(nextProjectId);
        persistSelectedProjectId(nextProjectId);
        setLoadState("ready");
      })
      .catch((reason: unknown) => {
        if (!active) return;
        setProjects([]);
        setSelectedProjectId(null);
        setError(
          reason instanceof Error ? reason.message : "Project discovery failed",
        );
        setLoadState("error");
      });
    return () => {
      active = false;
    };
  }, [requestNumber]);

  const selectProject = useCallback(
    (projectId: string) => {
      if (!projects.some((project) => project.id === projectId)) return;
      setSelectedProjectId(projectId);
      persistSelectedProjectId(projectId);
    },
    [projects],
  );
  const refresh = useCallback(() => {
    setLoadState("loading");
    setError(null);
    setRequestNumber((current) => current + 1);
  }, []);
  const value = useMemo(
    () => ({
      projects,
      selectedProjectId,
      selectProject,
      loadState,
      error,
      refresh,
    }),
    [projects, selectedProjectId, selectProject, loadState, error, refresh],
  );

  return <ProjectContext.Provider value={value}>{children}</ProjectContext.Provider>;
}

export function useProjects() {
  const context = useContext(ProjectContext);
  if (!context) {
    throw new Error("useProjects must be used inside ProjectProvider");
  }
  return context;
}
