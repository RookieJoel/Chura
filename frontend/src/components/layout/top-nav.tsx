"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { ProfileMenu } from "@/components/layout/profile-menu";
import { useProjects } from "@/components/layout/project-provider";

function isProjectSettingsPath(pathname: string) {
  return /^\/projects\/(?!new(?:\/|$))[^/]+/.test(pathname);
}

export function TopNav({ minimal = false }: { minimal?: boolean }) {
  const pathname = usePathname();
  const {
    projects,
    selectedProjectId,
    selectedProject: projectDetails,
    selectedProjectLoadState,
    selectedProjectError,
    selectProject,
    loadState,
    error,
    refresh,
  } = useProjects();
  const selectedProject = projects.find(
    (project) => project.id === selectedProjectId,
  );
  const projectSwitchLocked = isProjectSettingsPath(pathname);
  const selectorDisabled =
    projectSwitchLocked || loadState !== "ready" || projects.length === 0;

  return (
    <header className="flex h-16 shrink-0 items-center justify-between border-b border-border bg-surface px-6">
      <div className="flex items-center gap-4">
        <Link href="/board" className="flex items-center gap-2.5">
          <div className="flex size-9 items-center justify-center bg-brand text-base font-bold text-white">
            C
          </div>
          <span className="text-base font-semibold tracking-tight text-text">
            Chura
          </span>
        </Link>
        <div className="h-6 w-px bg-border" />
        <label
          className="flex items-center gap-2 text-sm font-medium text-text-secondary"
          title={
            projectSwitchLocked
              ? "Leave this Project page to switch Projects"
              : undefined
          }
        >
          <div className="flex size-6 items-center justify-center bg-accent-tint text-[11px] font-bold text-accent">
            {selectedProject?.name.slice(0, 2).toUpperCase() ?? "P"}
          </div>
          <select
            aria-label="Project"
            className="max-w-56 truncate bg-transparent text-sm outline-none disabled:cursor-not-allowed disabled:opacity-60"
            value={selectedProjectId ?? ""}
            disabled={selectorDisabled}
            onChange={(event) => selectProject(event.target.value)}
          >
            {projects.length === 0 ? (
              <option value="">
                {loadState === "ready"
                  ? "No accessible Projects"
                  : "Select Project"}
              </option>
            ) : (
              projects.map((project) => (
                <option key={project.id} value={project.id}>
                  {project.name}
                </option>
              ))
            )}
          </select>
        </label>
        <Link
          href="/projects/new"
          className="border border-border px-3 py-2 text-sm font-semibold text-text-secondary hover:border-accent hover:text-accent"
        >
          Create Project
        </Link>
        {loadState === "loading" && (
          <span role="status" aria-label="Loading Projects" className="sr-only">
            Loading Projects…
          </span>
        )}
        {loadState === "error" && (
          <div
            role="alert"
            className="flex items-center gap-2 text-xs text-red-600"
          >
            <span>{error}</span>
            <button
              type="button"
              onClick={refresh}
              className="font-bold underline"
            >
              Retry
            </button>
          </div>
        )}
      </div>
      <div className="flex items-center gap-4">
        {!minimal && (
          <div
            aria-label="Project members"
            className="flex max-w-56 items-center gap-2 text-xs text-text-tertiary"
          >
            {selectedProjectLoadState === "loading" && (
              <span role="status">Loading Project members…</span>
            )}
            {selectedProjectLoadState === "error" && (
              <span role="alert" className="text-red-600">
                {selectedProjectError ?? "Unable to load Project members"}
              </span>
            )}
            {selectedProjectLoadState === "ready" &&
              (projectDetails?.id === selectedProjectId ? (
                projectDetails.members.length > 0 ? (
                  projectDetails.members.map((member, index, visibleMembers) => (
                    <WorkItemAvatar
                      key={member.id}
                      member={member}
                      size="lg"
                      className={`border-2 border-surface ${index < visibleMembers.length - 1 ? "-mr-2" : ""}`}
                    />
                  ))
                ) : (
                  <span>No Project members</span>
                )
              ) : (
                <span>Select a Project to view members</span>
              ))}
          </div>
        )}
        <ProfileMenu />
      </div>
    </header>
  );
}
