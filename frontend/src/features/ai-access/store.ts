import { create } from 'zustand'
import { toAppError, type AppError } from '@/platform/errors'
import type { PendingApproval } from '@/platform/mappers/ai-access'
import { answerApproval, APPROVAL_ALREADY_RESOLVED, APPROVAL_UNKNOWN, listPendingApprovals } from '@/platform/services/ai-access'

interface AIAccessState {
  /** AI sends waiting for Approve / Deny, oldest first; the prompt shows the first. */
  pending: PendingApproval[]
  /** The approval being answered, so its buttons disable. */
  answering: string | null
  error: AppError | null
  /** Reads the approvals already waiting (on start, before any event arrives). */
  load: () => Promise<void>
  add: (approval: PendingApproval) => void
  /** Drops an approval that was answered, timed out or cancelled. */
  resolve: (id: string) => void
  answer: (id: string, approved: boolean) => Promise<void>
}

function withApproval(pending: PendingApproval[], approval: PendingApproval): PendingApproval[] {
  return pending.some((known) => known.id === approval.id) ? pending : [...pending, approval]
}

export const useAIAccessStore = create<AIAccessState>((set, get) => ({
  pending: [],
  answering: null,
  error: null,
  load: async () => {
    try {
      const waiting = await listPendingApprovals()
      set((state) => ({ pending: waiting.reduce(withApproval, state.pending) }))
    } catch (error) {
      set({ error: toAppError(error) })
    }
  },
  add: (approval) => set((state) => ({ pending: withApproval(state.pending, approval) })),
  resolve: (id) =>
    set((state) => ({
      pending: state.pending.filter((approval) => approval.id !== id),
      answering: state.answering === id ? null : state.answering,
      error: state.answering === id ? null : state.error,
    })),
  answer: async (id, approved) => {
    set({ answering: id, error: null })
    try {
      await answerApproval(id, approved)
      get().resolve(id)
    } catch (error) {
      const failure = toAppError(error)
      if (failure.code === APPROVAL_ALREADY_RESOLVED || failure.code === APPROVAL_UNKNOWN) {
        get().resolve(id)
        return
      }
      set({ answering: null, error: failure })
    }
  },
}))
