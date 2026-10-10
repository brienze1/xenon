import * as VariableService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/variableservice'
import { AppError, toAppError } from '../errors'
import { toVariableEditDto, toVariableRow, toVariableTable, type VariableEdit, type VariableRow, type VariableTable } from '../mappers/variable'

const EMPTY_RESPONSE = 'UNK-00204'
export const INVALID_VARIABLE_KEY = 'VAR-00400'
export const CONFIRM_REVEAL_REQUIRED = 'VAR-01409'
export const PERSIST_SECRET = 'VAR-02422'

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

export function getVariableTable(localId: string, scope: string): Promise<VariableTable> {
  return call(async () => toVariableTable(await VariableService.GetTable({ local_id: localId, scope }), scope))
}

/** Replaces the whole table: rows left out are removed. */
export function saveVariableTable(localId: string, scope: string, rows: VariableEdit[], confirmReveal: boolean): Promise<VariableTable> {
  return call(async () =>
    toVariableTable(
      await VariableService.SaveTable({ local_id: localId, scope, confirm_reveal: confirmReveal, rows: rows.map(toVariableEditDto) }),
      scope,
    ),
  )
}

/** Sets a row's local current value (a secret row's value goes to the vault); `""` clears it. */
export function setCurrentValue(localId: string, scope: string, key: string, value: string): Promise<VariableRow> {
  return call(async () => toVariableRow(required(await VariableService.SetCurrent({ local_id: localId, scope, key, value }))))
}

/** Resets the named rows (every row when `keys` is empty) back to their initial value. */
export function resetVariables(localId: string, scope: string, keys: string[]): Promise<VariableTable> {
  return call(async () => toVariableTable(await VariableService.Reset({ local_id: localId, scope, keys }), scope))
}

/** Persists the named plain rows' current values as initial values (every plain row when `keys` is empty). */
export function persistVariables(localId: string, scope: string, keys: string[]): Promise<VariableTable> {
  return call(async () => toVariableTable(await VariableService.Persist({ local_id: localId, scope, keys }), scope))
}
