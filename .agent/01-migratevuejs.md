# Frontend migration to Vue 3 + TypeScript + Vite

This documents the instructions and decisions given during the session that
migrated `frontend/` from hand-written vanilla JS (served unbundled out of
`frontend/dist/`) to Vue 3 + TypeScript, built with Vite. Follow these when
continuing frontend work or making similar build/tooling decisions elsewhere
in this repo.

## Framework and scope decisions

- **Vite as the build tool, modeled on Wails' own official templates** —
  not a bespoke setup. When in doubt about config shape (`vite.config.ts`,
  `tsconfig.json`, `wails.json` fields, `frontend/index.html` entry,
  generated-bindings layout under `frontend/wailsjs/go/`), fetch and match
  the relevant official Wails `*-ts` template
  (`wailsapp/wails` repo, `v2/pkg/templates/templates/`) rather than
  improvising.
- **Vue 3 + TypeScript, not React** — chosen deliberately for this specific
  codebase. The app is built around small imperative controllers doing
  direct DOM/pointer/canvas manipulation (watermark dragging, `ResizeObserver`,
  canvas drawing). Vue's `<script setup>` composables map onto that style
  directly; React would push the same imperative logic into
  `useRef`/`useEffect` escape hatches anyway. Don't relitigate this choice
  without a reason specific to new work, not a generic framework preference.
- **Full rewrite, not a Vite-only wrapper** — every module (`camera.js`,
  `gallery.js`, `layout.js`, `location.js`, `settings.js`,
  `view-preferences.js`, `watermark.js`, `dom.js`) was ported into a Vue
  composable of the same name/responsibility (`useCamera`, `useGallery`,
  etc.), and `app.js`'s orchestration logic now lives in `App.vue`. Preserve
  this 1:1 module mapping when adding new frontend features — new
  responsibilities get their own composable, not code stuffed into an
  existing one or into a component.
- **Real generated Wails bindings, not hand-written `window.go.main.App`
  calls.** Bindings live in `frontend/wailsjs/go/main/App.{js,d.ts}` and
  `frontend/wailsjs/go/models.ts`, produced by
  `go run github.com/wailsapp/wails/v2/cmd/wails@<version> generate module`
  (or a full `wails build`/`wails dev` without `-skipbindings`). Never
  hand-author these — regenerate them after any change to bound methods on
  `App` in `app.go`.

## Warnings: fix the root cause, don't silence

This was raised explicitly and applies beyond this one file:

> Why use deprecated API, and silence warnings? No option to upgrade?

When a compiler/linter warning shows up (deprecation, availability, etc.):

1. **First ask whether the constraint the warning is protecting against is
   still real.** Here, that meant checking what macOS version the app
   actually claims to support (`build/darwin/Info.plist`
   `LSMinimumSystemVersion`) before assuming the old-OS fallback code was
   needed at all. It wasn't going to be reachable on the app's declared
   floor for one of the two warnings, and the floor itself was up for
   negotiation for the other.
2. **Prefer deleting the need for the workaround over suppressing the
   symptom.** Confirmed with the user that macOS 10.15–10.x support could be
   dropped (long past Apple EOL); raised `LSMinimumSystemVersion` to
   `11.0.0` and deleted all the `@available`/deprecated-fallback branching
   in `native_location_darwin.go` entirely, rather than wrapping the
   deprecated call in `#pragma clang diagnostic ignored`. A pragma-based
   suppression was tried first and explicitly rejected once a real upgrade
   path existed — only reach for suppression when the old behavior is
   genuinely still required (e.g. a real minimum-OS/API constraint that
   can't be moved) and even then, comment *why*.
3. **Chase warnings to their actual source, not just the nearest fix.** The
   `requestWhenInUseAuthorization` availability warning wasn't a code
   problem at all — it was Wails' Go toolchain defaulting the C compile
   target to `-mmacosx-version-min=10.13`, independent of what the Info.plist
   declared. The fix was pinning `#cgo CFLAGS: -mmacosx-version-min=11.0` in
   `native_location_darwin.go` (CFLAGS only — putting it on LDFLAGS breaks
   linking against Go's own runtime objects built for the host SDK).
4. **A support-matrix change (minimum OS version, minimum browser, etc.) is
   a product decision, not a unilateral technical one** — it was confirmed
   with the user via a direct question before touching `Info.plist` or
   deleting the compat code, even though "drop it" was the recommended
   option.

## Verification practice for this kind of change

- `go build ./...`, `go vet ./...`, `go test ./...` after every backend/cgo
  change.
- For anything touching deployment-target-sensitive code
  (`native_location_darwin.go`'s CoreLocation calls), verify under **both**
  the plain toolchain default and Wails' actual (lower) default explicitly:
  `CGO_CFLAGS="-mmacosx-version-min=10.13" go build -a ./...` — don't trust
  a clean plain `go build` alone, since Wails' build environment differs.
- For the frontend: `npx vue-tsc --noEmit` and `npm run build` (which runs
  the same) must be clean; then actually run the app
  (`wails dev`, drive the browser-debug bridge at `http://localhost:34115`
  with Playwright, or the native window directly) and exercise real user
  flows — a clean type-check is necessary but not sufficient.
- After a fix, re-diff (`git diff --stat`, read the actual diff) to confirm
  the change is exactly what's expected — no stray formatting, no
  unintended file-mode changes, no leftover dead branches.
