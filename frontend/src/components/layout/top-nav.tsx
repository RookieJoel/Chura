"use client";

import Link from "next/link";
import { members } from "@/lib/mock-data";
import { BellIcon, SearchIcon } from "@/components/icons";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { ProfileMenu } from "@/components/layout/profile-menu";
import { useProjects } from "@/components/layout/project-provider";

export function TopNav() {
  const {
    projects,
    selectedProjectId,
    selectProject,
    loadState,
    error,
    refresh,
  } = useProjects();
  const selectedProject = projects.find(
    (project) => project.id === selectedProjectId,
  );

  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-border bg-surface px-5">
      <div className="flex items-center gap-3.5">
        <Link href="/board" className="flex items-center gap-2">
          <div className="flex size-[28px] items-center justify-center rounded-md bg-brand text-sm font-extrabold text-white shadow-sm">
            C
          </div>
          <span className="text-[15px] font-bold tracking-tight text-text">
            Chura
          </span>
        </Link>
        <div className="h-5 w-px bg-border" />
        <label className="flex items-center gap-1.5 text-[13px] font-semibold text-text-secondary">
          <div className="flex size-[18px] items-center justify-center rounded-[5px] bg-task-tint text-[10px] font-extrabold text-task">
            {selectedProject?.name.slice(0, 2).toUpperCase() ?? "P"}
          </div>
          <select
            aria-label="Project"
            className="max-w-48 truncate bg-transparent outline-none"
            value={selectedProjectId ?? ""}
            disabled={loadState !== "ready" || projects.length === 0}
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
        <div className="flex w-[240px] items-center gap-2 rounded-md border border-border bg-background px-3 py-2 text-text-tertiary transition-colors focus-within:border-brand">
          <SearchIcon className="size-3.5" />
          <span className="text-[12.5px]">Search work items</span>
        </div>
        <div className="flex items-center">
          {members.map((member, index) => (
            <WorkItemAvatar
              key={member.id}
              member={member}
              size="lg"
              className={`border-2 border-surface ${index < members.length - 1 ? "-mr-2" : ""}`}
            />
          ))}
        </div>
        <BellIcon className="size-4 text-text-secondary" />
        <ProfileMenu />
      </div>
    </header>
  );
}
