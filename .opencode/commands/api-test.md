---
description: Test a Go API route with curl, auto-handle auth tokens via token.md
agent: build
model: opencode/hy3-free
---

User will give a route/handler line (e.g. `r.POST("", userHandler.CreateUser)`).

Steps:

1. Find the handler via grep (don't read whole files) — get full path, method, and body fields.
2. If protected, read the matching role's token from `./token.md` and set `Authorization: Bearer <token>`.
3. Build the curl command, show it, then run it (`-w "\nStatus: %{http_code}\n"`).
4. If the response is from login/register, extract the token and save/update it in `./token.md` under a `## role` section (Token, User, Obtained).
5. On 401/expired token, re-login and refresh `token.md`.

Rules:

- Always show the curl command before running it.
- Use targeted grep, not full-file/codebase reads — save context.
- Make sure `token.md` is in `.gitignore`; add it if missing.
