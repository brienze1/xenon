/**
 * The UI zoom: the body is scaled with CSS `zoom` (see app/zoom.ts). Pointer and getBoundingClientRect coordinates
 * are in screen pixels, while `left`/`top`/sizes inside the body are page pixels, so they are divided by the zoom.
 */
export function pageZoom(): number {
  const zoom = Number.parseFloat(document.body.style.zoom || getComputedStyle(document.body).zoom)
  return zoom > 0 ? zoom : 1
}

export function toPagePixels(screenPixels: number): number {
  return screenPixels / pageZoom()
}
