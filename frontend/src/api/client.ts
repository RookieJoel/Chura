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
    const message =
      typeof data === "object" &&
      data !== null &&
      "error" in data &&
      typeof data.error === "string"
        ? data.error
        : `Backend request failed (${response.status})`;
    throw new Error(`${message} (${response.status})`);
  }
  return data as T;
}

export async function apiRequest<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  return sameOriginRequest<T>(`/api/backend${path}`, init);
}
