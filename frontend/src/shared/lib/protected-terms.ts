export interface TextPart {
  text: string
  protected: boolean
}

const SPECIAL_CHARACTERS = /[.*+?^${}()|[\]\\]/g

function escapeTerm(term: string): string {
  return term.replace(SPECIAL_CHARACTERS, '\\$&')
}

interface TermsMatcher {
  pattern: RegExp
  terms: ReadonlySet<string>
}

const MAX_CACHED_MATCHERS = 64

// One compiled pattern per distinct term list, built once and reused on every render.
const matchers = new Map<string, TermsMatcher>()

function matcherFor(usable: readonly string[]): TermsMatcher {
  const longestFirst = [...usable].sort((a, b) => b.length - a.length)
  const key = longestFirst.join('\u0000')
  const cached = matchers.get(key)
  if (cached) {
    return cached
  }
  if (matchers.size >= MAX_CACHED_MATCHERS) {
    matchers.clear()
  }
  const matcher = { pattern: new RegExp(`(${longestFirst.map(escapeTerm).join('|')})`), terms: new Set(usable) }
  matchers.set(key, matcher)
  return matcher
}

export function splitProtectedTerms(text: string, terms: readonly string[]): TextPart[] {
  if (text === '') {
    return []
  }
  const usable = terms.filter((term) => term !== '')
  if (usable.length === 0) {
    return [{ text, protected: false }]
  }
  const { pattern, terms: protectedTerms } = matcherFor(usable)
  return text.split(pattern).flatMap((part) => (part === '' ? [] : [{ text: part, protected: protectedTerms.has(part) }]))
}
