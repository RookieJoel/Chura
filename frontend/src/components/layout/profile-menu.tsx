"use client";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { ChevronDownIcon } from "@/components/icons";
import type { WhoAmI } from "@/types/auth";

export function ProfileMenu() {
  const pathname = usePathname();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [identity, setIdentity] = useState<WhoAmI>();
  const [error, setError] = useState(false);

  useEffect(() => {
    let active = true;
    fetch("/api/whoami", { cache: "no-store" })
      .then(async (response) => {
        if (response.status === 401) {
          const callbackUrl = `${pathname}${window.location.search}`;
          router.replace(`/login?callbackUrl=${encodeURIComponent(callbackUrl)}`);
          throw new Error("Session expired");
        }
        if (!response.ok) {
          throw new Error(`Identity verification failed (${response.status})`);
        }
        return (await response.json()) as WhoAmI;
      })
      .then((value) => {
        if (active) {
          setIdentity(value);
          setError(false);
        }
      })
      .catch(() => {
        if (active) {
          setIdentity(undefined);
          setError(true);
        }
      });

    return () => {
      active = false;
    };
  }, [pathname, router]);

  if (!identity) {
    return (
      <span className="text-[12px] font-semibold text-text-tertiary">
        {error ? "Account unavailable" : "Loading account…"}
      </span>
    );
  }

  return (
    <div className="relative">
      <button
        type="button"
        aria-expanded={open}
        aria-haspopup="menu"
        className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-left hover:bg-background"
        onClick={() => setOpen((current) => !current)}
      >
        <span className="flex size-7 items-center justify-center rounded-full bg-brand text-xs font-bold text-white">
          {identity.email.charAt(0).toUpperCase()}
        </span>
        <span className="hidden max-w-[150px] truncate text-[12px] font-semibold text-text sm:block">
          {identity.email}
        </span>
        <ChevronDownIcon className="size-3.5 text-text-tertiary" />
      </button>
      {open && (
        <div
          className="absolute right-0 top-full z-10 mt-2 w-64 rounded-xl border border-border bg-surface p-3 shadow-lg"
          role="menu"
        >
          <div className="border-b border-border pb-3">
            <p className="truncate text-sm font-bold text-text">{identity.email}</p>
            <p className="mt-1 break-all text-[11px] text-text-tertiary">
              ID: {identity.user_id}
            </p>
            <div className="mt-2 flex flex-wrap gap-1.5">
              {identity.roles.length > 0 ? (
                identity.roles.map((role) => (
                  <span
                    key={role}
                    className="rounded-full bg-brand-tint px-2 py-1 text-[11px] font-bold text-brand-dark"
                  >
                    {role}
                  </span>
                ))
              ) : (
                <span className="text-[11px] text-text-tertiary">No roles</span>
              )}
            </div>
          </div>
          <button
            type="button"
            role="menuitem"
            className="mt-3 w-full rounded-lg px-2 py-2 text-left text-[12px] font-bold text-text-secondary hover:bg-background"
            onClick={() => {
              window.location.href = new URL(
                "/api/auth/logout",
                window.location.origin,
              ).toString();
            }}
          >
            Log out
          </button>
        </div>
      )}
    </div>
  );
}
