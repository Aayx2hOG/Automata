# Automata demo

The screenshots and recording use the real React workspace and Go API with a
fresh account in a disposable PostgreSQL database. The browser harness replaces
the Clerk profile widget and supplies a server-issued legacy JWT; it does not
record Google sign-in. No API responses are mocked in this demo.

- [Workspace screenshot](workspace.png)
- [Editor and successful execution](editor.png)
- [Demo recording](demo.webm)
- [Measured recovery and latency results](../benchmarks/verification.json)

## Reproduce the browser demo

Start a **dedicated test backend** with all migrations applied, using its local JWT
mode (no Clerk secret). Keep your ordinary application's environment and database
separate. Start the API from a directory without `.env` or `.env.local` files if
necessary. Vite expects the API on port 8080.

From `frontend/`:

```sh
npm ci
npx playwright install --with-deps chromium
npm run demo
```

The capture creates a workflow, loads a deterministic branching graph through the
authenticated API, edits and saves it, executes it, inspects checkpoints, and shows
the updated home screen and history. Screenshots are written into `docs/demo/`;
Playwright writes the recording under `frontend/test-results/`. Copy that test's
`video.webm` to `docs/demo/demo.webm` after a successful recording. This is a short,
un-narrated automated walkthrough, not a recording of manually drawing the graph.

The unused branch has no checkpoint: an absent node is not evidence of a failure.
The demo depends only on the local backend and database, with no external HTTP
service required.


See [recovery and latency verification](../benchmarks/README.md) for the experiment and its measured results.
