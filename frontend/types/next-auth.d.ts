import "next-auth/jwt";
import "next-auth";

declare module "next-auth/jwt" {
  interface JWT {
    accessToken?: string;
    idToken?: string;
  }
}

declare module "next-auth" {
  interface Session {
    accessToken?: string;
    idToken?: string;
  }
}
