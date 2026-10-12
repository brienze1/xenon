import type {
  Snippet as SnippetDto,
  SnippetLanguage as SnippetLanguageDto,
  SnippetOption as SnippetOptionDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export type SnippetOptionType = 'boolean' | 'positiveInteger' | 'enum'

/** One of Postman's code-generator options; values travel as text ("true", "2", "single"). */
export interface SnippetOption {
  id: string
  label: string
  /** Postman's explanation of the option, shown under its label. */
  description: string
  type: SnippetOptionType
  defaultValue: string
  values: string[]
}

export interface SnippetLanguage {
  id: string
  label: string
  options: SnippetOption[]
}

export interface Snippet {
  text: string
  notices: string[]
}

function toOptionType(type: string): SnippetOptionType {
  return type === 'boolean' || type === 'positiveInteger' ? type : 'enum'
}

function toSnippetOption(dto: SnippetOptionDto): SnippetOption {
  return { id: dto.id, label: dto.label, description: dto.description ?? '', type: toOptionType(dto.type), defaultValue: String(dto.default ?? ''), values: dto.values ?? [] }
}

export function toSnippetLanguage(dto: SnippetLanguageDto): SnippetLanguage {
  return { id: dto.id, label: dto.label, options: (dto.options ?? []).map(toSnippetOption) }
}

export function toSnippet(dto: SnippetDto): Snippet {
  return { text: dto.text, notices: dto.notices ?? [] }
}
