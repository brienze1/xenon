import type { AppError } from '@/platform/errors'

export interface ProblemPosition {
  line: number
  column: number
}

/** The line and column Go reported for a document problem, when it reported them. */
export function problemPosition(error: AppError | null): ProblemPosition | null {
  if (!error) {
    return null
  }
  const line = Number(error.details.find((detail) => detail.attribute === 'line')?.messages[0])
  const column = Number(error.details.find((detail) => detail.attribute === 'column')?.messages[0])
  if (!Number.isInteger(line) || line < 1) {
    return null
  }
  return { line, column: Number.isInteger(column) && column > 0 ? column : 1 }
}

/** The problem without its "line 4, column 4: " prefix, which the position already shows. */
export function problemMessage(error: AppError): string {
  return (error.description || error.code).replace(/^line \d+, column \d+: /, '')
}
