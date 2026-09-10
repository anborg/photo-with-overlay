# Gallery thumbnails squashed on Windows — round 5

Continues [bugfix-galary-shows-minimized-windows.md](bugfix-galary-shows-minimized-windows.md). Rounds 1-4 there ruled out
the CSS height chain, WebKit-vs-Blink engine differences, Vue's async image mount timing, and WebView2 asset caching.

> **SOLVED — see [Round 6](#round-6--root-cause-found-and-measured) at the bottom.** The gallery collapses whenever the
> UI zoom is above 1.0, because `#gallery { padding-top: 58px }` plus the horizontal scrollbar consume the entire
> gallery track once the zoom rules shrink it from 160 px to 92 px / 76 px. Not an engine bug, not a Vue regression, not
> platform-specific. Everything between here and Round 6 is the trail that got there, including several dead ends.

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

## Round 5c — valid test, and the hypothesis is DEAD

The sequence above finally ran under the intended condition. Step 3:

```
['settings-closed gallery-closed', '0px', 'true']
```

Precondition held: `main.gallery-closed` present at first paint, `.gallery-wrap` measured **0 px**, localStorage key
`"true"`. So the `PhotoCard`s did mount and resolve their thumbnails inside a zero-height row — exactly the scenario
rounds 5/5a/5b were trying to reach.

After clicking the Gallery handle open, step 4:

```
['data:image/jpeg;base64,/9j/2wC', 280, 157,
 '.gallery-wrap = 160px', '#gallery = 146.143px', '.photo = 73.2607px',
 '.photo-open = 69.5397px', '.thumb = 69.5397px', 'img = 69.5397px']
```

**Those numbers are correct, not broken.** Working it through (note `getComputedStyle().height` reports the border-box
height here, because `* { box-sizing: border-box }`):

| box             | expected                                                              | measured      |
| --------------- | --------------------------------------------------------------------- | ------------- |
| `.gallery-wrap` | `--gallery-height` = 160 px                                           | 160 px ✓      |
| `#gallery`      | 160 − 6 − 6 padding − 2 border-top = 146 px                           | 146.143 px ✓  |
| `.photo`        | 146.143 − 58 `padding-top` − ~14.9 scrollbar = **73.3 px**            | 73.2607 px ✓  |
| `.photo-open`   | 73.26 − ~3.7 (2 px borders, DPI-snapped)                              | 69.5397 px ✓  |

The ~14.9 px that at first looked like a shortfall is the **horizontal scrollbar** of
`#gallery { overflow-x: auto }` — visible in the screenshot. Every link in the percentage chain resolved, the thumbnail
loaded (`data:image/jpeg;base64,…`, natural size 280×157), and the screenshot shows proper rectangular cards with
photos in them.

So: mounting the cards into a 0 px grid row **does not** strand the layout, and the localStorage/origin asymmetry —
while a real difference between the two launch modes — is not the cause of this bug. Rounds 1-5c have now eliminated
every theory reachable from reading code or from the dev build.

Retained from this round as useful facts:

- A correct gallery has `.photo` ≈ 73 px, not ≈ 146 px: 58 px goes to `#gallery`'s `padding-top` (clearance for the
  absolutely positioned selection bar) and ~15 px to the horizontal scrollbar. Use ~73 px as the reference when
  comparing against the failing build.
- The deduction in "Why the hairlines mean laid out while collapsed" still holds for the release build, just without the
  collapsed-at-mount explanation: a 4 px card means the `<img>` contributed **zero** height there, whereas here a
  loaded 280×157 thumbnail is present and the chain resolves normally.

Only next step 2 remains: measure inside the build that actually fails.

## Round 6 — root cause found and measured

Built with `wails build -debug`, ran `run.ps1`, reproduced the bug, and measured it in place. Console:

