import type { BulkTable } from '@/features/request-builder/model'
import type { KeyValueRow, RequestDraft } from '@/platform/mappers/request'
import { keepRowIds } from '@/shared/lib/rows'

export function withSavedRows(current: RequestDraft, saved: RequestDraft): RequestDraft {
  return {
    ...saved,
    params: keepRowIds(current.params, saved.params),
    pathVariables: keepRowIds(current.pathVariables, saved.pathVariables),
    headers: keepRowIds(current.headers, saved.headers),
    body: {
      ...saved.body,
      formdata: keepRowIds(current.body.formdata, saved.body.formdata),
      urlencoded: keepRowIds(current.body.urlencoded, saved.body.urlencoded),
    },
  }
}

export function rowsOf(draft: RequestDraft, table: BulkTable): KeyValueRow[] {
  return table === 'urlencoded' ? draft.body.urlencoded : draft[table]
}

export function withRows(draft: RequestDraft, table: BulkTable, rows: KeyValueRow[]): RequestDraft {
  return table === 'urlencoded' ? { ...draft, body: { ...draft.body, urlencoded: rows } } : { ...draft, [table]: rows }
}
