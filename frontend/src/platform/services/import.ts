import * as ImportService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/importservice'
import { AppError, toAppError } from '../errors'
import {
  toImportClash,
  toImportPlanDto,
  toImportPreview,
  toImportResult,
  type ImportClash,
  type ImportPlan,
  type ImportPreview,
  type ImportResult,
} from '../mappers/import'

const EMPTY_RESPONSE = 'UNK-00204'
export const UNRESOLVED_CLASHES = 'IMP-04409'
export const UNKNOWN_SOURCE = 'IMP-05404'

export interface ImportInput {
  paths?: string[]
  text?: string
}

export interface PendingImport {
  result: Promise<ImportResult>
  cancel: () => void
}

function required<T>(value: T | null): T {
  if (value === null) {
    throw new AppError(EMPTY_RESPONSE, 'The backend returned no data')
  }
  return value
}

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

export function detectImport(input: ImportInput): Promise<ImportPreview> {
  return call(async () => toImportPreview(required(await ImportService.Detect({ paths: input.paths ?? [], text: input.text ?? '' }))))
}

export function pickImportSources(): Promise<string[]> {
  return call(async () => (await ImportService.PickSources())?.paths ?? [])
}

export function findImportClashes(plan: ImportPlan): Promise<ImportClash[]> {
  return call(async () => (required(await ImportService.Clashes(toImportPlanDto(plan))).clashes ?? []).map(toImportClash))
}

export function runImport(plan: ImportPlan): PendingImport {
  const pending = ImportService.Import(toImportPlanDto(plan))
  return {
    result: call(async () => toImportResult(required(await pending))),
    cancel: () => pending.cancel(),
  }
}

export function discardImport(sourceId: string): Promise<void> {
  return call(() => ImportService.Discard({ source_id: sourceId }))
}
