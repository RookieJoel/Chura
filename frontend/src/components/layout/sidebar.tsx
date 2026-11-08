"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  BarChartIcon,
  CalendarIcon,
  ColumnsIcon,
  ListIcon,
  MessageSquareIcon,
  SettingsIcon,
} from "@/components/icons";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { cn } from "@/lib/cn";
import { useProjects } from "@/components/layout/project-provider";

const NAV_ITEMS = [
  { href: "/board", label: "Board", icon: ColumnsIcon },
  { href: "/backlog", label: "Backlog", icon: ListIcon },
  { href: "/sprints", label: "Sprints", icon: CalendarIcon },
  { href: "/reports", label: "Reports", icon: BarChartIcon },
];

export function Sidebar() {
  const pathname = usePathname();
  const {
    projects,
    selectedProjectId,
    selectedProject,
    selectedProjectLoadState,
    selectedProjectError,
  } = useProjects();
  const project = projects.find((item) => item.id === selectedProjectId);
  const settingsHref = selectedProjectId
    ? `/projects/${encodeURIComponent(selectedProjectId)}`
    : null;
  const feedbackHref = selectedProjectId
    ? `${settingsHref}/feedback`
    : null;
  const settingsActive = !!settingsHref && pathname === settingsHref;
  const feedbackActive = !!feedbackHref && pathname === feedbackHref;

  return (
    <nav className="flex w-[232px] shrink-0 flex-col border-r border-border bg-surface px-3 py-4">
      <div className="mb-4 border border-border bg-background p-3">
        <div className="text-[10px] font-semibold uppercase tracking-wider text-text-tertiary">
          Project
        </div>
        <div className="mt-1 truncate text-[13px] font-semibold text-text">
          {project?.name ?? "No Project selected"}
        </div>
      </div>
      {settingsHref && (
        <Link
          href={settingsHref}
          className={cn(
            "flex items-center gap-2.5 px-3 py-2 text-[13px] font-semibold transition-colors",
            settingsActive
              ? "bg-brand-tint text-brand-dark"
              : "text-text-secondary hover:bg-background hover:text-text",
          )}
        >
          <SettingsIcon className="size-3.5" />
          Project settings
        </Link>
      )}
      {feedbackHref && (
        <Link
          href={feedbackHref}
          className={cn(
            "flex items-center gap-2.5 px-3 py-2 text-[13px] font-semibold transition-colors",
            feedbackActive
              ? "bg-brand-tint text-brand-dark"
              : "text-text-secondary hover:bg-background hover:text-text",
          )}
        >
          <MessageSquareIcon className="size-3.5" />
          Project feedback
        </Link>
      )}
      {NAV_ITEMS.map(({ href, label, icon: ItemIcon }) => {
        const active = pathname.startsWith(href);
        return (
          <Link
            key={href}
            href={href}
            className={cn(
              "flex items-center gap-2.5 px-3 py-2 text-[13px] font-semibold transition-colors",
              active
                ? "bg-brand-tint text-brand-dark"
                : "text-text-secondary hover:bg-background hover:text-text",
            )}
          >
            <ItemIcon className="size-3.5" />
            {label}
          </Link>
        );
      })}
      <div className="mx-1.5 my-5 h-px bg-border" />
      <div className="px-2.5 py-1.5 text-[11px] font-semibold uppercase tracking-wide text-text-tertiary">
        Team members
      </div>
      {selectedProjectLoadState === "loading" && (
        <p role="status" className="px-2.5 py-1.5 text-xs text-text-tertiary">
          Loading members…
        </p>
      )}
      {selectedProjectLoadState === "error" && (
        <p role="alert" className="px-2.5 py-1.5 text-xs text-red-600">
          {selectedProjectError}
        </p>
      )}
      {selectedProjectLoadState === "ready" &&
        selectedProject?.id === selectedProjectId &&
        (selectedProject.members.length > 0 ? (
          selectedProject.members.map((member) => (
            <div
              key={member.id}
              className="flex items-center gap-2 px-2.5 py-1.5 text-[12.5px] text-text-secondary"
            >
              <WorkItemAvatar member={member} size="sm" />
              <span className="min-w-0 truncate">
                {member.name.split(" ")[0]}
              </span>
            </div>
          ))
        ) : (
          <p className="px-2.5 py-1.5 text-xs text-text-tertiary">
            No Project members
          </p>
        ))}
    </nav>
  );
}
