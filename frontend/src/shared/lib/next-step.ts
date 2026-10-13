const SENTENCE_END = /[.!?…]$/

export function withNextStep(message: string, nextStep: string): string {
  const trimmed = message.trim()
  if (trimmed === '') {
    return nextStep
  }
  return `${trimmed}${SENTENCE_END.test(trimmed) ? '' : '.'} ${nextStep}`
}