```
['settings-closed gallery-closed zoom-focus', '0px', 'false']

['data:image/jpeg;base64,/9j/2wC', 280, 157,
 '.gallery-wrap = 75.9981px', '#gallery = 62.1415px', '.photo = 3.72093px',
 '.photo-open = 0px', '.thumb = 0px', 'img = 0px']
```

Two things jump out. `main`'s class list contains **`zoom-focus`**, which was absent in the dev window. And
`.gallery-wrap` is **76 px**, not the 160 px it was in dev.

### The chain, measured

| box             | why it is that size                                                    | measured       |
| --------------- | ---------------------------------------------------------------------- | -------------- |
| `.gallery-wrap` | `main.zoom-focus .camera-panel` forces the gallery track to **76 px**  | 75.9981 px     |
| `#gallery`      | 76 − 12 padding − 2 border-top = 62 px                                 | 62.1415 px     |
| content left    | 62.14 − **58 px `padding-top`** = 4.14 px, then ~14.9 px scrollbar → **negative, clamped to 0** | — |
| `.photo`        | `height: 100%` of 0, plus its own 2 px borders (DPI-snapped)           | **3.72093 px** |
| `.photo-open` / `.thumb` / `img` | `height: 100%` of 0                                   | **0 px**       |

The thumbnail itself is perfectly healthy — `data:image/jpeg;base64,…`, natural size 280×157. It is simply being
rendered into a box with zero height. `.photo = 3.72 px` is nothing but its own border, which is exactly the hairline in
the original bug screenshot.

### Root cause

`#gallery { padding-top: 58px }` (`frontend/src/style.css:254-256`) reserves clearance for
`.gallery-selection-bar`, which is `position: absolute; height: 52px` (`style.css:206-216`) and therefore out of flow.
That 58 px is a **fixed** cost, while the gallery track height is not:

- `.camera-panel { grid-template-rows: minmax(0, 1fr) var(--gallery-height) }` — `style.css:977-982`, 160 px normally
- `main.zoom-compact .camera-panel { grid-template-rows: minmax(0, 1fr) 92px }` — `style.css:675-677`
- `main.zoom-focus .camera-panel { grid-template-rows: minmax(0, 1fr) 76px }` — `style.css:681-683`

Card height available = track − 14 (`.gallery-wrap` padding 12 + border-top 2) − 58 (`#gallery` padding-top) − ~15
(horizontal scrollbar from `overflow-x: auto`):

| mode           | track | card height                | result                     |
| -------------- | ----- | -------------------------- | -------------------------- |
| normal         | 160   | 160 − 14 − 58 − 15 = **73** | fine (matches round 5c)   |
| `zoom-compact` | 92    | 92 − 14 − 58 − 15 = **5**   | hairline                  |
| `zoom-focus`   | 76    | 76 − 14 − 58 − 15 = **−11** | clamped to 0 — hairline   |

So the gallery is broken in **both** zoom modes and fine only at zoom 1.0. Nothing platform-specific about it.

### Why it looked like a Windows-only / dev-vs-prod / post-Vue bug

`zoom` is read from `localStorage.getItem('uiZoom')` (`useViewPreferences.ts:24`) and the classes are derived from it:
`zoomCompact = zoom > 1 && zoom < 1.3`, `zoomFocus = zoom >= 1.3` (`useViewPreferences.ts:30-31`).

- **dev vs prod** — localStorage is per-origin, and dev runs on `wails.localhost:34115` while the release build runs on
  `wails.localhost`. The dev origin had no `uiZoom` (class list was just `settings-closed gallery-closed`, track
  160 px); the production origin had `uiZoom >= 1.3`. The round-5 instinct that a per-origin localStorage key explained
  the split was right in kind — just the wrong key. It is `uiZoom`, not `galleryDrawerClosed`.
- **Windows vs macOS** — nothing to do with WebKit vs Blink. Someone pressed the zoom-in control on the tablet (very
  plausible on a small rugged screen) and it persisted; the Mac is still at 1.0.
