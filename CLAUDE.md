# Working in this repo

`apicorex-ui` — the suite's shared UI kit: shell, table, form, icons, i18n,
Bangla formatting, and the palette. Public. Consumed by Schoolyze's panel,
Accounting's panel and Identity's login.

It lived at `schoolyze-server/ui` until 2026-08-23. It moved because the
platform layer may not import a domain product's repository, and the suite's
login belongs in Identity. The history came with it.

## The rule that makes this reusable

**Nothing here may know what a student is.** No domain types, no plugin
imports — a component that knows what a student is belongs with the screens.
That discipline is the reason the extraction was mechanical when it finally
happened, and it is the only thing keeping this usable by a second product.

## Three things that fail silently

- **Tailwind cannot see these templates from a consumer.** The kit's path is a
  module cache directory with a version in it, so consumers generate their
  `@source` line with `cmd/kitsources` rather than hardcoding it. Get it wrong
  and every class the kit uses is stripped as unused: the build succeeds, the
  tests pass, and the panel renders completely unstyled. Consumers guard it with
  a byte-size floor on the compiled stylesheet — a check that probes for class
  names passes while proving nothing.
- **An unknown icon name renders an empty `<svg>`**, not an error. Three
  mistyped names once looked like a styling problem in a whole panel.
- **A consumer's local build is not CI's build.** `go.work` points at this
  checkout; Docker and Jenkins use `GOWORK=off` and the published tag. A change
  here is not real to them until it is tagged, and a published tag must never be
  re-pointed — the proxy caches it, so consumers silently keep the old content.

## Strings and theme

Kit strings are namespaced `kit.*` and registered, not embedded, so two plugins
in one suite can each have their own catalogue without colliding. `theme.css`
holds the suite palette; `cmd/kitsources --theme` copies it into a consumer.
Before it lived here, a second panel rendered in DaisyUI's stock purple beside
Schoolyze's teal — one shell that looked like two companies.

## Committing

Do not run `git commit` or `git push` unless the person you are working with
asks for it in that moment. Editing files freely is fine; publishing is theirs
to decide — and here publishing means a tag other repos will resolve.
