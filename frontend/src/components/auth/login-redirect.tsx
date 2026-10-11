"use client";

import { useEffect } from "react";
import { signIn } from "next-auth/react";

export function LoginRedirect({ callbackUrl }: { callbackUrl: string }) {
  useEffect(() => {
    void signIn("keycloak", { callbackUrl });
  }, [callbackUrl]);

  return (
    <main className="flex min-h-screen items-center justify-center bg-background text-sm text-text-secondary">
      Redirecting to Keycloak…
    </main>
  );
}
