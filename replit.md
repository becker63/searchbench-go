# Replit / quick start

Short version: use Go + Pkl as in [`README.md`](README.md).

**Full Replit-oriented guide:** [`docs/guides/replit.md`](docs/guides/replit.md)

**Nix (flake):** `nix develop` installs Git hooks; **`git commit`** stages Repomix then runs **`buck2 test //:check`**; **`git push`** runs **`buck2 test //:check_full`**. See [`AGENTS.md`](AGENTS.md).
