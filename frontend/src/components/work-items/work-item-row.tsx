"use client"

import Link from "next/link";
import type { Member, WorkItem } from "@/types/work-item";
import { TypeBadge } from "@/components/work-items/type-badge";
import { PriorityIcon } from "@/components/work-items/priority-icon";
import { PointsBadge } from "@/components/work-items/points-badge";
import { EpicPill } from "@/components/work-items/epic-pill";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { cn } from "@/lib/cn";
import { PlusIcon } from "../icons";
import React from "react";

interface WorkItemRowProps {
  item: WorkItem;
  assignee?: Member;
  className?: string;
  state?: "pending" | "resolved" | "error";
  error?: string;
}

export function WorkItemRow({ item, assignee, className, state, error }: WorkItemRowProps) {
  if (!state || state === "resolved") {
    return (
      <Link
        href={`/work-items/${item.id}`}
        className={cn(
          "flex items-center gap-3 border-b border-border px-3.5 py-[9px] last:border-b-0 hover:bg-background",
          className,
        )}
      >
        {item.type ? <TypeBadge type={item.type} size="md" /> : null}
        <span className="w-[76px] shrink-0 overflow-hidden text-ellipsis whitespace-nowrap font-mono text-[11.5px] font-bold text-text-tertiary">
          {item.id}
        </span>
        <span className="flex-1 truncate text-[13px] font-medium text-text">
          {item.title}
        </span>
        <EpicPill epic={item.epic} />
        <PriorityIcon priority={item.priority} />
        <PointsBadge points={item.storyPoints} />
        <WorkItemAvatar member={assignee} size="md" />
      </Link>
    );
  }

  if (state === "pending") {
      return (
      <div
        className={cn(
          "flex items-center gap-3 border-b border-border px-3.5 py-[9px] last:border-b-0",
          className,
        )}
      >
        <span
          aria-hidden="true"
          className="h-3.5 w-6 shrink-0 animate-pulse rounded bg-border-strong"
        />
        <span
          aria-hidden="true"
          className="h-3 w-[76px] shrink-0 animate-pulse rounded bg-border-strong"
        />
        <span
          aria-hidden="true"
          className="h-3 w-full min-w-0 animate-pulse rounded bg-border-strong"
        />
        <span
          aria-hidden="true"
          className="h-4 w-16 shrink-0 animate-pulse rounded-full bg-border-strong"
        />
        <span
          aria-hidden="true"
          className="size-3.5 shrink-0 animate-pulse rounded bg-border-strong"
        />
        <span
          aria-hidden="true"
          className="size-5 shrink-0 animate-pulse rounded-md bg-border-strong"
        />
        <span
          aria-hidden="true"
          className="size-[22px] shrink-0 animate-pulse rounded-full bg-border-strong"
        />
      </div>
    );
  }

  return (
    <div
      className={cn(
        "flex items-center gap-3 border-b border-border px-3.5 py-[9px] last:border-b-0",
        className,
      )}
    >
      <span className="w-[76px] shrink-0 overflow-hidden text-ellipsis whitespace-nowrap font-mono text-[11.5px] font-bold text-text-tertiary">
        {item.id}
      </span>
      <span className="flex-1 truncate text-[13px] font-medium text-text">
        {item.title}
      </span>
      <span className="text-xs text-red-600">{error ?? "Unable to create"}</span>
    </div>
  );
}

export function WorkItemInsertRow({ className, onSubmit }: { className?: string, onSubmit?: (value: string) => void }) {

  const [value, setValue] = React.useState("");

  return (
    <div
      className={cn(
        "flex items-center gap-3 border-b border-border px-3.5 py-[9px] last:border-b-0 hover:bg-background",
        className,
      )}
    >
      <PlusIcon className="size-4 bg-primary text-brand"/>
      <input onKeyDown={(e) => {
        if (e.key === "Enter") {
          e.preventDefault();
          onSubmit?.(value);
          setValue("");
        }
      }} value={value} onChange={(e) => setValue(e.target.value)} className="flex-1 truncate text-[13px] font-medium text-text focus:outline-none focus:ring-0" placeholder="Add a new work item..." />
    </div>
  );
}