# apicorex-ui

The shared UI kit for the [ApiCoreX](https://github.com/msrsiddik/apicorex)
suite: the application shell, tables, form controls, icons, and the bilingual
formatting every screen and printed document depends on.

```go
kit.Shell(kit.Page{
    Title:    i18n.T(ctx, "account.title"),
    AppName:  i18n.T(ctx, "app.name"),   // your product's name, not the kit's
    BasePath: "/accounting",
    Locale:   i18n.FromContext(ctx),
    Nav:      merged,
}, body)
```

## What belongs here, and what does not

**Here:** anything a second product could use unchanged — layout, chrome,
controls, the i18n mechanism, Bangla numerals and taka formatting.

**Not here:** any product's vocabulary. A component that knows what a student is
belongs with the screens. That rule is what kept this reusable while it lived
inside one product, and it is why it could be lifted out at all.

## Strings

The kit carries only its own, namespaced `kit.*` — roughly two dozen words for
buttons, error pages and table chrome. A consumer registers what it says itself:

```go
//go:embed locale/*.toml
var localeFS embed.FS

for _, l := range i18n.Supported {
    i18n.RegisterFS(l, localeFS, "locale/"+string(l)+".toml")
}
```

The namespace is what makes this safe: a consumer's `nav.students` can never
collide with a kit key, so no merge-precedence rule is needed — and precedence
rules are always wrong for somebody.

`Page.AppName` is a field rather than a catalogue key for the same reason. It
used to be `i18n.T("app.name")`, which resolved to whichever product had
registered it; with two products the ledger's pages came out branded with the
school's name.

## Tailwind

Tailwind only generates the classes it can see, and these templates are in a
module directory with a version in its path. Point `@source` at it by generating
the line rather than writing it:

```sh
go list -m -f '{{.Dir}}' github.com/msrsiddik/apicorex-ui
```

Get this wrong and every class the kit uses is stripped as unused: the page
compiles, deploys, and renders completely unstyled, with nothing failing
anywhere. Check the built stylesheet's size in CI — see
`apicorex-accounting/scripts/check-css.sh` for one way.

## Where it came from

`schoolyze-server/ui`, whose git history holds everything before v0.1.0.
Extraction was proposed, rejected, and reopened: the deciding argument was not
tidiness but the layer model. The platform kit is defined as importing no
plugin, and the suite's login belongs in Identity — which would have meant the
platform layer depending on a school product's repository.

`git filter-repo` was unavailable on the machine where the move happened, so the
files were copied rather than replanted with their commits.
