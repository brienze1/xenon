import { create } from 'zustand'
import { toAppError, type AppError } from '@/platform/errors'
import { defaultAppSettings, type AppSettings } from '@/platform/mappers/settings'
import { getSettings, saveSettings } from '@/platform/services/settings'
import type { HostKeyUnknown } from '@/platform/events'
import type { GitAccount, GitHubDeviceCode, SshAccountInput, TokenAccountInput } from '@/platform/mappers/git-accounts'
import {
  addSshAccount,
  addTokenAccount,
  confirmHostKey,
  GITHUB_SIGN_IN_NOT_CONFIGURED,
  listGitAccounts,
  pollGitHubSignIn,
  removeGitAccount,
  startGitHubSignIn,
  updateGitAccount,
  uploadPublicKey,
} from '@/platform/services/git-accounts'

interface SettingsState {
  open: boolean
  settings: AppSettings
  loading: boolean
  error: AppError | null
  openSettings: () => Promise<void>
  close: () => void
  save: (settings: AppSettings) => Promise<AppError | null>
}

export const useSettingsStore = create<SettingsState>((set) => ({
  open: false,
  settings: defaultAppSettings(),
  loading: false,
  error: null,
  openSettings: async () => {
    set({ open: true, loading: true, error: null })
    try {
      set({ settings: await getSettings(), loading: false })
    } catch (error) {
      set({ error: toAppError(error), loading: false })
    }
  },
  close: () => set({ open: false, error: null }),
  save: async (settings) => {
    try {
      set({ settings: await saveSettings(settings), open: false, error: null })
      return null
    } catch (error) {
      return toAppError(error)
    }
  },
}))

export type GitHubSignInPhase = 'idle' | 'starting' | 'waiting' | 'done' | 'expired' | 'denied' | 'not_configured' | 'error'

export interface GitHubSignInState {
  phase: GitHubSignInPhase
  code: GitHubDeviceCode | null
  interval: number
  error: AppError | null
}

const IDLE_SIGN_IN: GitHubSignInState = { phase: 'idle', code: null, interval: 0, error: null }

interface GitAccountsState {
  accounts: GitAccount[]
  error: AppError | null
  signIn: GitHubSignInState
  hostKeyPrompt: HostKeyUnknown | null
  load: () => Promise<void>
  addToken: (input: TokenAccountInput) => Promise<AppError | null>
  addSsh: (input: SshAccountInput) => Promise<{ publicKey: string; account: GitAccount } | AppError>
  update: (id: string, displayName: string, commitName: string, commitEmail: string) => Promise<AppError | null>
  remove: (id: string) => Promise<AppError | null>
  uploadKey: (sshAccountId: string, gitHubAccountId: string, title: string) => Promise<AppError | null>
  startSignIn: () => Promise<void>
  pollSignIn: () => Promise<void>
  resetSignIn: () => void
  promptHostKey: (prompt: HostKeyUnknown) => void
  answerHostKey: (accept: boolean) => Promise<AppError | null>
}

async function attempt(operation: () => Promise<unknown>): Promise<AppError | null> {
  try {
    await operation()
    return null
  } catch (error) {
    return toAppError(error)
  }
}

export const useGitAccountsStore = create<GitAccountsState>((set, get) => ({
  accounts: [],
  error: null,
  signIn: IDLE_SIGN_IN,
  hostKeyPrompt: null,
  load: async () => {
    try {
      set({ accounts: await listGitAccounts(), error: null })
    } catch (error) {
      set({ error: toAppError(error) })
    }
  },
  addToken: async (input) => {
    const failure = await attempt(() => addTokenAccount(input))
    if (!failure) {
      await get().load()
    }
    return failure
  },
  addSsh: async (input) => {
    try {
      const added = await addSshAccount(input)
      await get().load()
      return added
    } catch (error) {
      return toAppError(error)
    }
  },
  update: async (id, displayName, commitName, commitEmail) => {
    const failure = await attempt(() => updateGitAccount(id, displayName, commitName, commitEmail))
    if (!failure) {
      await get().load()
    }
    return failure
  },
  remove: async (id) => {
    const failure = await attempt(() => removeGitAccount(id))
    if (!failure) {
      set((state) => ({ accounts: state.accounts.filter((account) => account.id !== id) }))
    }
    return failure
  },
  uploadKey: (sshAccountId, gitHubAccountId, title) => attempt(() => uploadPublicKey(sshAccountId, gitHubAccountId, title)),
  startSignIn: async () => {
    set({ signIn: { ...IDLE_SIGN_IN, phase: 'starting' } })
    try {
      const code = await startGitHubSignIn()
      set({ signIn: { phase: 'waiting', code, interval: code.interval, error: null } })
    } catch (error) {
      const failure = toAppError(error)
      set({ signIn: { ...IDLE_SIGN_IN, phase: failure.code === GITHUB_SIGN_IN_NOT_CONFIGURED ? 'not_configured' : 'error', error: failure } })
    }
  },
  pollSignIn: async () => {
    const { code } = get().signIn
    if (!code || get().signIn.phase !== 'waiting') {
      return
    }
    try {
      const result = await pollGitHubSignIn(code.sessionId)
      set((state) => ({ signIn: { ...state.signIn, phase: result.status === 'pending' ? 'waiting' : result.status, interval: result.interval } }))
      if (result.status === 'done') {
        await get().load()
      }
    } catch (error) {
      set((state) => ({ signIn: { ...state.signIn, phase: 'error', error: toAppError(error) } }))
    }
  },
  resetSignIn: () => set({ signIn: IDLE_SIGN_IN }),
  promptHostKey: (prompt) => set({ hostKeyPrompt: prompt }),
  answerHostKey: async (accept) => {
    const prompt = get().hostKeyPrompt
    if (!prompt) {
      return null
    }
    const failure = await attempt(() => confirmHostKey(prompt.repository, prompt.fingerprint, accept))
    if (!failure) {
      set({ hostKeyPrompt: null })
    }
    return failure
  },
}))
