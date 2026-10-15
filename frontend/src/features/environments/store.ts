import { create } from 'zustand'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { AppError, toAppError } from '@/platform/errors'
import type { VariablesChanged } from '@/platform/events'
import type { Environment, EnvironmentSummary } from '@/platform/mappers/environment'
import { environmentScope, type VariableEdit, type VariableRow, type VariableTable } from '@/platform/mappers/variable'
import { emptyVaultStatus, type VaultStatus } from '@/platform/mappers/vault'
import {
  createEnvironment,
  deleteEnvironment,
  duplicateEnvironment,
  getEnvironment,
  listEnvironments,
  reorderEnvironment,
  setActiveEnvironment,
  updateEnvironment,
} from '@/platform/services/environments'
import {
  addVariable,
  CONFIRM_REVEAL_REQUIRED,
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
import { toDraftRows, toVariableEdits, typeSwitches, type DraftRow } from './lib/variable-draft'
import type { EnvironmentLook, EnvironmentsStatus, PassphrasePrompt, TableState, VariableDraft } from './model'

interface EnvironmentsState {
  environments: EnvironmentSummary[]
  activeId: string | null
  status: EnvironmentsStatus
  error: AppError | null
  switcherOpen: boolean
  tables: Record<string, TableState>
  /** The edits of every loaded variable table, kept here so an environment tab keeps them while hidden. */
  drafts: Record<string, VariableDraft>
  vault: VaultStatus
  passphrasePrompt: PassphrasePrompt | null
  dismissedPrompts: PassphrasePrompt[]
  load: () => Promise<void>
  setActive: (id: string | null) => Promise<void>
  openSwitcher: () => void
  setSwitcherOpen: (open: boolean) => void
  clearError: () => void
  createEnvironment: (name: string) => Promise<Result<Environment, AppError>>
  renameEnvironment: (id: string, name: string) => Promise<AppError | null>
  updateEnvironmentLook: (id: string, look: EnvironmentLook) => Promise<AppError | null>
  duplicateEnvironment: (id: string) => Promise<Result<Environment, AppError>>
  deleteEnvironment: (id: string) => Promise<AppError | null>
  /** Moves the environment to `index` (0 = first environment) at once, then saves; restores and returns the error on failure. */
  reorderEnvironment: (id: string, index: number) => Promise<AppError | null>
  loadTable: (scope: string) => Promise<void>
  editDraft: (scope: string, change: (rows: DraftRow[]) => DraftRow[]) => void
  /** Drops the unsaved edits of a table and shows its stored rows again. */
  discardDraft: (scope: string) => void
  /** Saves the edits of a table; `false` when a type switch must be confirmed first (`confirming`). */
  saveDraft: (scope: string) => Promise<Result<boolean, AppError>>
  confirmDraftSave: (scope: string) => Promise<Result<boolean, AppError>>
  cancelDraftSave: (scope: string) => void
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

/** Workspace-scoped state: cleared on every workspace switch. The vault status is app-wide. */
const INITIAL_STATE = {
  environments: [],
  activeId: null,
  status: 'idle',
  error: null,
  switcherOpen: false,
  tables: {},
  drafts: {},
  passphrasePrompt: null,
} satisfies Partial<EnvironmentsState>

const EMPTY_TABLE: TableState = { rows: [], warnings: [], status: 'idle', error: null }

function draftOf(rows: VariableRow[]): VariableDraft {
  return { base: rows, rows: toDraftRows(rows), dirty: false, confirming: null, failure: null, saving: false }
}

/** Whether the variable table of `scope` has unsaved edits. */
export function isScopeDirty(state: Pick<EnvironmentsState, 'drafts'>, scope: string): boolean {
  return state.drafts[scope]?.dirty === true
}

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

/** The list with environment `id` moved to `index` among the readable ones (broken ones stay last); null when nothing moves. */
function movedEnvironments(environments: EnvironmentSummary[], id: string, index: number): EnvironmentSummary[] | null {
  const readable = environments.filter((environment) => environment.error === null)
  const from = readable.findIndex((environment) => environment.id === id)
  const to = Math.min(index, readable.length - 1)
  if (from < 0 || from === to) {
    return null
  }
  const reordered = [...readable]
  const [moving] = reordered.splice(from, 1)
  reordered.splice(to, 0, moving!)
  return [...reordered, ...environments.filter((environment) => environment.error !== null)]
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
    if (!isScopeDirty(get(), table.scope)) {
      set((state) => ({ drafts: { ...state.drafts, [table.scope]: draftOf(table.rows) } }))
    }
  }

  function patchDraft(scope: string, patch: Partial<VariableDraft>) {
    set((state) => {
      const draft = state.drafts[scope]
      return draft ? { drafts: { ...state.drafts, [scope]: { ...draft, ...patch } } } : state
    })
  }

  function freshDraft(scope: string) {
    set((state) => ({ drafts: { ...state.drafts, [scope]: draftOf(state.tables[scope]?.rows ?? []) } }))
  }

  async function commitDraft(scope: string, confirmReveal: boolean): Promise<Result<boolean, AppError>> {
    const draft = get().drafts[scope]
    if (!draft) {
      return { ok: true, value: true }
    }
    const rows = draft.rows
    patchDraft(scope, { saving: true, failure: null })
    const error = await get().saveTable(scope, toVariableEdits(rows), confirmReveal)
    if (error?.code === CONFIRM_REVEAL_REQUIRED) {
      patchDraft(scope, { saving: false, confirming: { toSecret: [], toPlain: typeSwitches(rows).toPlain, reveal: true } })
      return { ok: true, value: false }
    }
    if (error !== null) {
      patchDraft(scope, { saving: false, failure: error })
      return { ok: false, error }
    }
    freshDraft(scope)
    return { ok: true, value: true }
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
        const drafts = { ...state.drafts }
        delete tables[environmentScope(id)]
        delete drafts[environmentScope(id)]
        return { activeId: deleted.value, tables, drafts }
      })
      await get().load()
      return null
    },

    reorderEnvironment: async (id, index) => {
      const previous = get().environments
      const moved = movedEnvironments(previous, id, index)
      if (moved === null) {
        return null
      }
      set({ environments: moved })
      const saved = await inWorkspace((localId) => reorderEnvironment(localId, id, index))
      if (!saved.ok) {
        set({ environments: previous })
        return saved.error
      }
      set({ environments: saved.value.environments, activeId: saved.value.activeId })
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

    editDraft: (scope, change) => {
      const draft = get().drafts[scope]
      if (draft) {
        patchDraft(scope, { rows: change(draft.rows), dirty: true })
      }
    },

    discardDraft: (scope) => {
      if (get().drafts[scope]) {
        freshDraft(scope)
      }
    },

    saveDraft: async (scope) => {
      const draft = get().drafts[scope]
      if (!draft?.dirty) {
        return { ok: true, value: true }
      }
      const switches = typeSwitches(draft.rows)
      if (switches.toSecret.length > 0 || switches.toPlain.length > 0) {
        patchDraft(scope, { confirming: switches, failure: null })
        return { ok: true, value: false }
      }
      return commitDraft(scope, false)
    },

    confirmDraftSave: (scope) => commitDraft(scope, get().drafts[scope]?.confirming?.reveal ?? false),

    cancelDraftSave: (scope) => patchDraft(scope, { confirming: null }),

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
