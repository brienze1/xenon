import type { KeyValueRow, RequestDraft } from '@/platform/mappers/request'

const ROW_FIELD = /^(params|headers|body\.urlencoded|body\.formdata)\.(\d+)\.(key|value)$/
const PATH_VARIABLE_FIELD = /^path_variables\.(\d+)\.value$/

type RowList = 'params' | 'headers' | 'body.urlencoded' | 'body.formdata'

function rowsOf(draft: RequestDraft, list: RowList): KeyValueRow[] {
  switch (list) {
    case 'params':
      return draft.params
    case 'headers':
      return draft.headers
    case 'body.urlencoded':
      return draft.body.urlencoded
    default:
      return draft.body.formdata
  }
}

type RowsUpdate = <T extends KeyValueRow>(rows: T[]) => T[]

function withRowsOf(draft: RequestDraft, list: RowList, update: RowsUpdate): RequestDraft {
  switch (list) {
    case 'params':
      return { ...draft, params: update(draft.params) }
    case 'headers':
      return { ...draft, headers: update(draft.headers) }
    case 'body.urlencoded':
      return { ...draft, body: { ...draft.body, urlencoded: update(draft.body.urlencoded) } }
    default:
      return { ...draft, body: { ...draft.body, formdata: update(draft.body.formdata) } }
  }
}

/** The text of a request field named like the Inspect references (`url`, `headers.0.value`, …), or `null`. */
export function fieldText(draft: RequestDraft, field: string): string | null {
  if (field === 'url') {
    return draft.url
  }
  const row = ROW_FIELD.exec(field)
  if (row) {
    const found = rowsOf(draft, row[1] as RowList)[Number(row[2])]
    return found ? found[row[3] as 'key' | 'value'] : null
  }
  const variable = PATH_VARIABLE_FIELD.exec(field)
  return variable ? (draft.pathVariables[Number(variable[1])]?.value ?? null) : null
}

/** Returns the draft with the field's text replaced; unknown fields leave it unchanged. */
export function withFieldText(draft: RequestDraft, field: string, text: string): RequestDraft {
  if (field === 'url') {
    return { ...draft, url: text }
  }
  const row = ROW_FIELD.exec(field)
  if (row) {
    const index = Number(row[2])
    const part = row[3] as 'key' | 'value'
    return withRowsOf(draft, row[1] as RowList, (rows) => rows.map((candidate, position) => (position === index ? { ...candidate, [part]: text } : candidate)))
  }
  const variable = PATH_VARIABLE_FIELD.exec(field)
  if (variable) {
    const index = Number(variable[1])
    return { ...draft, pathVariables: draft.pathVariables.map((candidate, position) => (position === index ? { ...candidate, value: text } : candidate)) }
  }
  return draft
}
