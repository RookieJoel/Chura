import type { ReactNode } from "react";
import { TopNav } from "@/components/layout/top-nav";
import { Sidebar } from "@/components/layout/sidebar";

export function AppShell({ children }: { children: ReactNode }) {
  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      <TopNav />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar />
        <div className="flex flex-1 flex-col overflow-hidden">{children}</div>
      </div>
    </div>
  );
}