- **"not present before Vue"** — retired. `git grep` on the pre-Vue commit `d9de302` shows
  `frontend/dist/view-preferences.js:15-22` used the *same* `uiZoom` key with the *same* 1.0/1.3 thresholds, and
  `frontend/dist/style.css:673-682` had the same 92 px / 76 px overrides. The defect predates the Vue migration; the
  zoom setting simply got changed around that time.

### Secondary defect spotted in the same screenshot

The two zoom systems disagree. `main.zoom-focus` also sets `--gallery-height: 105px` (`style.css:1084-1088`), but
`main.zoom-focus .camera-panel` (specificity 0,2,1) beats `.camera-panel` (0,1,0), so the **track** is 76 px while
`--gallery-height` still reads 105 px. Since `.gallery-drawer-handle { bottom: var(--gallery-height) }`
(`style.css:1039-1045`), the Gallery button is positioned 105 px up while the band it belongs to is only 76 px tall —
which is why it floats above the gallery with a black gap in the screenshot. The `grid-template-rows` overrides at
`style.css:675-683` are leftovers from the pre-drawer layout and should not be fighting `--gallery-height` at all.

### Immediate workaround (no rebuild)

Press the zoom-out control in the header until the zoom is back to 1.0, or from the console:

```js
localStorage.setItem('uiZoom', '1'); location.reload()
```

### Reproducing at will, in dev

The bug is now a one-liner in the dev window, which makes verifying any fix trivial:

```js
localStorage.setItem('uiZoom', '1.3'); location.reload()   // zoom-focus → hairlines
localStorage.setItem('uiZoom', '1.1'); location.reload()   // zoom-compact → hairlines
localStorage.setItem('uiZoom', '1');   location.reload()   // normal → correct
```

### Fix options

The core problem is that a fixed 58 px clearance cannot coexist with a 76-92 px track. Three changes compose; numbers
below are the resulting card heights.

1. **Delete the stale track overrides** at `style.css:675-677` and `style.css:681-683` so `--gallery-height` is the
   single source of truth (130 px compact / 105 px focus). Also fixes the drawer-handle gap above. → compact 43 px,
   focus 18 px.
2. **Make the bar's footprint a variable** instead of a magic number: `--gallery-bar: 52px` driving both
   `.gallery-selection-bar { height: var(--gallery-bar) }` and `#gallery { padding-top: calc(var(--gallery-bar) + 6px) }`,
   set to `40px` under `main.zoom-compact` / `main.zoom-focus`. → compact 55 px, focus 30 px.
3. **Stop paying 15 px for the scrollbar**: `#gallery { scrollbar-width: thin }` (~8 px instead of ~15). → compact
   62 px, focus 37 px.

With all three: normal 80 px, compact 62 px, focus 37 px — every mode shows a real thumbnail.

Structural alternative, more invasive but immune to this class of bug: make `.gallery-wrap` a flex column, put
`.gallery-selection-bar` back in flow as `flex: 0 0 auto`, and give `#gallery` `flex: 1 1 auto; min-height: 0` with no
`padding-top`. The leftover space can then never go negative — though the bar still needs to shrink in zoom modes for
the cards to be usefully tall.

Blunter option if bigger cards in zoom modes matter more than the Select/Delete affordance: hide the selection bar in
zoom modes and drop the padding, giving focus 76 px and compact 101 px cards.

## Round 7 — fix applied

Took the structural route rather than re-tuning the magic numbers, because it **deletes** the coupling instead of
maintaining it: the selection bar now takes the height it needs, `#gallery` takes whatever is left, and no rule
reserves space with a hard-coded number. There is nothing left to get out of sync when `--gallery-height` changes.

### `frontend/src/style.css`

1. `.gallery-wrap` — added `display: flex; flex-direction: column`.
2. `.gallery-selection-bar` — dropped `position: absolute`, `inset: 0 0 auto 0`, `z-index: 4` and `height: 52px`; now
   `flex: 0 0 auto` with `padding: 0 10px 6px`, so it is in normal flow and sized by its own content (the 38 px delete
   button). Background and border-bottom kept, so it still reads as a header strip.
