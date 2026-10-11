import { create } from 'zustand'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import type { CookieDomain, CookieKey } from '@/platform/mappers/cookie'
import {
  clearCookies,
  deleteCookie,
  deleteCookieDomain,
  getCookieAllowlist,
  listCookies,
  saveCookieAllowlist,
  upsertCookie,
} from '@/platform/services/cookies'

export type CookiesView = 'cookies' | 'allowlist'

interface CookiesState {
  open: boolean
  view: CookiesView
  loading: boolean
  domains: CookieDomain[]
  /** Domains added in the dialog that hold no cookie yet; they are kept until the dialog closes. */
  addedDomains: string[]
  selectedDomain: string | null
  allowlist: string[]
  error: AppError | null
  openDialog: () => Promise<void>
  close: () => void
  setView: (view: CookiesView) => void
  load: () => Promise<void>
  selectDomain: (domain: string) => void
  addDomain: (domain: string) => void
  saveCookie: (domain: string, raw: string, replaces?: CookieKey) => Promise<AppError | null>
  removeCookie: (cookie: CookieKey) => Promise<void>
  removeDomain: (domain: string) => Promise<void>
  clearAll: () => Promise<void>
  saveAllowlist: (domains: string[]) => Promise<AppError | null>
  /** The jar of `localId` changed elsewhere (a send, a script); reloads when the dialog shows it. */
  jarChanged: (localId: string) => void
  reset: () => void
}

function activeLocalId(): string | null {
  return useWorkspacesStore.getState().activeLocalId
}

function normalizedDomain(domain: string): string {
  return domain.trim().toLowerCase().replace(/^\.+/, '')
}

const initialState = {
  open: false,
  view: 'cookies' as CookiesView,
  loading: false,
  domains: [] as CookieDomain[],
  addedDomains: [] as string[],
  selectedDomain: null as string | null,
  allowlist: [] as string[],
  error: null as AppError | null,
}

export const useCookiesStore = create<CookiesState>()((set, get) => {
  async function attempt(operation: (localId: string) => Promise<void>): Promise<AppError | null> {
    const localId = activeLocalId()
    if (!localId) {
      return null
    }
    try {
      await operation(localId)
      return null
    } catch (error) {
      return toAppError(error)
    }
  }

  async function change(operation: (localId: string) => Promise<void>) {
    const failure = await attempt(async (localId) => {
      await operation(localId)
      await refresh(localId)
    })
    set({ error: failure })
  }

  async function refresh(localId: string) {
    const [domains, allowlist] = await Promise.all([listCookies(localId), getCookieAllowlist(localId)])
    if (activeLocalId() !== localId) {
      return
    }
    set((state) => {
      const held = new Set(domains.map((group) => group.domain))
      const selectedDomain = state.selectedDomain ?? domains[0]?.domain ?? null
      return { domains, allowlist, addedDomains: state.addedDomains.filter((domain) => !held.has(domain)), selectedDomain }
    })
  }

  return {
    ...initialState,

    openDialog: async () => {
      set({ ...initialState, open: true, loading: true })
      await get().load()
    },

    close: () => set(initialState),

    setView: (view) => set({ view }),

    load: async () => {
      set({ loading: true })
      const failure = await attempt(refresh)
      set({ loading: false, error: failure })
    },

    selectDomain: (domain) => set({ selectedDomain: domain }),

    addDomain: (domain) => {
      const added = normalizedDomain(domain)
      if (added === '') {
        return
      }
      set((state) => ({
        addedDomains:
          state.addedDomains.includes(added) || state.domains.some((group) => group.domain === added) ? state.addedDomains : [...state.addedDomains, added],
        selectedDomain: added,
      }))
    },

    saveCookie: async (domain, raw, replaces) => {
      const failure = await attempt(async (localId) => {
        await upsertCookie(localId, domain, raw, replaces)
        await refresh(localId)
      })
      return failure
    },

    removeCookie: (cookie) => change((localId) => deleteCookie(localId, cookie)),

    removeDomain: async (domain) => {
      set((state) => ({
        addedDomains: state.addedDomains.filter((added) => added !== domain),
        selectedDomain: state.selectedDomain === domain ? null : state.selectedDomain,
      }))
      if (get().domains.some((group) => group.domain === domain)) {
        await change((localId) => deleteCookieDomain(localId, domain))
      }
    },

    clearAll: async () => {
      set({ addedDomains: [], selectedDomain: null })
      await change(clearCookies)
    },

    saveAllowlist: async (domains) =>
      attempt(async (localId) => {
        const saved = await saveCookieAllowlist(localId, domains)
        set({ allowlist: saved })
      }),

    jarChanged: (localId) => {
      if (get().open && activeLocalId() === localId) {
        void attempt(refresh)
      }
    },

    reset: () => set(initialState),
  }
})
