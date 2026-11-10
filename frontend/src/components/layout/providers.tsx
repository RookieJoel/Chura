"use client";

import type { ReactNode } from "react";
import { ProjectProvider } from "@/components/layout/project-provider";

/** Root client providers so pages can call useProjects outside AppShell. */
export function Providers({ children }: { children: ReactNode }) {
  return <ProjectProvider>{children}</ProjectProvider>;
}
