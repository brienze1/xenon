import * as ExportService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/exportservice'
import { AppError, toAppError } from '../errors'
import {
  toExportDto,
  toExportOptions,
  toExportSummary,
  type ExportFormat,
  type ExportItemRef,
  type ExportOptions,
  type ExportRequest,
  type ExportShape,
  type ExportSummary,
} from '../mappers/export'

const EMPTY_RESPONSE = 'UNK-00204'

export interface PendingExport {
  summary: Promise<ExportSummary>
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

export function getExportOptions(item: ExportItemRef): Promise<ExportOptions> {
  return call(async () => toExportOptions(required(await ExportService.Options({ local_id: item.localId, item_kind: item.kind, item_id: item.itemId }))))
}

/** Opens the native save dialog; null when it was cancelled. */
export function pickExportTarget(suggestedFileName: string, format: ExportFormat, shape: ExportShape): Promise<string | null> {
  return call(async () => {
    const picked = await ExportService.PickTarget({ suggested_file_name: suggestedFileName, format, shape })
    return picked?.path ? picked.path : null
  })
}

export function runExport(request: ExportRequest): PendingExport {
  const pending = ExportService.Export(toExportDto(request))
  return {
    summary: call(async () => toExportSummary(required(await pending))),
    cancel: () => pending.cancel(),
  }
}

export function showExportInFolder(path: string): Promise<void> {
  return call(() => ExportService.ShowInFolder({ path }))
}
