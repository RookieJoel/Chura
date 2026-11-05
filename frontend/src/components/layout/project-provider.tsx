"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
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
  registerCreatedProject: (project: ProjectSummary) => void;
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
  const createdProjects = useRef(new Map<string, ProjectSummary>());

  useEffect(() => {
    let active = true;
    listProjects()
      .then((accessibleProjects) => {
        if (!active) return;
        const mergedProjects = [
          ...createdProjects.current.values(),
          ...accessibleProjects.filter(
            (project) => !createdProjects.current.has(project.id),
          ),
        ];
        setProjects(mergedProjects);
        const storedProjectId = readStoredProjectId();
        const isAccessible = (id: string | null) =>
          id !== null && mergedProjects.some((project) => project.id === id);
        const nextProjectId = isAccessible(storedProjectId)
          ? storedProjectId
          : isAccessible(configuredDefaultProjectId)
            ? configuredDefaultProjectId
            : (mergedProjects[0]?.id ?? null);
        setSelectedProjectId(nextProjectId);
        persistSelectedProjectId(nextProjectId);
        setLoadState("ready");
      })
      .catch((reason: unknown) => {
        if (!active) return;
        const knownCreatedProjects = [...createdProjects.current.values()];
        setProjects(knownCreatedProjects);
        const storedProjectId = readStoredProjectId();
        const selectedCreatedProject = knownCreatedProjects.some(
          (project) => project.id === storedProjectId,
        )
          ? storedProjectId
          : null;
        setSelectedProjectId(selectedCreatedProject);
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
  const registerCreatedProject = useCallback((project: ProjectSummary) => {
    createdProjects.current.set(project.id, project);
    setProjects((current) => [
      project,
      ...current.filter((item) => item.id !== project.id),
    ]);
    setSelectedProjectId(project.id);
    persistSelectedProjectId(project.id);
    setLoadState("ready");
    setError(null);
  }, []);
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
      registerCreatedProject,
      loadState,
      error,
      refresh,
    }),
    [
      projects,
      selectedProjectId,
      selectProject,
      registerCreatedProject,
      loadState,
      error,
      refresh,
    ],
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
