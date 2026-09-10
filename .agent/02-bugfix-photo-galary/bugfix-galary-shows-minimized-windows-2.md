# Gallery thumbnails squashed on Windows — round 5

Continues [bugfix-galary-shows-minimized-windows.md](bugfix-galary-shows-minimized-windows.md). Rounds 1-4 there ruled out
the CSS height chain, WebKit-vs-Blink engine differences, Vue's async image mount timing, and WebView2 asset caching.

## The new clue

Two ways of launching the same code behave differently on the same Windows machine:

- `dev.ps1` (`wails dev`) — **thumbnails render correctly**, no bug.
- `run.ps1` (the built `build\bin\PhotoWithOverlay.exe`) — **bug present**.

Same source, same machine, same WebView2. So whatever causes this lives in the difference between those two launch
modes, not in the stylesheet.

## What actually differs between `dev.ps1` and `run.ps1`

Checked against the Wails v2.15.0 source in the Go module cache rather than from memory:

|                        | `dev.ps1` (`wails dev`)                                                          | `run.ps1` (built exe)                                        |
| ---------------------- | -------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| URL the WebView2 loads | `http://wails.localhost:34115/` (dev server port from `internal/project/project.go:182`, proxying Vite on 5173) — confirmed by the DevTools window title | `http://wails.localhost/` — `internal/frontend/desktop/windows/frontend.go:40` |
| Frontend form          | unbundled ES modules; `style.css` injected as a `<style>` tag by JS               | one minified bundle + `<link rel=stylesheet>` in `<head>`    |
| Build tags             | `dev` → DevTools always on (`internal/app/app_dev.go:243`)                        | production → DevTools off unless built with `-devtools`      |
| Boot speed             | slow (per-module transform on every request)                                      | fast (single ~100 KB bundle)                                 |

### Ruled out: the CSS delivery path

Diffed the emitted `frontend/dist/assets/index-DZdjrJIN.css` against `frontend/src/style.css`, rule by rule, across the
whole gallery chain — `main`, `.camera-panel`, `.gallery-wrap`, `#gallery`, `.photo`, `.photo-open`, `.photo .thumb`,
`.photo img`, both media queries, and the `main.gallery-closed` overrides. **Semantically identical**; the minifier only
collapsed shorthands and added `-webkit-` prefixes.

Also confirmed there are **zero `<style>` blocks in any `.vue` file** — all CSS comes from the single `style.css`
imported in `main.ts:3`. So there is no dev-vs-prod cascade-ordering shuffle either (that is the usual suspect when a
Vite app renders differently in dev and prod, and it does not apply here).

Build freshness checked too: `dist` and the exe were both produced at 14:22, `style.css` was last modified at 11:16 —
the exe is not stale.

## The difference that does bite: `localStorage` is per-origin

Dev serves the app from `http://wails.localhost:34115` and the release build from `http://wails.localhost` (port 80).
Same host, **different port, therefore different origin** — so the two get completely separate localStorage. And the
gallery drawer's open/closed state is persisted there:

```ts
// frontend/src/composables/useLayout.ts:11
const galleryClosed = ref(savedState('galleryDrawerClosed', compactScreen));
```

`toggleGallery()` (`useLayout.ts:36-40`) writes that key on every manual toggle. The 3-second auto-collapse
(`useLayout.ts:23-27`) deliberately does **not** write it. Consequences:

- **`run.ps1`** — the production origin holds whatever state the gallery was last left in. If that value is `"true"`,
  the app starts with `main.gallery-closed` applied, and that rule sets:

  ```css
  main.gallery-closed .camera-panel { grid-template-rows: minmax(0, 1fr) 0 }
  ```

  The gallery row is therefore **0 px tall at first paint**. Every `PhotoCard` mounts and resolves its
  `await Thumbnail(...)` (`PhotoCard.vue:16-18`) inside that zero-height subtree. The gallery is only opened afterwards,
  by clicking the drawer handle.

- **`dev.ps1`** — fresh origin, key absent, so `savedState` falls back to `compactScreen`, which is `false` on a wide
  window. The gallery starts **open**, cards mount into a correctly sized container, and the percentage chain resolves
  normally.

The bug screenshot agrees: the `#galleryDrawerToggle` handle is still wearing its focus ring, i.e. the gallery had just
been opened by a click rather than having been open all along.

### Why the hairlines mean "laid out while collapsed"

Each card in the bug screenshot renders at essentially just its own `border: 2px` — about 4 px total. That means the
`<img>` contributes **zero** height.

