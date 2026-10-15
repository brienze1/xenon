import type { UpdateFailureReason, UpdateStatus } from '@/platform/mappers/update-status'

export type UpdateAction = 'update' | 'check' | 'restart'

/** What the About & Updates section shows for one status (contracts/ui.md). */
export interface UpdateRow {
  /** The status line; null in `idle`. */
  message: string | null
  showLatest: boolean
  showWhatsNew: boolean
  /** The download progress, 0–100, while downloading. */
  progress: number | null
  /** `not_writable` links the release page instead of offering Update. */
  showReleasePageLink: boolean
  actions: readonly UpdateAction[]
  checkDisabled: boolean
}

const CHECK: readonly UpdateAction[] = ['check']
const RETRY: readonly UpdateAction[] = ['update', 'check']
const NONE: readonly UpdateAction[] = []

const FAILURES: Record<UpdateFailureReason, { message: (status: UpdateStatus) => string; actions: readonly UpdateAction[] }> = {
  unreachable: { message: () => "Couldn't check for updates", actions: CHECK },
  download_failed: { message: () => 'The update could not be downloaded', actions: RETRY },
  unverified: { message: () => 'The update could not be verified — nothing was changed', actions: RETRY },
  not_packaged: { message: () => 'Updates install only in a packaged Xenon', actions: CHECK },
  not_writable: {
    message: (status) =>
      `Xenon can't replace itself at ${status.folder}. Move Xenon to a folder you own, such as ~/Applications, or install it from the release page.`,
    actions: CHECK,
  },
  install_failed: { message: () => 'The update could not be installed — Xenon was not changed', actions: RETRY },
}

const BASE: UpdateRow = { message: null, showLatest: false, showWhatsNew: false, progress: null, showReleasePageLink: false, actions: CHECK, checkDisabled: false }

function percent(received: number, total: number): number {
  return total > 0 ? Math.min(100, Math.floor((received * 100) / total)) : 0
}

export function updateRow(status: UpdateStatus): UpdateRow {
  switch (status.state) {
    case 'idle':
      return BASE
    case 'checking':
      return { ...BASE, message: 'Checking…', checkDisabled: true }
    case 'up_to_date':
      return { ...BASE, message: 'Xenon is up to date', showLatest: true }
    case 'available':
      return { ...BASE, message: 'Update available', showLatest: true, showWhatsNew: true, actions: RETRY }
    case 'no_package':
      return { ...BASE, message: 'No update package for this system', showLatest: true, showWhatsNew: true }
    case 'check_failed':
      return { ...BASE, message: "Couldn't check for updates" }
    case 'downloading': {
      const progress = percent(status.received, status.total)
      return { ...BASE, message: `Downloading… ${progress}%`, progress, showLatest: true, showWhatsNew: true, actions: NONE }
    }
    case 'ready':
      return { ...BASE, message: 'Update ready', showLatest: true, showWhatsNew: true, actions: ['restart'] }
    case 'installing':
      return { ...BASE, message: 'Installing…', actions: NONE }
    case 'failed': {
      const failure = (status.reason ? FAILURES[status.reason] : undefined) ?? FAILURES.download_failed
      return { ...BASE, message: failure.message(status), actions: failure.actions, showReleasePageLink: status.reason === 'not_writable' }
    }
  }
}
