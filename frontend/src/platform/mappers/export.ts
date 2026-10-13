import type {
  Export as ExportDto,
  ExportCounts as ExportCountsDto,
  ExportOptions as ExportOptionsDto,
  ExportSummary as ExportSummaryDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export type ExportItemKind = 'workspace' | 'collection' | 'folder' | 'request' | 'environment' | 'globals'
export type ExportFormat = 'xenon_yaml' | 'postman'
export type ExportShape = 'zip' | 'bundle' | 'file'
export type ExportSecrets = 'included' | 'omitted' | 'names_only'

/** What to export: the workspace itself, or one of its items. */
export interface ExportItemRef {
  localId: string
  kind: ExportItemKind
  /** Empty for the workspace and its globals. */
  itemId: string
}

export interface ExportFormatOption {
  format: ExportFormat
  suggestedFileName: string
  shapes: ExportShape[]
  includeSecretsAllowed: boolean
}

export interface ExportOptions {
  suggestedFileName: string
  formats: ExportFormatOption[]
  hasSecrets: boolean
  hasCurrentValues: boolean
}

export interface ExportRequest extends ExportItemRef {
  format: ExportFormat
  shape: ExportShape
  includeSecrets: boolean
  /** Encrypts the included secrets; never stored or logged. */
  passphrase: string
  passphraseConfirmation: string
  currentValuesAsInitial: boolean
  targetPath: string
}

export interface ExportCounts {
  collections: number
  folders: number
  requests: number
  examples: number
  environments: number
  variables: number
}

export interface ExportDrop {
  itemPath: string
  field: string
  reason: string
}

export interface ExportWarning {
  itemPath: string
  code: string
  message: string
}

export interface ExportSummary {
  path: string
  secrets: ExportSecrets
  counts: ExportCounts
  dropped: ExportDrop[]
  warnings: ExportWarning[]
}

const FORMATS: ReadonlySet<string> = new Set<ExportFormat>(['xenon_yaml', 'postman'])
const SHAPES: ReadonlySet<string> = new Set<ExportShape>(['zip', 'bundle', 'file'])
const SECRETS: ReadonlySet<string> = new Set<ExportSecrets>(['included', 'omitted', 'names_only'])

function isFormat(value: string): value is ExportFormat {
  return FORMATS.has(value)
}

function isShape(value: string): value is ExportShape {
  return SHAPES.has(value)
}

export function toExportOptions(dto: ExportOptionsDto): ExportOptions {
  return {
    suggestedFileName: dto.suggested_file_name,
    hasSecrets: dto.has_secrets,
    hasCurrentValues: dto.has_current_values,
    formats: (dto.formats ?? []).flatMap((option) =>
      isFormat(option.format)
        ? [
            {
              format: option.format,
              suggestedFileName: option.suggested_file_name,
              shapes: (option.shapes ?? []).filter(isShape),
              includeSecretsAllowed: option.include_secrets_allowed,
            },
          ]
        : [],
    ),
  }
}

export function toExportDto(request: ExportRequest): ExportDto {
  return {
    local_id: request.localId,
    item_kind: request.kind,
    item_id: request.itemId,
    format: request.format,
    shape: request.shape,
    include_secrets: request.includeSecrets,
    passphrase: request.includeSecrets ? request.passphrase : '',
    passphrase_confirmation: request.includeSecrets ? request.passphraseConfirmation : '',
    current_values_as_initial: request.currentValuesAsInitial,
    target_path: request.targetPath,
  }
}

function toCounts(dto: ExportCountsDto): ExportCounts {
  return {
    collections: dto.collections,
    folders: dto.folders,
    requests: dto.requests,
    examples: dto.examples,
    environments: dto.environments,
    variables: dto.variables,
  }
}

export function toExportSummary(dto: ExportSummaryDto): ExportSummary {
  return {
    path: dto.path,
    secrets: SECRETS.has(dto.secrets) ? (dto.secrets as ExportSecrets) : 'omitted',
    counts: toCounts(dto.counts),
    dropped: (dto.dropped ?? []).map((drop) => ({ itemPath: drop.item_path, field: drop.field, reason: drop.reason })),
    warnings: (dto.warnings ?? []).map((warning) => ({ itemPath: warning.item_path, code: warning.code, message: warning.message })),
  }
}
