const formatters = new Map<string, Intl.DateTimeFormat>()

export function formatDateTime(value: Date, options: Intl.DateTimeFormatOptions, locale?: string): string {
  const key = `${locale ?? ''}:${JSON.stringify(options)}`
  let formatter = formatters.get(key)
  if (!formatter) {
    formatter = new Intl.DateTimeFormat(locale, options)
    formatters.set(key, formatter)
  }
  return formatter.format(value)
}
