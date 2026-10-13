import * as ConsoleService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/consoleservice'
import { toAppError } from '../errors'
import { toConsoleEntry, type ConsoleEntry, type ConsoleQuery } from '../mappers/console'

/** Get fails with this code when the entry fell out of the 5,000-entry console. */
export const CONSOLE_ENTRY_GONE = 'CON-01404'

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

export function listConsole({ afterId = '', limit, filter = 'all', search = '' }: ConsoleQuery): Promise<ConsoleEntry[]> {
  return call(async () => ((await ConsoleService.List({ after_id: afterId, limit, filter, search })) ?? []).map(toConsoleEntry))
}

export function getConsoleEntry(id: string): Promise<ConsoleEntry> {
  return call(async () => toConsoleEntry(await ConsoleService.Get({ id })))
}

export function clearConsole(): Promise<void> {
  return call(() => ConsoleService.Clear())
}
