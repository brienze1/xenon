import * as SnippetService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/snippetservice'
import { AppError, toAppError } from '../errors'
import { toRequestDto, type RequestDraft } from '../mappers/request'
import { toSnippet, toSnippetLanguage, type Snippet, type SnippetLanguage } from '../mappers/snippet'

const EMPTY_RESPONSE = 'UNK-00204'

export interface SnippetInput {
  localId: string
  tabId: string
  draft: RequestDraft
  language: string
  options: Record<string, string>
  includeSecrets: boolean
}

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

function toGenerateSnippetDto(input: SnippetInput) {
  return {
    local_id: input.localId,
    tab_id: input.tabId,
    language: input.language,
    options: input.options,
    include_secrets: input.includeSecrets,
    request: toRequestDto(input.draft),
  }
}

export function listSnippetLanguages(): Promise<SnippetLanguage[]> {
  return call(async () => ((await SnippetService.Languages()) ?? []).map(toSnippetLanguage))
}

/** The request as code, generated in Go; secrets stay masked unless `includeSecrets`. */
export function generateSnippet(input: SnippetInput): Promise<Snippet> {
  return call(async () => {
    const snippet = await SnippetService.Generate(toGenerateSnippetDto(input))
    if (snippet === null) {
      throw new AppError(EMPTY_RESPONSE, 'The backend returned no data')
    }
    return toSnippet(snippet)
  })
}

/** Go puts the same text on the clipboard, so masked secrets never pass through the webview. */
export function copySnippet(input: SnippetInput): Promise<void> {
  return call(() => SnippetService.Copy(toGenerateSnippetDto(input)))
}
