import * as UpdateService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/updateservice'
import { toAppError } from '../errors'
import { toUpdateStatus, type UpdateStatus } from '../mappers/update-status'

async function call(operation: () => Promise<Parameters<typeof toUpdateStatus>[0]>): Promise<UpdateStatus> {
  try {
    return toUpdateStatus(await operation())
  } catch (error) {
    throw toAppError(error)
  }
}

/** The current status; `idle` with the running version before any check. */
export function getUpdateStatus(): Promise<UpdateStatus> {
  return call(() => UpdateService.GetStatus())
}

/** Asks GitHub for the latest release; failures come back as a `check_failed` status. */
export function checkForUpdates(): Promise<UpdateStatus> {
  return call(() => UpdateService.Check())
}

/** The once-per-run startup check; starts the background download when Update automatically is on. */
export function checkOnStartup(): Promise<UpdateStatus> {
  return call(() => UpdateService.CheckOnStartup())
}

/** Downloads and verifies the update: `ready`, or the failure status. It never restarts. */
export function startUpdate(): Promise<UpdateStatus> {
  return call(() => UpdateService.Update())
}

/** Installs the verified update and restarts Xenon; flush the tabs' UI state first. */
export function restartToUpdate(): Promise<UpdateStatus> {
  return call(() => UpdateService.Restart())
}

/** Opens the latest release page in the system browser. */
export function openReleasePage(): Promise<UpdateStatus> {
  return call(() => UpdateService.OpenReleasePage())
}
