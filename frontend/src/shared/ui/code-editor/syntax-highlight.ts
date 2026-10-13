import { HighlightStyle, syntaxHighlighting } from '@codemirror/language'
import type { Extension } from '@codemirror/state'
import { tags } from '@lezer/highlight'

// One palette for every body and script (request and response): keys, values and markup each get their own colour.
const highlightStyle = HighlightStyle.define([
  { tag: [tags.propertyName, tags.definition(tags.propertyName)], color: 'var(--syntax-key)' },
  { tag: [tags.string, tags.special(tags.string), tags.regexp], color: 'var(--syntax-string)' },
  { tag: [tags.number, tags.integer, tags.float], color: 'var(--syntax-number)' },
  { tag: [tags.bool, tags.null, tags.atom, tags.keyword, tags.controlKeyword, tags.definitionKeyword, tags.moduleKeyword, tags.self], color: 'var(--syntax-literal)' },
  { tag: [tags.punctuation, tags.bracket, tags.separator, tags.operator, tags.angleBracket], color: 'var(--syntax-punctuation)' },
  { tag: [tags.comment, tags.lineComment, tags.blockComment, tags.docComment], color: 'var(--syntax-comment)', fontStyle: 'italic' },
  { tag: [tags.tagName, tags.typeName, tags.className], color: 'var(--syntax-tag)' },
  { tag: [tags.attributeName, tags.processingInstruction, tags.documentMeta], color: 'var(--syntax-attribute)' },
  { tag: [tags.attributeValue], color: 'var(--syntax-string)' },
  { tag: [tags.function(tags.variableName), tags.function(tags.propertyName)], color: 'var(--syntax-function)' },
  { tag: tags.invalid, color: 'var(--neon-red)' },
])

export const syntaxColors: Extension = syntaxHighlighting(highlightStyle)
