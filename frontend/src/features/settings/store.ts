import { create } from 'zustand'
import { toAppError, type AppError } from '@/platform/errors'
import type { AIAccessRuntime, AIAccessStatus } from '@/platform/mappers/ai-access'
import { defaultAppSettings, type AppSettings } from '@/platform/mappers/settings'
import { copyText, getAIAccessStatus } from '@/platform/services/ai-access'
import {
  AI_PORT_OUT_OF_RANGE,
  getSettings,
  HISTORY_AI_CAP_OUT_OF_RANGE,
  HISTORY_USER_CAP_OUT_OF_RANGE,
  saveSettings,
  SCRIPT_TIMEOUT_OUT_OF_RANGE,
} from '@/platform/services/settings'
import type { HostKeyUnknown } from '@/platform/events'
import type { Result } from '@/shared/types/result'
import type { GitAccount, GitHubDeviceCode, SshAccountInput, TokenAccountInput } from '@/platform/mappers/git-accounts'
import {
  addSshAccount,
  addTokenAccount,
  confirmHostKey,
  copyPublicKey,
  GITHUB_SIGN_IN_NOT_CONFIGURED,
  listGitAccounts,
  pollGitHubSignIn,
  removeGitAccount,
  startGitHubSignIn,
  updateGitAccount,
  uploadPublicKey,
} from '@/platform/services/git-accounts'
import { checkDraft, FIELD_MESSAGES, sameDraft, toDraft } from './lib/settings-form'
import type { SettingsDraft, SettingsField, SettingsFieldErrors } from './model'

/** The Go validation codes that belong to one form field (shown inline, not as a banner). */
const FIELD_CODES: Record<string, SettingsField> = {
  [SCRIPT_TIMEOUT_OUT_OF_RANGE]: 'scriptTimeout',
  [HISTORY_USER_CAP_OUT_OF_RANGE]: 'historyUserCap',
  [HISTORY_AI_CAP_OUT_OF_RANGE]: 'historyAiCap',
  [AI_PORT_OUT_OF_RANGE]: 'aiPort',
}

/** The validated field each draft value feeds; editing a value clears that field's error. */
const DRAFT_FIELDS: Partial<Record<keyof SettingsDraft, SettingsField>> = {
  scriptTimeoutSeconds: 'scriptTimeout',
  historyUserCap: 'historyUserCap',
  historyAiCap: 'historyAiCap',
  aiPort: 'aiPort',
}

const NO_FIELD_ERRORS: SettingsFieldErrors = {}

export type SettingsStatus = 'idle' | 'loading' | 'ready'

interface SettingsState {
  /** The settings as saved in app.json. */
  settings: AppSettings
  /** The Settings tab's form; equal to `settings` when nothing is edited. */
  draft: SettingsDraft
  status: SettingsStatus
  loadError: AppError | null
  fieldErrors: SettingsFieldErrors
  saving: boolean
  saveError: AppError | null
  /** The MCP server's status (Status at load, then ai:access-status-changed). */
  aiStatus: AIAccessStatus | null
  /** With no workspace open, Settings shows as a full main-area page instead of a tab. */
  pageOpen: boolean
  /** Reads the settings and the AI access status; keeps unsaved edits. */
  load: () => Promise<void>
  edit: (patch: Partial<SettingsDraft>) => void
  /**
   * Validates every field and saves them together. `true` when saved, `false` when a field is
   * invalid (its inline error is set), an error when Go could not save.
   */
  save: () => Promise<Result<boolean, AppError>>
  /** Forgets the unsaved edits. */
  discard: () => void
  applyStatus: (status: AIAccessRuntime) => void
  /** Copies a client config or the stdio command; true when the clipboard took it. */
  copy: (text: string) => Promise<boolean>
  openPage: () => void
  closePage: () => void
}

/** Whether the Settings form has unsaved edits (the tab's dirty dot). */
export function isSettingsDirty(state: Pick<SettingsState, 'settings' | 'draft'>): boolean {
  return !sameDraft(state.draft, toDraft(state.settings))
}

const EMPTY_STATUS: AIAccessStatus = { state: 'off', address: null, reason: null, code: null, clients: 0, stdioCommand: '', httpConfig: '', stdioConfig: '' }

export const useSettingsStore = create<SettingsState>((set, get) => ({
  settings: defaultAppSettings(),
  draft: toDraft(defaultAppSettings()),
  status: 'idle',
  loadError: null,
  fieldErrors: NO_FIELD_ERRORS,
  saving: false,
  saveError: null,
  aiStatus: null,
  pageOpen: false,
  load: async () => {
    set((state) => ({ status: state.status === 'ready' ? 'ready' : 'loading', loadError: null }))
    const [settings, aiStatus] = await Promise.allSettled([getSettings(), getAIAccessStatus()])
    if (aiStatus.status === 'fulfilled') {
      set({ aiStatus: aiStatus.value })
    }
    if (settings.status === 'rejected') {
      set({ status: 'ready', loadError: toAppError(settings.reason) })
      return
    }
    set((state) => ({
      status: 'ready',
      settings: settings.value,
      draft: isSettingsDirty(state) ? state.draft : toDraft(settings.value),
    }))
  },
  edit: (patch) =>
    set((state) => {
      const edited = new Set(Object.keys(patch).map((key) => DRAFT_FIELDS[key as keyof SettingsDraft]))
      const fieldErrors = Object.fromEntries(Object.entries(state.fieldErrors).filter(([field]) => !edited.has(field as SettingsField)))
      return { draft: { ...state.draft, ...patch }, fieldErrors, saveError: null }
    }),
  save: async () => {
    const checked = checkDraft(get().draft)
    if (!checked.ok) {
      set({ fieldErrors: checked.errors, saveError: null })
      return { ok: true, value: false }
    }
    set({ saving: true, fieldErrors: NO_FIELD_ERRORS, saveError: null })
    try {
      const saved = await saveSettings(checked.settings)
      set({ settings: saved, draft: toDraft(saved), saving: false })
      return { ok: true, value: true }
    } catch (error) {
      const failure = toAppError(error)
      const field = FIELD_CODES[failure.code]
      if (field) {
        set({ saving: false, fieldErrors: { [field]: FIELD_MESSAGES[field] } })
        return { ok: true, value: false }
      }
      set({ saving: false, saveError: failure })
      return { ok: false, error: failure }
    }
  },
  discard: () => set((state) => ({ draft: toDraft(state.settings), fieldErrors: NO_FIELD_ERRORS, saveError: null })),
  applyStatus: (status) => set((state) => ({ aiStatus: { ...(state.aiStatus ?? EMPTY_STATUS), ...status } })),
  copy: async (text) => {
    try {
      return await copyText(text)
    } catch {
      return false
    }
  },
  openPage: () => set({ pageOpen: true }),
  closePage: () => set({ pageOpen: false }),
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
  copyKey: (publicKey: string) => Promise<AppError | null>
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
  copyKey: (publicKey) => attempt(() => copyPublicKey(publicKey)),
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
