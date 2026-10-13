import { onViewZoomRequested, type ZoomDirection } from '@/platform/events'
import { detectPlatform } from '@/shared/hooks/use-hotkey'

/** Zoom steps, like a browser's. 1 is actual size. */
export const ZOOM_LEVELS: readonly number[] = [0.67, 0.75, 0.8, 0.9, 1, 1.1, 1.25, 1.5, 1.75, 2]
const SMALLEST = 0.67
const LARGEST = 2

// A per-machine display preference, not product state (frontend-architecture: localStorage only for these).
const STORAGE_KEY = 'xenon.zoom'

export function nextZoom(current: number, direction: ZoomDirection): number {
  if (direction === 'reset') {
    return 1
  }
  if (direction === 'in') {
    return ZOOM_LEVELS.find((level) => level > current + 0.001) ?? LARGEST
  }
  return [...ZOOM_LEVELS].reverse().find((level) => level < current - 0.001) ?? SMALLEST
}

function readSaved(): number {
  try {
    const saved = Number(window.localStorage.getItem(STORAGE_KEY))
    return ZOOM_LEVELS.some((level) => level === saved) ? saved : 1
  } catch {
    return 1
  }
}

function save(zoom: number) {
  try {
    window.localStorage.setItem(STORAGE_KEY, String(zoom))
  } catch {
    // Not remembered this time; the zoom still applies.
  }
}

/**
 * Zooms the whole UI the way a browser does: everything is drawn bigger and the layout is recomputed to fill the
 * same window, so no page-wide scrollbars appear. The body is scaled with CSS `zoom` and sized to the window divided
 * by the zoom, so after scaling it fits the window exactly. Dialogs and tooltips live in the body and scale too.
 */
export function applyZoom(zoom: number, body: HTMLElement = document.body) {
  body.style.setProperty('zoom', String(zoom))
  body.style.width = `${window.innerWidth / zoom}px`
  body.style.height = `${window.innerHeight / zoom}px`
}

function directionOf(event: KeyboardEvent): ZoomDirection | null {
  if (!event.ctrlKey || event.altKey || event.metaKey) {
    return null
  }
  if (event.key === '=' || event.key === '+') {
    return 'in'
  }
  if (event.key === '-' || event.key === '_') {
    return 'out'
  }
  return event.key === '0' ? 'reset' : null
}

/** Applies the saved zoom and follows View menu (macOS) and Ctrl+= / Ctrl+- / Ctrl+0 (Windows, Linux). */
export function installZoom(platform = detectPlatform()): () => void {
  let zoom = readSaved()
  const change = (direction: ZoomDirection) => {
    zoom = nextZoom(zoom, direction)
    save(zoom)
    applyZoom(zoom)
  }
  const onResize = () => applyZoom(zoom)
  const onKeyDown = (event: KeyboardEvent) => {
    const direction = directionOf(event)
    if (direction) {
      event.preventDefault()
      change(direction)
    }
  }
  applyZoom(zoom)
  window.addEventListener('resize', onResize)
  if (platform !== 'mac') {
    window.addEventListener('keydown', onKeyDown, true)
  }
  const unsubscribe = onViewZoomRequested(change)
  return () => {
    unsubscribe()
    window.removeEventListener('resize', onResize)
    window.removeEventListener('keydown', onKeyDown, true)
  }
}
