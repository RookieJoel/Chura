import type { ReactNode } from "react";
import { TopNav } from "@/components/layout/top-nav";
import { Sidebar } from "@/components/layout/sidebar";
import { ProjectProvider } from "@/components/layout/project-provider";

export function AppShell({ children }: { children: ReactNode }) {
  return (
    <ProjectProvider>
      <div className="flex min-h-screen flex-1 flex-col overflow-hidden bg-background">
        <TopNav />
        <div className="flex flex-1 overflow-hidden">
          <Sidebar />
          <main className="flex min-w-0 flex-1 flex-col overflow-hidden">
            {children}
          </main>
        </div>
      </div>
    </ProjectProvider>
  );
}