3. `#gallery` — the two duplicate rules merged into one: `flex: 1 1 auto; min-height: 0`, replacing `height: 100%` and
   **deleting `padding-top: 58px`** — the actual cause. Added `scrollbar-width: thin` to stop the horizontal scrollbar
   eating ~15 px.
4. Deleted `main.zoom-compact .camera-panel { grid-template-rows: … 92px }` and
   `main.zoom-focus .camera-panel { grid-template-rows: … 76px }`. `--gallery-height` (130 px / 105 px) is now the only
   thing that sets the gallery height, which also fixes the drawer-handle gap, since the handle is positioned at
   `bottom: var(--gallery-height)`.
5. Deleted the dead `grid-template-rows: minmax(0, 1fr) 105px` from `.camera-panel` in the `max-width: 700px` media
   query — it was already being overridden by the later `.camera-panel` rule at equal specificity, so it had no effect.
6. Deleted the `.gallery-selection-bar { height: 46px }` and `#gallery { padding-top: 52px }` overrides from the mobile
   media query; with the bar in flow there is nothing to keep in sync. Kept its `padding` tweak.

### `frontend/src/composables/useLayout.ts`

Removed the 3-second auto-collapse: the top-level `setTimeout`, the `userToggledDrawer` module flag it needed, and the
two `userToggledDrawer = true` assignments in `toggleSettings`/`toggleGallery`. The drawers now do exactly one thing —
what the user last set, restored from localStorage. Both toggle functions are two lines each.

The CSS drawer animations (`transition: transform/opacity`) were left in place: they are declarative CSS and touch no
state logic, so they were not part of what made this hard to reason about. The `prefers-reduced-motion` block that
disables them stays too.

### Expected card heights after the fix

Track − 14 (`.gallery-wrap` padding 12 + border-top 2) − ~45 (bar: 38 px button + 6 px padding + 1 px border) − ~8
(thin scrollbar):

| mode           | track  | card height |
| -------------- | ------ | ----------- |
| normal         | 160    | ~93 px      |
| `zoom-compact` | 130    | ~63 px      |
| `zoom-focus`   | 105    | ~38 px      |
| mobile normal  | 140    | ~73 px      |

Every mode now leaves real room for a thumbnail. Previously: 73 / 5 / −11.

### Verified so far

- `vue-tsc --noEmit && vite build` clean; `go test ./...` passes; `wails build -debug` succeeds.
- Emitted `dist` CSS checked directly: no `padding-top: 58px`/`52px` and no `92px`/`76px`/`105px` track overrides
  remain, and the `--gallery-height` values are intact.
- Confirmed working on the Windows tablet: full-height thumbnails in the gallery.
- Still unverified: the `zoom-compact` (`uiZoom` 1.1) and mobile/portrait breakpoints.

### Follow-up — translucent chrome strips

The opaque black bars either side of the gallery were raised separately. Both now use the same translucency as `aside`,
which was already the app's most transparent surface:

- `.gallery-selection-bar` — was `color-mix(in srgb, var(--panel) 92%, #000)`, i.e. effectively solid black. Now
  `color-mix(in srgb, var(--panel) 30%, transparent)` plus `backdrop-filter: blur(7px)` to match `.gallery-wrap`. The
  camera preview shows through it, the same way it already showed through the gallery band around the thumbnails.
- `footer` — was 72 %, now 30 %, for one consistent value across the chrome.

Note the footer will still read as dark: it sits in `main`'s second grid row, **below** `.camera-panel`, so what is
behind it is the page background (`--page`), not the camera preview. Translucency cannot reveal a photo that is not
painted there. Showing the preview through the footer needs a layout change — the footer would have to overlay
`.camera-panel`, which is where the gallery band already lives, so the two would have to be stacked or the gallery
moved. Not attempted here.

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
