This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

## Run the project

Start the services in this order from the repository root:

### 1. Start Keycloak

Keycloak provides authentication for the frontend and backend. Start it
first so that the `chura-keycloak` Docker network exists before the Agile
Execution service starts.

```bash
docker compose -f deploy/docker-compose.identity.yml up -d
```

Wait until the Keycloak container is healthy at
[http://localhost:8080](http://localhost:8080).

The imported realm is `chura`. The default development users are:

```text
Username: jojo
Password: test1234

Username: tonnam
Password: test1234
```

### 2. Start the Agile Execution services

This starts the backend API, PostgreSQL, MongoDB, migrations, and the
WorkItem WebSocket endpoint.

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

The backend API is available at
[http://localhost:8083](http://localhost:8083), and the WorkItem WebSocket is
available at `ws://localhost:8083/ws/work-items`.

### 3. Start the frontend

Install frontend dependencies if needed, then start Next.js:

```bash
cd frontend
pnpm install
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000). Opening a protected page
automatically starts the Keycloak login flow. After signing in, the frontend
verifies the user through the backend `/api/v1/whoami` endpoint.

To stop the services:

```bash
docker compose -f deploy/docker-compose.yml down
docker compose -f deploy/docker-compose.identity.yml down
```

## Frontend development

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

## Authentication

The frontend uses the repository's Keycloak realm (`chura`) through
Auth.js/NextAuth. Set these variables before starting the app:

```bash
AUTH_SECRET=replace-with-a-long-random-value
KEYCLOAK_ISSUER=http://localhost:8080/realms/chura
KEYCLOAK_CLIENT_ID=chura-auth-client
KEYCLOAK_CLIENT_SECRET=
CHURA_API_URL=http://localhost:8083
NEXT_PUBLIC_WORK_ITEM_RPC_URL=ws://localhost:8083/ws/work-items
```

`KEYCLOAK_CLIENT_SECRET` remains empty for the imported public Keycloak client.
Every frontend page requires a Keycloak session. After login, the server
forwards the session's access token to `/api/v1/whoami`; the verified ID,
email, and roles are available from the profile menu in the top navigation.

Unauthenticated requests are sent to the `/login` server page. That page calls
`signIn("keycloak", { redirectTo })` directly, so Auth.js starts the Keycloak
authorization flow without showing its provider-selection page. The imported
`chura` realm and its client use the `flowline` Keycloak login theme.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
