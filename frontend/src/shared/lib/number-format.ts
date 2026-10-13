export const NBSP = ' '

const formatters = new Map<string, Intl.NumberFormat>()

export function formatNumber(value: number, fractionDigits: number, locale?: string): string {
  const key = `${locale ?? ''}:${fractionDigits}`
  let formatter = formatters.get(key)
  if (!formatter) {
    formatter = new Intl.NumberFormat(locale, { minimumFractionDigits: fractionDigits, maximumFractionDigits: fractionDigits })
    formatters.set(key, formatter)
  }
  return formatter.format(value)
}
