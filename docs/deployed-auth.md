# Fixing the Vercel / Render sign-in configuration

The deployed frontend calls `https://automata-fxzp.onrender.com/users/me` after
Clerk signs the user in. It does not use Vercel's `/api` path.

On October 17, 2026, a request with the Vercel origin and an OPTIONS preflight
both returned no `Access-Control-Allow-Origin` header. The unauthenticated GET
returned `missing authorization header`, the response from the legacy JWT
middleware in this repository. Clerk middleware returns `missing or invalid
authorization header`. These indicate a CORS configuration problem and a backend
running legacy authentication (or older code).

## Render

Set these variables on the **existing Render API service**, then redeploy the
latest backend:

```dotenv
AUTH_REQUIRE_CLERK=true
CLERK_SECRET_KEY=<secret key from the same Clerk instance as the Vercel publishable key>
HTTP_ALLOWED_ORIGINS=https://automata-azure.vercel.app
CLERK_AUTHORIZED_PARTIES=https://automata-azure.vercel.app
```

Keep the service's existing database, JWT secret, and other settings. Never put
the Clerk secret into a `VITE_` variable. Apply all database migrations to the
database selected by Render's `DATABASE_URL` before restarting. The API checks
for migrations 14 and 15 at startup. `AUTH_REQUIRE_CLERK=true` makes a missing
Clerk secret a startup error rather than selecting legacy authentication.

Vercel should retain `VITE_API_BASE_URL=https://automata-fxzp.onrender.com` and
`VITE_CLERK_PUBLISHABLE_KEY` from that same Clerk instance. Redeploy Vercel after
changing a `VITE_` variable, because Vite embeds it at build time.

Verify the browser preflight:

```sh
curl -i -X OPTIONS https://automata-fxzp.onrender.com/users/me \
  -H 'Origin: https://automata-azure.vercel.app' \
  -H 'Access-Control-Request-Method: GET' \
  -H 'Access-Control-Request-Headers: authorization,content-type'
```

It must include `Access-Control-Allow-Origin: https://automata-azure.vercel.app`
and permit GET and both requested headers. After signing in, `/users/me` must
return 200. If it still fails, inspect Render logs and the browser response.
An existing legacy account with the same email requires an explicit account
migration; do not automatically merge accounts by email.

## Password minimum and sign-in choices

These are **Clerk instance settings**, not React props or the legacy API's
password validator (which already accepts 8 characters). Select the instance
matching Vercel's publishable key in the Clerk Dashboard.

1. Open **User & authentication → Password** and enable password authentication.
2. Open **Update password requirements** and set minimum length to **8**.
   Keep compromised-password protection enabled. Save the changes.
3. Under **Email**, keep email enabled as an identifier. Keep **email verification
   code** enabled for sign-in if you want both password and passwordless access.
   To require passwords for sign-in, disable email-code sign-in after existing
   users have added passwords. Keep sign-up email verification and password
   recovery enabled; they serve different purposes.
4. Test an existing email-only account as well as a new account. Existing users
   can sign in by code and add a password through the profile button's security
   settings. Enabling passwords does not create passwords for existing users.

Keeping both methods is useful during migration and for users who prefer codes.
Clerk's prebuilt SignIn component follows these settings and offers alternate
methods; the exact first screen depends on the account and instance settings.
The application displays Weak / Medium / Strong feedback while users type a
password. The meter is guidance; Clerk remains the authority on acceptance.

References: [Clerk sign-in options](https://clerk.com/docs/guides/configure/auth-strategies/sign-up-sign-in-options)
and [password rules](https://clerk.com/docs/guides/secure/password-protection-and-rules).
