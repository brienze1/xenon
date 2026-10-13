import { splitProtectedTerms } from '@/shared/lib/protected-terms'

interface NoTranslateTextProps {
  text: string
  terms: readonly string[]
}

export function NoTranslateText({ text, terms }: NoTranslateTextProps) {
  return splitProtectedTerms(text, terms).map((part, index) =>
    part.protected ? (
      <span key={index} translate="no">
        {part.text}
      </span>
    ) : (
      part.text
    ),
  )
}
