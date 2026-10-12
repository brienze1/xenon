import type { CodeLanguage } from '@/shared/ui/code-editor/languages'
import { beautify } from './beautify'

/** Bodies larger than this open in Raw first (PRD: large bodies). */
export const RAW_FIRST_BYTES = 5 * 1024 * 1024

export type BodyView = 'pretty' | 'raw' | 'preview'

export function initialBodyView(sizeBytes: number): BodyView {
  return sizeBytes > RAW_FIRST_BYTES ? 'raw' : 'pretty'
}

export function mediaType(contentType: string): string {
  return contentType.split(';')[0]?.trim().toLowerCase() ?? ''
}

export function languageFor(contentType: string): CodeLanguage {
  const type = mediaType(contentType)
  if (type.includes('json')) {
    return 'json'
  }
  if (type.includes('html')) {
    return 'html'
  }
  if (type.includes('xml')) {
    return 'xml'
  }
  if (type.includes('javascript')) {
    return 'javascript'
  }
  return 'text'
}

/** The Pretty view: the body re-indented with the same Beautify used for request bodies (big numbers and key order kept). */
export function prettyBody(body: string, language: CodeLanguage): string {
  return beautify(body, language) ?? body
}

export function isImage(contentType: string): boolean {
  return mediaType(contentType).startsWith('image/')
}

export function isHtml(contentType: string): boolean {
  return mediaType(contentType) === 'text/html'
}
