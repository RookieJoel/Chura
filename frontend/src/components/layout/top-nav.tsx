import Link from "next/link";
import { members } from "@/lib/mock-data";
import { BellIcon, ChevronDownIcon, SearchIcon } from "@/components/icons";
import { WorkItemAvatar } from "@/components/work-items/avatar";

export function TopNav() {
  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-border bg-surface px-5">
      <div className="flex items-center gap-3.5">
        <Link href="/board" className="flex items-center gap-2">
          <div className="flex size-[26px] items-center justify-center rounded-[7px] bg-brand text-sm font-extrabold text-white">
            C
          </div>
          <span className="text-[15px] font-bold tracking-tight text-text">
            Chura
          </span>
        </Link>
        <div className="h-5 w-px bg-border" />
        <div className="flex items-center gap-1.5 text-[13px] font-semibold text-text-secondary">
          <div className="flex size-[18px] items-center justify-center rounded-[5px] bg-task-tint text-[10px] font-extrabold text-task">
            CD
          </div>
          Chura Demo App
          <ChevronDownIcon className="size-3.5 text-text-tertiary" />
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="flex w-[220px] items-center gap-2 rounded-lg border border-border bg-background px-2.5 py-1.5 text-text-tertiary">
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
      </div>
    </header>
  );
}