That is the important part: if only the percentage chain were broken while the thumbnail itself had a size,
`.photo img { width: 100% }` against a 150 px-wide `.thumb` plus the JPEG's own aspect ratio would still produce a
roughly 112 px-tall card. A hairline requires the image to have contributed no height at all — which is what you get
when the subtree's layout is the one it had while the row was 0 px, and it was never re-resolved on reopen.

## Why round 3's Chrome reproduction missed this

Round 3 tested, in real Blink:

1. old vs new DOM-construction order for the `<img>`,
2. the compiled `dist/` app with a stubbed backend,
3. open → 3-second auto-collapse → scripted reopen,
4. thumbnails delayed 4-7 s, arriving while collapsed, then reopen.

All four **start from a gallery that was already laid out open**. The untested scenario is the one above: *collapsed at
first paint, cards mounting for the very first time into a 0 px grid row.* That state is only reachable when
localStorage says the drawer is closed — which only ever happens on the production origin, never on a fresh dev origin.
That is exactly why the bug tracks `run.ps1` vs `dev.ps1` and not macOS vs Windows.

## Status

Hypothesis, not yet confirmed. The origin split and the resulting localStorage split are verified facts; the claim that
mounting into the 0 px row is what strands the layout still needs a measurement.

### Result of next step 1 — inconclusive, test needs re-running

Ran the localStorage flip below. The gallery bug did **not** reproduce visually. But the console showed:

```
template.js:42 Uncaught TypeError: Cannot read properties of null (reading 'nodes')
    at new In (Overlay.svelte:1:1)
    at main.js:31:26
```

That stack is **not from this app**. `frontend/package.json` lists exactly one dependency — `vue` — and a
case-insensitive search for "svelte" across `frontend/src`, `frontend/dist`, `frontend/wailsjs` and
`frontend/index.html` returns nothing. `Overlay.svelte` / `main.js` / `template.js` is injected third-party code, i.e. a
browser extension content script (a lot of extensions are built with Svelte).

Which matters, because **WebView2 does not load browser extensions** by default and Wails does not enable them. So that
console almost certainly belonged to an ordinary Chrome/Edge tab — `dev.ps1 -Browser` — not to the WebView2 window
`run.ps1` uses. A browser tab is a *third* environment, and a passing result there does not clear WebView2. The
"doesn't show in the browser" observation from the start of this investigation is subject to the same caveat.

The test needs re-running inside the WebView2 dev window: plain `dev.ps1` (no `-Browser`), DevTools opened in the app
window itself. That console shows `[vite] connected` and no extension noise, which is how you know you are in the right
window.

### Opening DevTools in the WebView2 window

**Plain F12 does nothing** — Wails disables the browser accelerator keys unconditionally
(`PutAreBrowserAcceleratorKeysEnabled(false)`, `internal/frontend/desktop/windows/frontend.go:596`). The three routes
that do work:

1. **Ctrl+Shift+F12** — Wails' own accelerator, gated on devtools being enabled (`frontend.go:508-517`).
2. **Right-click → Inspect** — the default context menu is enabled in dev and `-debug` builds (`frontend.go:569`).
3. `Debug: options.Debug{OpenInspectorOnStartup: true}` in `main.go` (`pkg/options/debug.go:5`) — opens the inspector
   automatically at startup. Gated on `f.debug`, so it applies to `wails dev` and `wails build -debug`, not to a plain
   `-devtools` build (`frontend.go:602`).

Note that Ctrl+R is disabled by the same accelerator-key setting; reload from the console with `location.reload()`.

### Round 5b — the re-run was still not a valid test

DevTools opened correctly this time (title bar: `DevTools - wails.localhost:34115/`, console shows
`[vite] connecting… / connected`, so it is the real dev window). But the precondition check returned:

```js
document.querySelector('main').classList.contains('gallery-closed')   // → false
```

`false` means `main.gallery-closed` was **not** applied, so the cards did not mount into a 0 px row and the hypothesis
was never exercised. Two likely reasons:

1. The earlier `localStorage.setItem('galleryDrawerClosed','true')` was run in the **browser** tab. That is a different
   origin, so this window never had the key.
2. The check was run **after** clicking the Gallery handle open. `toggleGallery()` (`useLayout.ts:36-40`) removes the
   class *and* sets `userToggledDrawer = true`, which permanently disables the 3-second auto-collapse
   (`useLayout.ts:24`) — so from that point on the class can never come back on its own.

Order matters and there is a 3-second budget: check the class *immediately* after the reload and *before* touching the
handle. Exact sequence, all in the dev window's console:

