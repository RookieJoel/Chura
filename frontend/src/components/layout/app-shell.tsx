import type { ReactNode } from "react";
import { TopNav } from "@/components/layout/top-nav";
import { Sidebar } from "@/components/layout/sidebar";

export function AppShell({
  children,
  minimal = false,
}: {
  children: ReactNode;
  minimal?: boolean;
}) {
  return (
    <div
      className={`flex ${
        minimal ? "h-screen" : "min-h-screen"
      } flex-1 flex-col overflow-hidden bg-background`}
    >
      <TopNav minimal={minimal} />
      <div className="flex flex-1 overflow-hidden">
        {!minimal && <Sidebar />}
        <main
          className={`flex min-w-0 flex-1 flex-col ${
            minimal ? "min-h-0 overflow-y-auto" : "overflow-hidden"
          }`}
        >
          {children}
        </main>
      </div>
    </div>
  );
}
