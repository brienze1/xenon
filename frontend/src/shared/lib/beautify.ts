import { html as beautifyHtml, js as beautifyJs } from 'js-beautify'
import type { CodeLanguage } from '@/shared/ui/code-editor/languages'
import { beautifyJson } from './beautify-json'

const INDENT_SIZE = 2
const VARIABLE = /\{\{[^{}]*\}\}/g
const PLACEHOLDER = /__xenonVar(\d+)__/g

// XML has no inline or void elements: every element gets its own line.
const XML_OPTIONS = {
  indent_size: INDENT_SIZE,
  inline: [],
  unformatted: [],
  content_unformatted: [],
  void_elements: [],
  indent_inner_html: true,
}
const HTML_OPTIONS = { indent_size: INDENT_SIZE }
const JS_OPTIONS = { indent_size: INDENT_SIZE }

/** Languages Beautify is offered for; plain text has nothing to re-indent. */
export function canBeautify(language: CodeLanguage): boolean {
  return language !== 'text'
}

/**
 * Re-indents a body in the given language. `{{variables}}` are kept exactly as typed.
 * Returns null when the text cannot be formatted (unbalanced JSON, or a language without a formatter).
 */
export function beautify(text: string, language: CodeLanguage): string | null {
  switch (language) {
    case 'json':
      return beautifyJson(text)
    case 'xml':
      return withVariablesKept(text, (protectedText) => beautifyHtml(protectedText, XML_OPTIONS))
    case 'html':
      return withVariablesKept(text, (protectedText) => beautifyHtml(protectedText, HTML_OPTIONS))
    case 'javascript':
      return withVariablesKept(text, (protectedText) => beautifyJs(protectedText, JS_OPTIONS))
    default:
      return null
  }
}

// Formatters would split `{{name}}` into nested blocks, so variables travel as identifiers and are put back afterwards.
function withVariablesKept(text: string, format: (text: string) => string): string {
  const variables: string[] = []
  const protectedText = text.replace(VARIABLE, (variable) => `__xenonVar${variables.push(variable) - 1}__`)
  return format(protectedText).replace(PLACEHOLDER, (placeholder, index: string) => variables[Number(index)] ?? placeholder)
}
