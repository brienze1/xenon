import type { UpdateStatus as UpdateStatusDto } from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

/** Where the auto-update stands (Go owns the state machine). */
export type UpdateState =
  | 'idle'
  | 'checking'
  | 'up_to_date'
  | 'available'
  | 'no_package'
  | 'check_failed'
  | 'downloading'
  | 'ready'
  | 'installing'
  | 'failed'

/** Why a check or an update failed; null when nothing failed. */
export type UpdateFailureReason = 'unreachable' | 'download_failed' | 'unverified' | 'not_packaged' | 'not_writable' | 'install_failed'

/** The update status Go returns from every UpdateService call and pushes on update:status. */
export interface UpdateStatus {
  state: UpdateState
  reason: UpdateFailureReason | null
  currentVersion: string
  latestVersion: string
  releaseUrl: string
  /** The folder Xenon can't write to; set only with `not_writable`. */
  folder: string
  /** Bytes downloaded so far, and the package size (0 when unknown). */
  received: number
  total: number
  /** True when the startup check started the download. */
  automatic: boolean
}

export function idleUpdateStatus(): UpdateStatus {
  return { state: 'idle', reason: null, currentVersion: '', latestVersion: '', releaseUrl: '', folder: '', received: 0, total: 0, automatic: false }
}

export function toUpdateStatus(dto: UpdateStatusDto | null): UpdateStatus {
  if (!dto) {
    return idleUpdateStatus()
  }
  return {
    state: (dto.state || 'idle') as UpdateState,
    reason: dto.reason ? (dto.reason as UpdateFailureReason) : null,
    currentVersion: dto.current_version ?? '',
    latestVersion: dto.latest_version ?? '',
    releaseUrl: dto.release_url ?? '',
    folder: dto.folder ?? '',
    received: dto.received ?? 0,
    total: dto.total ?? 0,
    automatic: dto.automatic ?? false,
  }
}
