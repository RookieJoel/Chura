"use client";

import { useEffect } from "react";
import { signIn } from "next-auth/react";

export function LoginRedirect({ callbackUrl }: { callbackUrl: string }) {
  useEffect(() => {
    void signIn("keycloak", { callbackUrl });
  }, [callbackUrl]);

  return (
    <main className="fl-wash flex min-h-screen flex-col items-center justify-center gap-3 px-6 text-center">
      <span className="fl-accent-bar fl-accent-bar-lg" aria-hidden />
      <p className="m-0 text-2xl font-light tracking-tight text-text">Chura</p>
      <p className="m-0 text-sm text-text-secondary">
        Redirecting to Keycloak…
      </p>
    </main>
  );
}
