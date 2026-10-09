import { LoginRedirect } from "@/components/auth/login-redirect";

function safeCallbackUrl(value: string | undefined) {
  if (!value) {
    return "/";
  }

  try {
    const callbackUrl = new URL(value, "http://localhost");
    if (callbackUrl.origin !== "http://localhost") {
      return "/";
    }
    return `${callbackUrl.pathname}${callbackUrl.search}${callbackUrl.hash}`;
  } catch {
    return "/";
  }
}

export default async function LoginPage({
  searchParams,
}: PageProps<"/login">) {
  const params = await searchParams;
  const callbackUrl =
    typeof params.callbackUrl === "string" ? params.callbackUrl : undefined;
  return <LoginRedirect callbackUrl={safeCallbackUrl(callbackUrl)} />;
}
