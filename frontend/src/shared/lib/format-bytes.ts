import { formatNumber, NBSP } from './number-format'

const UNITS = ['B', 'KB', 'MB', 'GB']

export function formatBytes(bytes: number, locale?: string): string {
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < UNITS.length - 1) {
    value /= 1024
    unit += 1
  }
  const digits = unit === 0 ? 0 : value < 10 ? 2 : 1
  return `${formatNumber(value, digits, locale)}${NBSP}${UNITS[unit]}`
}
