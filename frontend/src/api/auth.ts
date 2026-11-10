import { sameOriginRequest } from "@/api/client";
import type { WhoAmI } from "@/types/auth";

export function getCurrentIdentity(): Promise<WhoAmI> {
  return sameOriginRequest<WhoAmI>("/api/whoami");
}
