import { formatNumber, NBSP } from './number-format'

export function formatDuration(milliseconds: number, locale?: string): string {
  if (milliseconds >= 1000) {
    return `${formatNumber(milliseconds / 1000, 2, locale)}${NBSP}s`
  }
  return `${formatNumber(Math.round(milliseconds), 0, locale)}${NBSP}ms`
}
