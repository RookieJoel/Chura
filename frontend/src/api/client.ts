export const SESSION_EXPIRED_MESSAGE =
  "Your session has expired. Please sign in again to continue.";

export async function sameOriginRequest<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  let response: Response;
  try {
    response = await fetch(path, {
      ...init,
      headers: {
        "content-type": "application/json",
        ...init?.headers,
      },
      cache: "no-store",
    });
  } catch (error) {
    throw new Error(
      `Unable to reach the backend: ${
        error instanceof Error ? error.message : "network request failed"
      }`,
    );
  }

  const body = await response.text();
  let data: unknown;
  try {
    data = body ? JSON.parse(body) : undefined;
  } catch {
    const detail = body.trim().slice(0, 160);
    throw new Error(
      detail
        ? `Backend returned an invalid response (${response.status}): ${detail}`
        : `Backend returned an invalid response (${response.status})`,
    );
  }
  if (!response.ok) {
    // A 401 always means the sign-in session ran out (or was never there),
    // whichever service answered; say so instead of "Unauthenticated (401)".
    if (response.status === 401) {
      throw new Error(SESSION_EXPIRED_MESSAGE);
    }
    if (typeof data === "object" && data !== null) {
      if ("violations" in data && Array.isArray(data.violations) && data.violations.length > 0) {
        const messages = data.violations
          .map((violation) => {
            if (
              typeof violation === "object" &&
              violation !== null &&
              "message" in violation &&
              typeof violation.message === "string" &&
              violation.message.trim()
            ) {
              return violation.message.trim();
            }
            return "";
          })
          .filter(Boolean);
        if (messages.length > 0) {
          throw new Error(`${messages.join("; ")} (${response.status})`);
        }
      }
      if ("error" in data && typeof data.error === "string") {
        throw new Error(`${data.error} (${response.status})`);
      }
    }
    throw new Error(`Backend request failed (${response.status})`);
  }
  return data as T;
}

export async function apiRequest<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  return sameOriginRequest<T>(`/api/backend${path}`, init);
}
