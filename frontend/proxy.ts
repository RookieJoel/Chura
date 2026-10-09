import { auth } from "./auth";

const protectedProxy = auth((request) => {
  if (!request.auth) {
    const loginUrl = new URL("/login", request.nextUrl.origin);
    loginUrl.searchParams.set("callbackUrl", request.nextUrl.href);
    return Response.redirect(loginUrl);
  }
});

export default protectedProxy;

export const config = {
  matcher: [
    "/((?!login|api/auth|api/whoami|api/auth/logout|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico|css|js)$).*)",
  ],
};
