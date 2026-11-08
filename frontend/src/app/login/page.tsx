import { LoginRedirect } from "@/components/auth/login-redirect";

function safeCallbackUrl(value: string | undefined) {
  if (!value) {
    return "/board";
  }

  try {
    const callbackUrl = new URL(value, "http://localhost");
    if (callbackUrl.origin !== "http://localhost") {
      return "/board";
    }
    return `${callbackUrl.pathname}${callbackUrl.search}${callbackUrl.hash}`;
  } catch {
    return "/board";
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
