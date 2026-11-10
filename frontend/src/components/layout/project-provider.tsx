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
import { getCurrentIdentity } from "@/api/auth";
import {
  getProject,
  listProjects,
  type ProjectDetails,
  type ProjectSummary,
} from "@/api/projects";
import { getTemplate, type ProjectTemplate } from "@/api/templates";
import type { WhoAmI } from "@/types/auth";

const selectedProjectStorageKey = "chura.selectedProjectId";
const configuredDefaultProjectId =
  process.env.NEXT_PUBLIC_CHURA_PROJECT_ID ?? "";

type ProjectLoadState = "loading" | "error" | "ready";

interface SelectedProjectResult {
  projectId: string;
  project?: ProjectDetails;
  template?: ProjectTemplate;
  error?: string;
}

interface ProjectContextValue {
  projects: ProjectSummary[];
  selectedProjectId: string | null;
  selectedProject: ProjectDetails | null;
  selectedTemplate: ProjectTemplate | null;
  selectedProjectLoadState: ProjectLoadState;
  selectedProjectError: string | null;
  currentUser: WhoAmI | null;
  identityError: string | null;
  selectProject: (projectId: string) => void;
  registerCreatedProject: (project: ProjectSummary) => void;
  updateSelectedProject: (project: ProjectDetails) => void;
  removeDeletedProject: (projectId: string) => void;
  loadState: ProjectLoadState;
  error: string | null;
  refresh: () => void;
  refreshSelectedProject: () => void;
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

export function ProjectProvider({
  children,
  initialProjectId,
}: {
  children: ReactNode;
  initialProjectId?: string;
}) {
  const [projects, setProjects] = useState<ProjectSummary[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(
    null,
  );
  const [loadState, setLoadState] = useState<ProjectLoadState>("loading");
  const [error, setError] = useState<string | null>(null);
  const [requestNumber, setRequestNumber] = useState(0);
  const [selectedProjectRequest, setSelectedProjectRequest] = useState(0);
  const [selectedProjectResult, setSelectedProjectResult] =
    useState<SelectedProjectResult | null>(null);
  const [currentUser, setCurrentUser] = useState<WhoAmI | null>(null);
  const [identityError, setIdentityError] = useState<string | null>(null);
  const createdProjects = useRef(new Map<string, ProjectSummary>());
  const selectedProjectSummary = projects.find(
    (project) => project.id === selectedProjectId,
  );
  const selectedTemplateId = selectedProjectSummary?.template_id;

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
        const isAccessible = (id: string | null | undefined) =>
          id != null && mergedProjects.some((project) => project.id === id);
        const nextProjectId = isAccessible(initialProjectId)
          ? (initialProjectId ?? null)
          : isAccessible(storedProjectId)
            ? (storedProjectId ?? null)
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
          (project) => project.id === (initialProjectId ?? storedProjectId),
        )
          ? (initialProjectId ?? storedProjectId)
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
  }, [initialProjectId, requestNumber]);

  useEffect(() => {
    let active = true;
    getCurrentIdentity()
      .then((identity) => {
        if (active) setCurrentUser(identity);
      })
      .catch((reason: unknown) => {
        if (active) {
          setIdentityError(
            reason instanceof Error
              ? reason.message
              : "Unable to load current identity",
          );
        }
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    if (!selectedProjectId) return;
    if (!selectedTemplateId) return;
    let active = true;
    Promise.all([
      getProject(selectedProjectId),
      getTemplate(selectedTemplateId),
    ])
      .then(([project, template]) => {
        if (active) {
          setSelectedProjectResult({ projectId: selectedProjectId, project, template });
        }
      })
      .catch((reason: unknown) => {
        if (active) {
          setSelectedProjectResult({
            projectId: selectedProjectId,
            error:
              reason instanceof Error
                ? reason.message
                : "Unable to load selected Project",
          });
        }
      });
    return () => {
      active = false;
    };
  }, [selectedProjectId, selectedProjectRequest, selectedTemplateId]);

  const selectProject = useCallback(
    (projectId: string) => {
      if (!projects.some((project) => project.id === projectId)) return;
      setSelectedProjectResult(null);
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
  const updateSelectedProject = useCallback(
    (project: ProjectDetails) => {
      // Auditors are not Project members; keep their selection when overview refreshes.
      const currentUserRemains =
        !currentUser ||
        currentUser.roles.includes("auditor") ||
        project.members.some((member) => member.id === currentUser.user_id);
      if (!currentUserRemains) {
        const remainingProjects = projects.filter(
          (item) => item.id !== project.id,
        );
        setProjects(remainingProjects);
        if (selectedProjectId === project.id) {
          const nextProjectId = remainingProjects[0]?.id ?? null;
          setSelectedProjectId(nextProjectId);
          persistSelectedProjectId(nextProjectId);
        }
        setSelectedProjectResult(null);
        return;
      }

      setSelectedProjectResult((current) =>
        current?.projectId === project.id
          ? { ...current, project, error: undefined }
          : current,
      );
      setProjects((current) =>
        current.map((item) =>
          item.id === project.id
            ? {
                id: project.id,
                name: project.name,
                description: project.description,
                template_id: project.template_id,
                mode: project.mode,
              }
            : item,
        ),
      );
    },
    [currentUser, projects, selectedProjectId],
  );
  const removeDeletedProject = useCallback(
    (projectId: string) => {
      createdProjects.current.delete(projectId);
      const remainingProjects = projects.filter((item) => item.id !== projectId);
      setProjects(remainingProjects);
      setSelectedProjectResult(null);
      if (selectedProjectId === projectId) {
        const nextProjectId = remainingProjects[0]?.id ?? null;
        setSelectedProjectId(nextProjectId);
        persistSelectedProjectId(nextProjectId);
      }
    },
    [projects, selectedProjectId],
  );
  const refresh = useCallback(() => {
    setLoadState("loading");
    setError(null);
    setRequestNumber((current) => current + 1);
  }, []);
  const refreshSelectedProject = useCallback(() => {
    setSelectedProjectRequest((current) => current + 1);
  }, []);

  const currentResult =
    selectedProjectResult?.projectId === selectedProjectId
      ? selectedProjectResult
      : null;
  const selectedProject = currentResult?.project ?? null;
  const selectedTemplate = currentResult?.template ?? null;
  const selectedProjectLoadState: ProjectLoadState = !selectedProjectId
    ? "ready"
    : currentResult?.error || !selectedProjectSummary
      ? "error"
      : selectedProject && selectedTemplate
        ? "ready"
        : "loading";
  const selectedProjectError =
    currentResult?.error ??
    (selectedProjectId && !selectedProjectSummary
      ? "Selected Project is no longer available"
      : null);

  const value = useMemo(
    () => ({
      projects,
      selectedProjectId,
      selectedProject,
      selectedTemplate,
      selectedProjectLoadState,
      selectedProjectError,
      currentUser,
      identityError,
      selectProject,
      registerCreatedProject,
      updateSelectedProject,
      removeDeletedProject,
      loadState,
      error,
      refresh,
      refreshSelectedProject,
    }),
    [
      projects,
      selectedProjectId,
      selectedProject,
      selectedTemplate,
      selectedProjectLoadState,
      selectedProjectError,
      currentUser,
      identityError,
      selectProject,
      registerCreatedProject,
      updateSelectedProject,
      removeDeletedProject,
      loadState,
      error,
      refresh,
      refreshSelectedProject,
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
