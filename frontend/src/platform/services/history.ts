import * as HistoryService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/historyservice'
import { AppError, toAppError } from '../errors'
import {
  toHistoryEntry,
  toHistoryEntryDetail,
  type HistoryEntry,
  type HistoryEntryDetail,
  type HistoryQuery,
  type HistorySourceFilter,
} from '../mappers/history'

const EMPTY_RESPONSE = 'UNK-00204'

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

export function listHistory(localId: string, { query = '', limit = 0, beforeId, source = 'all' }: HistoryQuery = {}): Promise<HistoryEntry[]> {
  return call(async () => {
    const entries = await HistoryService.List({ local_id: localId, source, query, limit, ...(beforeId ? { before_id: beforeId } : {}) })
    return (entries ?? []).map(toHistoryEntry)
  })
}

export function getHistoryEntry(localId: string, id: string): Promise<HistoryEntryDetail> {
  return call(async () => {
    const detail = await HistoryService.Get({ local_id: localId, id })
    if (detail === null) {
      throw new AppError(EMPTY_RESPONSE, 'The backend returned no data')
    }
    return toHistoryEntryDetail(detail)
  })
}

/** Clears only the entries of `source` (`all` clears everything). */
export function clearHistory(localId: string, source: HistorySourceFilter = 'all'): Promise<void> {
  return call(() => HistoryService.Clear({ local_id: localId, source }))
}
