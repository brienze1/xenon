import { isImage, languageFor, mediaType } from '@/shared/lib/body-format'
import type { CodeLanguage } from '@/shared/ui/code-editor/languages'

export { RAW_FIRST_BYTES, initialBodyView, isHtml, isImage, languageFor, prettyBody, type BodyView } from '@/shared/lib/body-format'

export function suggestedFileName(contentType: string): string {
  const extensions: Record<CodeLanguage, string> = { json: 'json', xml: 'xml', html: 'html', javascript: 'js', text: 'txt' }
  if (isImage(contentType)) {
    return `response.${mediaType(contentType).slice('image/'.length).replace('svg+xml', 'svg')}`
  }
  const language = languageFor(contentType)
  return language === 'text' && !mediaType(contentType).startsWith('text/') ? 'response.bin' : `response.${extensions[language]}`
}
