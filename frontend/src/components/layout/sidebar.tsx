"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { members } from "@/lib/mock-data";
import {
  CalendarIcon,
  ChevronDownIcon,
  ColumnsIcon,
  ListIcon,
} from "@/components/icons";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { cn } from "@/lib/cn";

const NAV_ITEMS = [
  { href: "/board", label: "Board", icon: ColumnsIcon },
  { href: "/backlog", label: "Backlog", icon: ListIcon },
  { href: "/sprints", label: "Sprints", icon: CalendarIcon },
  // { href: "/reports", label: "Reports", icon: BarChartIcon },
  // { href: "/settings", label: "Settings", icon: SettingsIcon },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <nav className="flex w-[232px] shrink-0 flex-col border-r border-border bg-surface px-3 py-4">
      <div className="mb-4 rounded-lg border border-border bg-background p-3">
        <div className="text-[10px] font-bold uppercase tracking-wider text-text-tertiary">
          Project
        </div>
        <div className="mt-1 flex items-center justify-between text-[13px] font-bold text-text">
          Chura Demo App
          <ChevronDownIcon className="size-3.5 text-text-tertiary" />
        </div>
      </div>
      {NAV_ITEMS.map(({ href, label, icon: ItemIcon }) => {
        const active = pathname.startsWith(href);
        return (
          <Link
            key={href}
            href={href}
            className={cn(
              "flex items-center gap-2.5 rounded-md px-3 py-2 text-[13px] font-semibold transition-colors",
              active ? "bg-brand-tint text-brand-dark" : "text-text-secondary hover:bg-background hover:text-text",
            )}
          >
            <ItemIcon className="size-3.5" />
            {label}
          </Link>
        );
      })}
      <div className="mx-1.5 my-5 h-px bg-border" />
      <div className="px-2.5 py-1.5 text-[11px] font-bold uppercase tracking-wide text-text-tertiary">
        Team members
      </div>
      {members.map((member) => (
        <div
          key={member.id}
          className="flex items-center gap-2 px-2.5 py-1.5 text-[12.5px] text-text-secondary"
        >
          <WorkItemAvatar member={member} size="sm" />
          {member.name.split(" ")[0]}
        </div>
      ))}
    </nav>
  );
}
