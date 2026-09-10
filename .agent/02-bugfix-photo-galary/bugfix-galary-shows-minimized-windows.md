Expected: Galary is supposed to show thumbnail of past photos.

Bug issue: It shows correctly in macos. However in windows, the thumnail area looks like a line instead of rectangle with taken pictures.
![bad-galary-in-windows.png](bad-galary-in-windows.png)
The bug is not present in macos: ![good-galary-in-macos.png](good-galary-in-macos.png)
This is ui related - i.e frontend/ code

Todo: Investigate the reason, and document here. 

## Investigation

Platform note: this is a Wails app. On macOS the webview is WKWebView (Safari/WebKit engine); on Windows it's WebView2 (Chromium/Edge engine). So "macOS vs Windows" here really means "WebKit vs Chromium" — a plain CSS rendering difference between engines, not a Go/backend issue.

Relevant markup (`frontend/src/components/GalleryPanel.vue`):
```
.gallery-wrap            <- no explicit height, just a CSS Grid item
  .gallery-selection-bar   (position: absolute — out of flow)
  #emptyGallery            (position: absolute — out of flow)
  #gallery                 (height: 100%; display: flex; overflow-x: auto)
    .photo (each card)     (height: 100%)
      .photo-open           (height: 100%)
        .thumb                (height: 100%)
          img                   (height: 100%; object-fit: cover)
```

Relevant CSS (`frontend/src/style.css`):
- `.camera-panel { display: grid; grid-template-rows: minmax(0, 1fr) var(--gallery-height); }` (base rule, and re-declared for the overlay-drawer layout at ~line 975) — `.gallery-wrap` is the 2nd grid row, sized to a fixed `var(--gallery-height)` (e.g. 160px).
- `.gallery-wrap { ... }` (line 196 and again at line 1002) — **never sets `height` anywhere**. It only gets its box size because it's a grid item and grid items default to `align-self: stretch`, which stretches the box to fill the 160px row track.
- `#gallery { height: 100%; ... }` — a real child depending on `.gallery-wrap`'s height being resolvable as a percentage base.
- `.photo`, `.photo-open`, `.thumb`, `img` — each depends on `height: 100%` of its parent, all the way down to the `<img>`.

Root cause: `.gallery-wrap`'s height comes purely from **implicit grid-item stretch alignment**, not from an explicit `height` declaration. Per the CSS sizing spec, a stretched grid item's used height is fine for painting the box itself (which is why the gray gallery band in the Windows screenshot still shows at full height with correct background/border) — but whether that stretched height counts as a *definite height* for **descendants** resolving `height: 100%` is exactly the kind of edge case where WebKit and Chromium have historically disagreed (Chromium has had long-standing bugs where a grid item's auto/stretched height is not treated as definite for percentage resolution of its children).

That matches the screenshots precisely:
- `.gallery-wrap` itself renders at the correct height on both platforms (grid stretch applied to the box is unambiguous).
- On macOS/WebKit, `#gallery`'s `height: 100%` resolves against that stretched height, so `.photo` → `.thumb` → `img` all resolve `height: 100%` down the chain and thumbnails render as full-size rectangles.
- On Windows/Chromium, `#gallery`'s `height: 100%` fails to resolve against `.gallery-wrap`'s implicit height, so `#gallery` collapses to its content's intrinsic height. That collapse cascades to `.photo`/`.thumb`/`img`, which end up rendered at near-zero height — the thin horizontal lines seen in the bug screenshot (just the card border/background peeking through).

This same structure is used in both the desktop and the mobile/portrait media-query layout (~line 1088+), so the bug is not limited to one breakpoint.

## Plan to fix

Give `.gallery-wrap` an **explicit** `height: 100%` (in addition to being a grid item), so its height is unambiguously definite for descendant percentage resolution on every engine, instead of relying on implicit stretch alignment:

- `frontend/src/style.css`, `.gallery-wrap` rule (line ~196): add `height: 100%;`.
- No other files need to change — `#gallery`, `.photo`, `.thumb`, and `img` already declare `height: 100%` explicitly, so once `.gallery-wrap` itself has a definite height, the rest of the chain should resolve correctly on both WebKit and Chromium.
- Verify the fix doesn't affect the `main.gallery-closed .gallery-wrap` collapse animation (it sets `transform`/`opacity`/`padding: 0`, not `height`, so it should be unaffected).

This is a minimal, low-risk one-line CSS change. Since I can't run/screenshot the Windows build from here, verification will need to happen on an actual Windows build (or via a Chromium-based browser dev build with the same HTML/CSS) after the fix lands.

Once I read through and understand I will approve, later you can fix.

## Status: Fixed

Applied: added `height: 100%;` to the `.gallery-wrap` rule in `frontend/src/style.css` (line ~196). `npm run build` (vue-tsc + vite build) passes cleanly. Could not reproduce/verify on an actual Windows build from here — please confirm thumbnails render as full rectangles (not thin lines) on Windows after this change.