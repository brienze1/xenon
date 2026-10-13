import type { AppError } from '@/platform/errors'
import type { Snippet } from '@/platform/mappers/snippet'

export const SIDE_PANEL_VIEWS = ['variables', 'code', 'info'] as const

export type SidePanelView = (typeof SIDE_PANEL_VIEWS)[number]

export type CodeSnippetStatus = 'idle' | 'loading' | 'ready' | 'error'

export interface CodeSnippetState {
  snippet: Snippet | null
  status: CodeSnippetStatus
  error: AppError | null
}

/** The language and the options to generate with. */
export interface SnippetRequest {
  language: string
  options: Readonly<Record<string, string>>
}
