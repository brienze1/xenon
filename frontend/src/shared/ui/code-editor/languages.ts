import type { Extension } from '@codemirror/state'

export type CodeLanguage = 'json' | 'xml' | 'html' | 'javascript' | 'text'

export async function loadLanguage(language: CodeLanguage): Promise<Extension> {
  switch (language) {
    case 'json':
      return (await import('@codemirror/lang-json')).json()
    case 'xml':
      return (await import('@codemirror/lang-xml')).xml()
    case 'html':
      return (await import('@codemirror/lang-html')).html()
    case 'javascript':
      return (await import('@codemirror/lang-javascript')).javascript()
    default:
      return []
  }
}
