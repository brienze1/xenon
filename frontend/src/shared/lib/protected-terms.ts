export interface TextPart {
  text: string
  protected: boolean
}

const SPECIAL_CHARACTERS = /[.*+?^${}()|[\]\\]/g

function escapeTerm(term: string): string {
  return term.replace(SPECIAL_CHARACTERS, '\\$&')
}

export function splitProtectedTerms(text: string, terms: readonly string[]): TextPart[] {
  const usable = terms.filter((term) => term !== '')
  if (usable.length === 0 || text === '') {
    return text === '' ? [] : [{ text, protected: false }]
  }
  const protectedTerms = new Set(usable)
  const pattern = new RegExp(`(${[...usable].sort((a, b) => b.length - a.length).map(escapeTerm).join('|')})`)
  return text
    .split(pattern)
    .filter((part) => part !== '')
    .map((part) => ({ text: part, protected: protectedTerms.has(part) }))
}
