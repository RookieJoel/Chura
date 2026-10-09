import { getToken } from "next-auth/jwt";
import { NextResponse } from "next/server";
import { sessionCookieName } from "../../../../../auth";

export async function GET(request: Request) {
  const token = await getToken({
    req: request,
    secret: process.env.AUTH_SECRET,
    cookieName: sessionCookieName,
    secureCookie: process.env.NODE_ENV === "production",
  });
  const cookieNames: string[] = [];
  for (const cookie of request.headers.get("cookie")?.split(";") ?? []) {
    const name = cookie.split("=")[0]?.trim();
    if (name === sessionCookieName || name?.startsWith(`${sessionCookieName}.`)) {
      cookieNames.push(name);
    }
  }

  const issuer =
    process.env.KEYCLOAK_ISSUER ?? "http://localhost:8080/realms/chura";
  const clientId = process.env.KEYCLOAK_CLIENT_ID ?? "chura-auth-client";
  const logoutUrl = new URL(`${issuer}/protocol/openid-connect/logout`);
  logoutUrl.searchParams.set("client_id", clientId);
  logoutUrl.searchParams.set(
    "post_logout_redirect_uri",
    new URL("/", request.url).toString(),
  );

  if (typeof token?.idToken === "string") {
    logoutUrl.searchParams.set("id_token_hint", token.idToken);
  }

  const response = NextResponse.redirect(logoutUrl);
  for (const name of cookieNames) {
    response.cookies.set(name, "", {
      expires: new Date(0),
      maxAge: 0,
      httpOnly: true,
      path: "/",
      sameSite: "lax",
      secure: process.env.NODE_ENV === "production",
    });
  }
  return response;
}
