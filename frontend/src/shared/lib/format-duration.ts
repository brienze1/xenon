export function formatDuration(milliseconds: number): string {
  if (milliseconds >= 1000) {
    return `${(milliseconds / 1000).toFixed(2)} s`
  }
  return `${Math.round(milliseconds)} ms`
}
