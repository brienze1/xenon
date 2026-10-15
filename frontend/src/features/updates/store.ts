import { create } from 'zustand'
import { idleUpdateStatus, type UpdateStatus } from '@/platform/mappers/update-status'
import { checkForUpdates, checkOnStartup, getUpdateStatus, openReleasePage, restartToUpdate, startUpdate } from '@/platform/services/updates'
import { useTabsStore } from '@/features/tabs/store'
import type { ReadyToastPhase } from './model'

interface UpdatesState {
  /** The latest status from Go (a call's result or an update:status event). */
  status: UpdateStatus
  toast: ReadyToastPhase
  /** Takes a status pushed by Go on update:status. */
  applyEvent: (status: UpdateStatus) => void
  load: () => Promise<void>
  /** The once-per-run startup check (Go also starts the automatic download from it). */
  checkOnStartup: () => Promise<void>
  check: () => Promise<void>
  /** Downloads and verifies the update, then restarts into it once it is ready. */
  update: () => Promise<void>
  /** Saves the open tabs' UI state, then installs the verified update and restarts. */
  restart: () => Promise<void>
  openReleasePage: () => Promise<void>
  closeToast: () => void
}

/** Counts update:status events, so a call's result never overwrites a newer pushed status. */
let eventCount = 0

export const useUpdatesStore = create<UpdatesState>((set, get) => {
  function apply(status: UpdateStatus) {
    set((state) => ({
      status,
      toast: state.toast === 'unseen' && status.state === 'ready' && status.automatic ? 'open' : state.toast,
    }))
  }

  /**
   * Runs one UpdateService call and applies its result unless Go pushed a status meanwhile (the
   * event stream is newer). Returns the result, or null when the call itself failed — update
   * failures are statuses, so a thrown error is an internal bug that leaves the status as it is.
   */
  async function run(operation: () => Promise<UpdateStatus>): Promise<UpdateStatus | null> {
    const before = eventCount
    try {
      const status = await operation()
      if (eventCount === before) {
        apply(status)
      }
      return status
    } catch {
      return null
    }
  }

  return {
    status: idleUpdateStatus(),
    toast: 'unseen',
    applyEvent: (status) => {
      eventCount += 1
      apply(status)
    },
    load: async () => {
      await run(getUpdateStatus)
    },
    checkOnStartup: async () => {
      await run(checkOnStartup)
    },
    check: async () => {
      await run(checkForUpdates)
    },
    update: async () => {
      const status = await run(startUpdate)
      if (status?.state === 'ready') {
        await get().restart()
      }
    },
    restart: async () => {
      set((state) => ({ toast: state.toast === 'open' ? 'closed' : state.toast }))
      await useTabsStore.getState().flushPersist()
      await run(restartToUpdate)
    },
    openReleasePage: async () => {
      await run(openReleasePage)
    },
    closeToast: () => set((state) => ({ toast: state.toast === 'open' ? 'closed' : state.toast })),
  }
})