```js
// 1. same origin as the app under test — set the key here, not in a browser tab
localStorage.setItem('galleryDrawerClosed', 'true');
// 2. tick "Keep log" in the console settings first, so the reload does not clear the output
location.reload();
// 3. IMMEDIATELY after reload, before clicking anything:
[document.querySelector('main').className,
 getComputedStyle(document.querySelector('.gallery-wrap')).height,   // expect "0px"
 localStorage.getItem('galleryDrawerClosed')]                        // expect "true"
// 4. only now click the Gallery handle, and look at the cards
```

If the class is present at step 3 and the cards still render at full height after step 4, the mount-into-0 px-row
mechanism is dead as an explanation. The remaining dev-vs-prod deltas would then be: bundled+minified single-file
frontend with `<link rel=stylesheet>` in `<head>` versus JS-injected `<style>`, much faster boot, and DevTools being
present — none of which are diagnosable by reading code. At that point only next step 2, measuring inside the actual
failing build, can make progress.

## Next steps to confirm

### 1. Reproduce in dev — no rebuild, ~1 minute

Run `dev.ps1`, open DevTools with **Ctrl+Shift+F12** or right-click → Inspect (not F12 — see "Opening DevTools in the
WebView2 window" above), then follow the exact 4-step sequence in round 5b. If the cards come up as hairlines, the bug
is reproduced **in dev with DevTools attached**, and these give the answer directly:

```js
const img = document.querySelector('.photo img');
img.src.slice(0, 40); img.naturalWidth; img.naturalHeight;
['.gallery-wrap', '#gallery', '.photo', '.photo-open', '.thumb', '.photo img']
  .map(s => [s, getComputedStyle(document.querySelector(s)).height]);
```

### 2. Give the release build DevTools — one rebuild

Wails supports DevTools in a production build via the `devtools` build tag (`internal/app/app_devtools.go:1`), enabled
with `wails build -devtools`. Prefer **`wails build -debug`** though: it turns on devtools *and* `f.debug`, which is
what also enables right-click → Inspect and honours `OpenInspectorOnStartup`. With `-devtools` alone, only
Ctrl+Shift+F12 works. Either way `run.ps1`'s exe becomes inspectable in the exact failing configuration. First thing to
check there:

```js
localStorage.getItem('galleryDrawerClosed')    // if "true", the asymmetry above is confirmed outright
```

Open question for review: add a `-DevTools` switch to `build.ps1` so this is one command?

### 3. Zero-tooling probe on the current exe

In the failing build, open the gallery and capture a new photo. Cards are keyed by `photo.path`
(`GalleryPanel.vue:37`), so existing DOM is reused and only the new card mounts — this time while the gallery is open.
If the new card is full height while the older ones stay hairlines, that pins the fault to layout state at mount time.

## If confirmed, likely fixes

Not implemented yet — listed for review, cheapest first:

1. Don't let the gallery subtree be sized 0 while it holds mounted content: collapse the drawer with
   `transform`/`opacity`/`visibility` only (`main.gallery-closed .gallery-wrap` already does this) and stop driving the
   grid track to `0` in `main.gallery-closed .camera-panel`. Keep the row at `var(--gallery-height)` and let the
   translate hide it.
2. Give the thumbnail an intrinsic size so its box never depends on the ancestor chain: set `width`/`height` attributes
   (or `aspect-ratio`) on the `<img>` in `PhotoCard.vue:39`, since `Thumbnail` always emits 280 px-wide JPEGs
   (`internal/photo/thumbnail.go:48`).
3. Defer mounting `PhotoCard`s until the drawer is actually open (`v-if` on the gallery contents), so nothing ever lays
   out inside the collapsed row.

Unrelated but worth noting from the same session: `PhotoCard.vue:16-18` has no error handling around
`await Thumbnail(...)`. If that call ever rejects, `thumbSrc` stays `''`, the `<img>` gets no `src` at all, and the card
silently becomes the same 4 px hairline — indistinguishable from this bug in the release build, where there is no
console to see the rejection.

## Side note: "Capture failed: Timeout expired" seen while testing

Not this bug, and resolved. `Timeout expired` is Blink's verbatim message for `GeolocationPositionError.TIMEOUT`,
raised by `lookupBrowserLocation()` (`frontend/src/composables/useLocation.ts:23-30`, `timeout: 20000`) and surfaced by
the catch in `App.vue:132`. Windows has no native location path — `tryNativeLocation()` is macOS-only
(`useLocation.ts:13-21`) and `native_location.go:7-9` is a hard stub for `!darwin` — so Windows depends entirely on
`navigator.geolocation`. Cause was the OS-level Windows location setting; turned on, it works.

Still latent: a GPS failure aborts the whole capture and no photo is written at all. Worth deciding separately whether
capture should degrade gracefully, fail faster than 20 s, or gain a native Windows location provider
(`Windows.Devices.Geolocation`, mirroring `native_location_darwin.go`).
