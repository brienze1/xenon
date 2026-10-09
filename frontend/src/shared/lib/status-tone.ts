export type StatusTone = 'success' | 'redirect' | 'error' | 'info'

export function statusTone(statusCode: number): StatusTone {
  if (statusCode >= 200 && statusCode < 300) {
    return 'success'
  }
  if (statusCode >= 300 && statusCode < 400) {
    return 'redirect'
  }
  if (statusCode >= 400) {
    return 'error'
  }
  return 'info'
}
