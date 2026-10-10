import { create } from 'zustand'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { AppError, toAppError } from '@/platform/errors'
import type { VariablesChanged } from '@/platform/events'
import type { Environment, EnvironmentColor, EnvironmentSummary } from '@/platform/mappers/environment'
import { environmentScope, type VariableEdit, type VariableTable } from '@/platform/mappers/variable'
import { emptyVaultStatus, type VaultStatus } from '@/platform/mappers/vault'
import {
  createEnvironment,
  deleteEnvironment,
  duplicateEnvironment,
  getEnvironment,
  listEnvironments,
  setActiveEnvironment,
  updateEnvironment,
} from '@/platform/services/environments'
import {
  addVariable,
  getVariableTable,
  persistVariables,
  resetVariables,
  saveVariableTable,
  setCurrentValue,
} from '@/platform/services/variables'
import {
  copySecret,
  getVaultStatus,
  PASSPHRASE_SETUP_REQUIRED,
  resetVault,
  revealSecret,
  setupPassphrase,
  unlockVault,
  VAULT_LOCKED,
} from '@/platform/services/vault'
import type { Result } from '@/shared/types/result'
import type { EnvironmentEditor, EnvironmentsStatus, PassphrasePrompt, TableState } from './model'

interface EnvironmentsState {
  environments: EnvironmentSummary[]
  activeId: string | null
  status: EnvironmentsStatus
  error: AppError | null
  switcherOpen: boolean
  editor: EnvironmentEditor | null
  tables: Record<string, TableState>
  vault: VaultStatus
  passphrasePrompt: PassphrasePrompt | null
  dismissedPrompts: PassphrasePrompt[]
  load: () => Promise<void>
  setActive: (id: string | null) => Promise<void>
  openSwitcher: () => void
  setSwitcherOpen: (open: boolean) => void
  clearError: () => void
  openEditor: (editor: EnvironmentEditor) => void
  closeEditor: () => void
  createEnvironment: (name: string) => Promise<Result<Environment, AppError>>
  renameEnvironment: (id: string, name: string) => Promise<AppError | null>
  updateEnvironmentLook: (id: string, look: EnvironmentLook) => Promise<AppError | null>
  duplicateEnvironment: (id: string) => Promise<Result<Environment, AppError>>
  deleteEnvironment: (id: string) => Promise<AppError | null>
  loadTable: (scope: string) => Promise<void>
  saveTable: (scope: string, edits: VariableEdit[], confirmReveal: boolean) => Promise<AppError | null>
  resetValues: (scope: string, keys: string[]) => Promise<AppError | null>
  persistValues: (scope: string, keys: string[]) => Promise<AppError | null>
  setCurrentValue: (scope: string, key: string, value: string) => Promise<AppError | null>
  addVariable: (scope: string, key: string, value: string, secret: boolean) => Promise<AppError | null>
  revealSecret: (scope: string, key: string) => Promise<Result<string, AppError>>
  copySecret: (scope: string, key: string) => Promise<AppError | null>
  handleVariablesChanged: (event: VariablesChanged) => void
  loadVault: () => Promise<void>
  applyVaultStatus: (status: VaultStatus) => void
  openPassphrase: (prompt: PassphrasePrompt) => void
  dismissPassphrase: () => void
  setupPassphrase: (passphrase: string, confirmation: string) => Promise<AppError | null>
  unlockVault: (passphrase: string) => Promise<AppError | null>
  resetVault: () => Promise<AppError | null>
  reset: () => void
}

const NO_WORKSPACE = 'WKS-00000'

/** How an environment warns before sending: its neon colour and "Confirm before send". */
export interface EnvironmentLook {
  color: EnvironmentColor
  confirmBeforeSend: boolean
}

/** Workspace-scoped state: cleared on every workspace switch. The vault status is app-wide. */
const INITIAL_STATE = {
  environments: [],
  activeId: null,
  status: 'idle',
  error: null,
  switcherOpen: false,
  editor: null,
  tables: {},
  passphrasePrompt: null,
} satisfies Partial<EnvironmentsState>

const EMPTY_TABLE: TableState = { rows: [], warnings: [], status: 'idle', error: null }

function activeLocalId(): string | null {
  return useWorkspacesStore.getState().activeLocalId
}

