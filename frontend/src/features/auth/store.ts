import { create } from 'zustand'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import type { TokenInfo, TokenPart } from '@/platform/mappers/oauth'
import type { OAuth2Auth, TokenUse } from '@/platform/mappers/request'
import { clearToken, copyToken, getNewToken, getToken, chooseToken, refreshToken, revealToken, type PendingToken } from '@/platform/services/oauth'

/** One auth owner's token as the UI may know it: never the value, except while the user reveals it. */
export interface TokenState {
  info: TokenInfo | null
  /** Waiting for the system browser sign-in; Cancel stops it. */
  waiting: boolean
  busy: boolean
  revealed: string | null
  copied: boolean
  error: AppError | null
}

interface AuthState {
  tokens: Record<string, TokenState>
  load: (ownerId: string) => Promise<void>
  getNew: (ownerId: string, oauth2: OAuth2Auth) => Promise<void>
  cancel: (ownerId: string) => void
  refresh: (ownerId: string, oauth2: OAuth2Auth) => Promise<void>
  use: (ownerId: string, which: TokenUse) => Promise<void>
  clear: (ownerId: string) => Promise<void>
  toggleReveal: (ownerId: string) => Promise<void>
  copy: (ownerId: string) => Promise<void>
  reset: () => void
}

const pending = new Map<string, PendingToken>()

export function emptyTokenState(): TokenState {
  return { info: null, waiting: false, busy: false, revealed: null, copied: false, error: null }
}

function partInUse(token: TokenState | undefined): TokenPart {
  return token?.info?.useToken === 'id_token' ? 'id' : 'access'
}

function activeLocalId(): string | null {
  return useWorkspacesStore.getState().activeLocalId
}

export const useAuthStore = create<AuthState>()((set, get) => {
  function patch(ownerId: string, update: Partial<TokenState>) {
    set((state) => ({ tokens: { ...state.tokens, [ownerId]: { ...(state.tokens[ownerId] ?? emptyTokenState()), ...update } } }))
  }

  async function run(ownerId: string, action: (localId: string) => Promise<Partial<TokenState>>) {
    const localId = activeLocalId()
    if (!localId) {
      return
    }
    patch(ownerId, { busy: true, error: null, copied: false })
    try {
      patch(ownerId, { ...(await action(localId)), busy: false })
    } catch (error) {
      patch(ownerId, { busy: false, error: toAppError(error) })
    }
  }

  return {
    tokens: {},

    load: (ownerId) => run(ownerId, async (localId) => ({ info: await getToken(localId, ownerId), revealed: null })),

    getNew: async (ownerId, oauth2) => {
      const localId = activeLocalId()
      if (!localId || pending.has(ownerId)) {
        return
      }
      const request = getNewToken(localId, ownerId, oauth2)
      pending.set(ownerId, request)
      patch(ownerId, { waiting: true, error: null, revealed: null, copied: false })
      try {
        patch(ownerId, { info: await request.result, waiting: false })
      } catch (error) {
        const appError = toAppError(error)
        patch(ownerId, { waiting: false, error: appError.cancelled || appError.code === 'OAU-08499' ? null : appError })
      } finally {
        pending.delete(ownerId)
      }
    },

    cancel: (ownerId) => pending.get(ownerId)?.cancel(),

    refresh: (ownerId, oauth2) => run(ownerId, async (localId) => ({ info: await refreshToken(localId, ownerId, oauth2), revealed: null })),

    use: (ownerId, which) => run(ownerId, async (localId) => ({ info: await chooseToken(localId, ownerId, which), revealed: null })),

    clear: (ownerId) =>
      run(ownerId, async (localId) => {
        await clearToken(localId, ownerId)
        return { info: null, revealed: null }
      }),

    toggleReveal: async (ownerId) => {
      if ((get().tokens[ownerId]?.revealed ?? null) !== null) {
        patch(ownerId, { revealed: null })
        return
      }
      await run(ownerId, async (localId) => ({ revealed: await revealToken(localId, ownerId, partInUse(get().tokens[ownerId])) }))
    },

    copy: (ownerId) =>
      run(ownerId, async (localId) => {
        await copyToken(localId, ownerId, partInUse(get().tokens[ownerId]))
        return { copied: true }
      }),

    reset: () => {
      for (const request of pending.values()) {
        request.cancel()
      }
      pending.clear()
      set({ tokens: {} })
    },
  }
})
