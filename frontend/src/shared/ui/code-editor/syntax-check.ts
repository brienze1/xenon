import type { Language } from '@codemirror/language'
import type { EditorDiagnostic } from './editor-extensions'
import type { CodeLanguage } from './languages'

const VARIABLE = /\{\{[^{}]*\}\}/g
const UNCLOSED_MARKUP = /^(MismatchedCloseTag|IncompleteCloseTag)$/

interface Problem {
  at: number
  message: string
}

async function importLanguage(language: CodeLanguage): Promise<Language | null> {
  switch (language) {
    case 'json':
      return (await import('@codemirror/lang-json')).jsonLanguage
    case 'xml':
      return (await import('@codemirror/lang-xml')).xmlLanguage
    case 'html':
      return (await import('@codemirror/lang-html')).htmlLanguage
    case 'javascript':
      return (await import('@codemirror/lang-javascript')).javascriptLanguage
    default:
      return null
  }
}

/**
 * Where a body stops being valid: the first syntax problem as a 1-based line and column with a short reason.
 * `{{variables}}` are valid anywhere, as in Postman. Plain text, and a grammar that fails to load, report nothing.
 */
export async function findSyntaxErrors(text: string, language: CodeLanguage): Promise<EditorDiagnostic[]> {
  if (text.trim() === '') {
    return []
  }
  let grammar: Language | null
  try {
    grammar = await importLanguage(language)
  } catch {
    return []
  }
  if (!grammar) {
    return []
  }
  const masked = maskVariables(text, language)
  const problem = firstProblem(grammar, masked, language)
  return problem ? [toDiagnostic(text, problem)] : []
}

// Same-length stand-ins keep every position the same: a number in JSON, an identifier elsewhere.
function maskVariables(text: string, language: CodeLanguage): string {
  return text.replace(VARIABLE, (variable) => (language === 'json' ? '1' + '0'.repeat(variable.length - 1) : 'x'.repeat(variable.length)))
}

function firstProblem(grammar: Language, text: string, language: CodeLanguage): Problem | null {
  let problem: Problem | null = null
  grammar.parser.parse(text).iterate({
    enter: (node) => {
      if (problem) {
        return false
      }
      if (UNCLOSED_MARKUP.test(node.name)) {
        problem = { at: node.from, message: `Closing tag ${text.slice(node.from, node.to).trim()} does not match the open element` }
      } else if (node.type.isError) {
        problem = describe(text, node.from, language)
      }
      return undefined
    },
  })
  return problem
}

function describe(text: string, at: number, language: CodeLanguage): Problem {
  const next = nextSignificant(text, at)
  if (next === -1) {
    return { at: Math.max(lastSignificant(text, text.length), 0), message: 'Unexpected end of body: something is not closed' }
  }
  const char = text[next] as string
  const previous = lastSignificant(text, next)
  const before = previous === -1 ? '' : (text[previous] as string)
  if (language === 'json' && before === ',' && (char === '}' || char === ']')) {
    return { at: previous, message: `Trailing comma before "${char}"` }
  }
  if (language === 'json' && previous !== -1 && !',:[{'.includes(before) && /["\d\-[{tfn]/.test(char)) {
    return { at: next, message: 'Missing comma before this value' }
  }
  return { at: next, message: `Unexpected "${char}"` }
}

function nextSignificant(text: string, from: number): number {
  for (let index = from; index < text.length; index++) {
    if (!/\s/.test(text[index] as string)) {
      return index
    }
  }
  return -1
}

function lastSignificant(text: string, before: number): number {
  for (let index = before - 1; index >= 0; index--) {
    if (!/\s/.test(text[index] as string)) {
      return index
    }
  }
  return -1
}

function toDiagnostic(text: string, problem: Problem): EditorDiagnostic {
  const upTo = text.slice(0, problem.at)
  const line = upTo.split('\n').length
  const column = problem.at - (upTo.lastIndexOf('\n') + 1) + 1
  return { line, column, message: problem.message }
}