function noWorkspace(): AppError {
  return new AppError(NO_WORKSPACE, 'Open a workspace first')
}

function wantedPrompt(status: VaultStatus): PassphrasePrompt | null {
  if (status.needsPassphraseSetup) {
    return 'setup'
  }
  const localId = activeLocalId()
  return localId !== null && status.lockedLocalIds.includes(localId) ? 'unlock' : null
}

export const useEnvironmentsStore = create<EnvironmentsState>()((set, get) => {
  /** Runs an operation for the active workspace and drops its answer if the workspace changed meanwhile. */
  async function inWorkspace<T>(operation: (localId: string) => Promise<T>): Promise<Result<T, AppError>> {
    const localId = activeLocalId()
    if (localId === null) {
      return { ok: false, error: noWorkspace() }
    }
    try {
      const value = await operation(localId)
      if (activeLocalId() !== localId) {
        return { ok: false, error: noWorkspace() }
      }
      return { ok: true, value }
    } catch (error) {
      const appError = toAppError(error)
      promptForVault(appError)
      return { ok: false, error: appError }
    }
  }

  function promptForVault(error: AppError) {
    if (error.code === PASSPHRASE_SETUP_REQUIRED) {
      set({ passphrasePrompt: 'setup' })
    } else if (error.code === VAULT_LOCKED) {
      set({ passphrasePrompt: 'unlock' })
    }
  }

  function patchTable(scope: string, patch: Partial<TableState>) {
    set((state) => ({ tables: { ...state.tables, [scope]: { ...(state.tables[scope] ?? EMPTY_TABLE), ...patch } } }))
  }

  function storeTable(table: VariableTable) {
    patchTable(table.scope, { rows: table.rows, warnings: table.warnings, status: 'ready', error: null })
  }

  async function tableAction(scope: string, operation: (localId: string) => Promise<VariableTable>): Promise<AppError | null> {
    const result = await inWorkspace(operation)
    if (!result.ok) {
      return result.error.code === NO_WORKSPACE ? null : result.error
    }
    storeTable({ ...result.value, scope })
    return null
  }

  function refreshTables(scopes: string[]) {
    const loaded = Object.keys(get().tables)
    for (const scope of scopes.length === 0 ? loaded : loaded.filter((scope) => scopes.includes(scope))) {
      void get().loadTable(scope)
    }
  }

  async function vaultAction(operation: () => Promise<VaultStatus>): Promise<AppError | null> {
    try {
      const status = await operation()
      set({ passphrasePrompt: null, dismissedPrompts: [] })
      get().applyVaultStatus(status)
      refreshTables([])
      return null
    } catch (error) {
      return toAppError(error)
    }
  }

  return {
    ...INITIAL_STATE,
    vault: emptyVaultStatus(),
    dismissedPrompts: [],

    load: async () => {
      const localId = activeLocalId()
      if (localId === null) {
        return
      }
      set({ status: 'loading', error: null })
      try {
        const list = await listEnvironments(localId)
        if (activeLocalId() === localId) {
          set({ environments: list.environments, activeId: list.activeId, status: 'ready' })
        }
      } catch (error) {
        if (activeLocalId() === localId) {
          set({ status: 'ready', error: toAppError(error) })
        }
      }
    },

    setActive: async (id) => {
      const localId = activeLocalId()
      if (localId === null) {
        return
      }
      set({ error: null })
      try {
        const activeId = await setActiveEnvironment(localId, id)
        if (activeLocalId() === localId) {
          set({ activeId })
        }
      } catch (error) {
        set({ error: toAppError(error) })
      }
    },

    openSwitcher: () => set({ switcherOpen: true }),

    setSwitcherOpen: (open) => set({ switcherOpen: open }),

    clearError: () => set({ error: null }),

    openEditor: (editor) => set({ editor, switcherOpen: false }),

    closeEditor: () => set({ editor: null }),

    createEnvironment: async (name) => {
      const created = await inWorkspace((localId) => createEnvironment(localId, name))
      if (created.ok) {
        await get().load()
      }
      return created
    },

    renameEnvironment: async (id, name) => {
      const renamed = await inWorkspace(async (localId) => {
        const current = await getEnvironment(localId, id)
        return updateEnvironment(localId, {
          id,
          name,
          description: current.description,
          color: current.color,
          confirmBeforeSend: current.confirmBeforeSend,
        })
      })
      if (!renamed.ok) {
        return renamed.error
      }
      await get().load()
      return null
    },

    updateEnvironmentLook: async (id, look) => {
      const updated = await inWorkspace(async (localId) => {
        const current = await getEnvironment(localId, id)
        return updateEnvironment(localId, { id, name: current.name, description: current.description, ...look })
      })
      if (!updated.ok) {
        return updated.error
      }
      await get().load()
      return null
    },

    duplicateEnvironment: async (id) => {
      const duplicated = await inWorkspace((localId) => duplicateEnvironment(localId, id))
      if (duplicated.ok) {
        await get().load()
      }
      return duplicated
    },

    deleteEnvironment: async (id) => {
      const deleted = await inWorkspace((localId) => deleteEnvironment(localId, id))
      if (!deleted.ok) {
        return deleted.error
      }
      set((state) => {
        const tables = { ...state.tables }
        delete tables[environmentScope(id)]
        return { activeId: deleted.value, tables }
      })
      await get().load()
      return null
    },

    loadTable: async (scope) => {
      if (!get().tables[scope]) {
        patchTable(scope, { status: 'loading' })
      }
      const loaded = await inWorkspace((localId) => getVariableTable(localId, scope))
      if (loaded.ok) {
        storeTable({ ...loaded.value, scope })
      } else if (loaded.error.code !== NO_WORKSPACE) {
        patchTable(scope, { status: 'ready', error: loaded.error })
      }
    },

    saveTable: (scope, edits, confirmReveal) => tableAction(scope, (localId) => saveVariableTable(localId, scope, edits, confirmReveal)),

    resetValues: (scope, keys) => tableAction(scope, (localId) => resetVariables(localId, scope, keys)),

    persistValues: (scope, keys) => tableAction(scope, (localId) => persistVariables(localId, scope, keys)),

    setCurrentValue: async (scope, key, value) => {
      const saved = await inWorkspace((localId) => setCurrentValue(localId, scope, key, value))
      if (!saved.ok) {
        return saved.error.code === NO_WORKSPACE ? null : saved.error
      }
      refreshTables([scope])
      return null
    },

    addVariable: async (scope, key, value, secret) => {
      const added = await inWorkspace((localId) => addVariable(localId, scope, key, value, secret))
      if (!added.ok) {
        return added.error.code === NO_WORKSPACE ? null : added.error
      }
      refreshTables([scope])
      return null
    },

    revealSecret: (scope, key) => inWorkspace((localId) => revealSecret(localId, scope, key)),

    copySecret: async (scope, key) => {
      const copied = await inWorkspace((localId) => copySecret(localId, scope, key))
      return copied.ok ? null : copied.error
    },

    handleVariablesChanged: ({ localId, scopes }) => {
      if (localId !== activeLocalId()) {
        return
      }
      void get().load()
      refreshTables(scopes)
    },

    loadVault: async () => {
      try {
        get().applyVaultStatus(await getVaultStatus())
      } catch {
        // Without a status the vault UI stays hidden; secret actions still report their own errors.
        return
      }
    },

    applyVaultStatus: (status) => {
      const wanted = wantedPrompt(status)
      set((state) => ({
        vault: status,
        passphrasePrompt: wanted === null || state.dismissedPrompts.includes(wanted) ? state.passphrasePrompt : wanted,
      }))
    },

    openPassphrase: (prompt) => set({ passphrasePrompt: prompt }),

    dismissPassphrase: () =>
      set((state) => ({
        passphrasePrompt: null,
        dismissedPrompts: state.passphrasePrompt ? [...state.dismissedPrompts, state.passphrasePrompt] : state.dismissedPrompts,
      })),

    setupPassphrase: (passphrase, confirmation) => vaultAction(() => setupPassphrase(passphrase, confirmation)),

    unlockVault: (passphrase) => vaultAction(() => unlockVault(passphrase)),

    resetVault: async () => {
      const localId = activeLocalId()
      if (localId === null) {
        return null
      }
      return vaultAction(() => resetVault(localId))
    },

    reset: () => set(INITIAL_STATE),
  }
})
