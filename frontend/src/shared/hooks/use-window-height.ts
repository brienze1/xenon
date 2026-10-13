import { useSyncExternalStore } from 'react'

function subscribe(onChange: () => void): () => void {
  window.addEventListener('resize', onChange, { passive: true })
  return () => window.removeEventListener('resize', onChange)
}

function getWindowHeight(): number {
  return window.innerHeight
}

export function useWindowHeight(): number {
  return useSyncExternalStore(subscribe, getWindowHeight)
}
