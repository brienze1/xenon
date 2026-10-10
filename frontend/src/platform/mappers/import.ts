import type {
  ImportClash as ImportClashDto,
  ImportCounts as ImportCountsDto,
  ImportPlan as ImportPlanDto,
  ImportPreview as ImportPreviewDto,
  ImportResult as ImportResultDto,
  ImportWarning as ImportWarningDto,
  PmUsage as PmUsageDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export type ClashResolution = 'replace' | 'keep_both' | 'skip'
export type ImportItemStatus = 'imported' | 'replaced' | 'skipped' | 'failed'

export interface ImportCounts {
  collections: number
  folders: number
  requests: number
  examples: number
  variables: number
  scripts: number
  environments: number
}

export interface ImportWarning {
  code: string
  severity: 'info' | 'warning'
  message: string
  itemKey: string
  itemId: string
  localId: string
  path: string
  location: string
}

export interface PmOccurrence {
  itemId: string
  localId: string
  path: string
  phase: string
  line: number
}

export interface PmUsage {
  api: string
  occurrences: PmOccurrence[]
}

export interface ImportPreviewItem {
  key: string
  kind: string
  name: string
  counts: ImportCounts
  warningsCount: number
}

export interface ImportPreview {
  sourceId: string
  kind: string
  version: string
  items: ImportPreviewItem[]
  warnings: ImportWarning[]
  pmUsages: PmUsage[]
}

export interface ImportPlanItem {
  key: string
  include: boolean
  targetLocalId: string
  targetParentId: string
  /** A new name for the imported item; absent or empty keeps the detected one. */
  name?: string
  resolution: ClashResolution | null
}

export interface ImportPlan {
  sourceId: string
  items: ImportPlanItem[]
  applyToAll: ClashResolution | null
}

export interface ImportClash {
  itemKey: string
  kind: string
  existingId: string
  existingName: string
  name: string
}

export interface ImportItemResult {
  key: string
  status: ImportItemStatus
  localId: string
  id: string
  name: string
}

export interface ImportWorkspaceResult {
  localId: string
  created: ImportCounts
  replaced: number
  skipped: number
}

export interface ImportFailure {
  itemKey: string
  name: string
  code: string
  message: string
}

export interface ImportResult {
  items: ImportItemResult[]
  workspaces: ImportWorkspaceResult[]
  warnings: ImportWarning[]
  pmUsages: PmUsage[]
  failed: ImportFailure[]
}

const STATUSES: ReadonlySet<string> = new Set<ImportItemStatus>(['imported', 'replaced', 'skipped', 'failed'])

function toCounts(dto: ImportCountsDto): ImportCounts {
  return {
    collections: dto.collections,
    folders: dto.folders,
    requests: dto.requests,
    examples: dto.examples,
    variables: dto.variables,
    scripts: dto.scripts,
    environments: dto.environments,
  }
}

function toWarning(dto: ImportWarningDto): ImportWarning {
  return {
    code: dto.code,
    severity: dto.severity === 'info' ? 'info' : 'warning',
    message: dto.message,
    itemKey: dto.item_key,
    itemId: dto.item_id,
    localId: dto.local_id,
    path: dto.path,
    location: dto.location,
  }
}

function toPmUsage(dto: PmUsageDto): PmUsage {
  return {
    api: dto.api,
    occurrences: (dto.occurrences ?? []).map((occurrence) => ({
      itemId: occurrence.item_id,
      localId: occurrence.local_id,
      path: occurrence.path,
      phase: occurrence.phase,
      line: occurrence.line,
    })),
  }
}

export function toImportPreview(dto: ImportPreviewDto): ImportPreview {
  return {
    sourceId: dto.source_id,
    kind: dto.kind,
    version: dto.version,
    items: (dto.items ?? []).map((item) => ({
      key: item.key,
      kind: item.kind,
      name: item.name,
      counts: toCounts(item.counts),
      warningsCount: item.warnings_count,
    })),
    warnings: (dto.warnings ?? []).map(toWarning),
    pmUsages: (dto.pm_usages ?? []).map(toPmUsage),
  }
}

export function toImportPlanDto(plan: ImportPlan): ImportPlanDto {
  return {
    source_id: plan.sourceId,
    apply_to_all: plan.applyToAll ?? '',
    items: plan.items.map((item) => ({
      key: item.key,
      include: item.include,
      target_local_id: item.targetLocalId,
      target_parent_id: item.targetParentId,
      name: item.name ?? '',
      resolution: item.resolution ?? '',
    })),
  }
}

export function toImportClash(dto: ImportClashDto): ImportClash {
  return { itemKey: dto.item_key, kind: dto.kind, existingId: dto.existing_id, existingName: dto.existing_name, name: dto.name }
}

export function toImportResult(dto: ImportResultDto): ImportResult {
  return {
    items: (dto.items ?? []).map((item) => ({
      key: item.key,
      status: STATUSES.has(item.status) ? (item.status as ImportItemStatus) : 'failed',
      localId: item.local_id,
      id: item.id,
      name: item.name,
    })),
    workspaces: (dto.workspaces ?? []).map((workspace) => ({
      localId: workspace.local_id,
      created: toCounts(workspace.created),
      replaced: workspace.replaced,
      skipped: workspace.skipped,
    })),
    warnings: (dto.warnings ?? []).map(toWarning),
    pmUsages: (dto.pm_usages ?? []).map(toPmUsage),
    failed: (dto.failed ?? []).map((failure) => ({ itemKey: failure.item_key, name: failure.name, code: failure.code, message: failure.message })),
  }
}
