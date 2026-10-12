import type { Extension } from '@codemirror/state'

export type CodeLanguage = 'json' | 'xml' | 'html' | 'javascript' | 'yaml' | 'text'

const PLAIN_TEXT: Extension = []

async function importLanguage(language: CodeLanguage): Promise<Extension> {
  switch (language) {
    case 'json':
      return (await import('@codemirror/lang-json')).json()
    case 'xml':
      return (await import('@codemirror/lang-xml')).xml()
    case 'html':
      return (await import('@codemirror/lang-html')).html()
    case 'javascript':
      return (await import('@codemirror/lang-javascript')).javascript()
    case 'yaml':
      return (await import('@codemirror/lang-yaml')).yaml()
    default:
      return PLAIN_TEXT
  }
}

// Highlighting is optional: when a grammar chunk fails to load the editor shows plain text.
export async function loadLanguage(language: CodeLanguage): Promise<Extension> {
  try {
    return await importLanguage(language)
  } catch {
    return PLAIN_TEXT
  }
}
