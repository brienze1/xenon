import { useEffect, useState, useTransition } from 'react'
import type { AppError } from '@/platform/errors'
import { GIT_HISTORY_WARNING, type VariableRow } from '@/platform/mappers/variable'
import { CONFIRM_REVEAL_REQUIRED } from '@/platform/services/variables'
import { patchRow, removeRow } from '@/shared/lib/rows'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { withNextStep } from '@/shared/lib/next-step'
import { useEnvironmentsStore } from '@/features/environments/store'
import {
  keyProblems,
  newDraftRow,
  toDraftRows,
  toVariableEdits,
  typeSwitches,
  type DraftRow,
  type TypeSwitches,
} from '@/features/environments/lib/variable-draft'
import type { TableState } from '@/features/environments/model'
import { TypeSwitchConfirm } from './TypeSwitchConfirm'
import { VariableTableRow, type RowActions } from './VariableTableRow'
import styles from './VariableTable.module.css'

const SAVE_NEXT_STEP = 'Fix the highlighted rows, or check that the workspace folder is writable, then try again.'
const LOAD_NEXT_STEP = 'Check that the workspace folder exists, then try again.'

const TABLE_HEAD = (
  <thead>
    <tr>
      <th scope="col">Name</th>
      <th scope="col">Initial Value</th>
      <th scope="col">Current Value</th>
      <th scope="col" className={styles.typeColumn}>
        Type
      </th>
      <th scope="col" className={styles.enabledColumn}>
        Enabled
      </th>
      <th scope="col">Description</th>
      <th scope="col" className={styles.actionColumn}>
        <span className="visually-hidden">Actions</span>
      </th>
    </tr>
  </thead>
)

const GIT_HISTORY_MESSAGE = (
  <Banner.Message>
    Values that were in the workspace files may still be in git history. Rotate them if the folder was ever pushed.
  </Banner.Message>
)

const STALE_MESSAGE = <Banner.Message>This table changed outside this editor. Save to keep your edits, or discard them to see the new values.</Banner.Message>

const LOADING = (
  <p role="status" className={styles.note}>
    Loading…
  </p>
)

interface VariableTableProps {
  scope: string
  label: string
}

export function VariableTable({ scope, label }: VariableTableProps) {
  const table = useEnvironmentsStore((state) => state.tables[scope])
  const loadTable = useEnvironmentsStore((state) => state.loadTable)

  useEffect(() => {
    void loadTable(scope)
  }, [scope, loadTable])

  if (!table || (table.status !== 'ready' && table.rows.length === 0)) {
    return LOADING
  }
  if (table.error && table.rows.length === 0) {
    return (
      <Banner.Root tone="error">
        <Banner.Message>
          <ErrorText error={table.error} nextStep={LOAD_NEXT_STEP} />
        </Banner.Message>
        <Banner.Actions>
          <Button variant="secondary" size="sm" onClick={() => void loadTable(scope)}>
            Try Again
          </Button>
        </Banner.Actions>
      </Banner.Root>
    )
  }
  return <VariableTableEditor key={scope} scope={scope} label={label} table={table} />
}

interface ErrorTextProps {
  error: AppError
  nextStep: string
}

function ErrorText({ error, nextStep }: ErrorTextProps) {
  return <NoTranslateText text={withNextStep(error.description || error.code, nextStep)} terms={[error.code]} />
}

interface DraftState {
  base: VariableRow[]
  rows: DraftRow[]
  dirty: boolean
}

function draftOf(rows: VariableRow[]): DraftState {
  return { base: rows, rows: toDraftRows(rows), dirty: false }
}

interface VariableTableEditorProps {
  scope: string
  label: string
  table: TableState
}

function VariableTableEditor({ scope, label, table }: VariableTableEditorProps) {
  const saveTable = useEnvironmentsStore((state) => state.saveTable)
  const resetValues = useEnvironmentsStore((state) => state.resetValues)
  const persistValues = useEnvironmentsStore((state) => state.persistValues)
  const [draft, setDraft] = useState(() => draftOf(table.rows))
  const [confirming, setConfirming] = useState<TypeSwitches | null>(null)
  const [failure, setFailure] = useState<AppError | null>(null)
  const [dismissedWarnings, setDismissedWarnings] = useState<string[] | null>(null)
  const [busy, startAction] = useTransition()

  if (table.rows !== draft.base && !draft.dirty) {
    setDraft(draftOf(table.rows))
  }

  const stale = draft.dirty && table.rows !== draft.base
  const hasOverrides = table.rows.some((row) => !row.secret && row.hasCurrentOverride)
  const showGitHistory = table.warnings.includes(GIT_HISTORY_WARNING) && dismissedWarnings !== table.warnings

  function edit(change: (rows: DraftRow[]) => DraftRow[]) {
    setDraft((current) => ({ ...current, rows: change(current.rows), dirty: true }))
  }

  const actions: RowActions = {
    onChange: (id, patch) => edit((rows) => patchRow(rows, id, patch)),
    onRemove: (id) => edit((rows) => removeRow(rows, id)),
    onReset: (key) => run(() => resetValues(scope, [key])),
    onPersist: (key) => run(() => persistValues(scope, [key])),
  }

  function run(operation: () => Promise<AppError | null>) {
    startAction(async () => {
      setFailure(await operation())
    })
  }

  function discard() {
    setConfirming(null)
    setFailure(null)
    setDraft(draftOf(table.rows))
  }

  function requestSave() {
    const switches = typeSwitches(draft.rows)
    if (switches.toSecret.length > 0 || switches.toPlain.length > 0) {
      setConfirming(switches)
      return
    }
    commit(false)
  }

  function commit(confirmReveal: boolean) {
    const rows = draft.rows
    startAction(async () => {
      const error = await saveTable(scope, toVariableEdits(rows), confirmReveal)
      if (error?.code === CONFIRM_REVEAL_REQUIRED) {
        setConfirming({ toSecret: [], toPlain: typeSwitches(rows).toPlain, reveal: true })
        return
      }
      setFailure(error)
      if (error === null) {
        setConfirming(null)
        setDraft((current) => ({ ...current, dirty: false }))
      }
    })
  }

  return (
    <div className={styles.root}>
      <div className={styles.toolbar}>
        <Button variant="ghost" size="sm" onClick={() => edit((rows) => [...rows, newDraftRow()])}>
          Add Variable
        </Button>
        <span className={styles.spacer} />
        <Button variant="ghost" size="sm" disabled={busy || draft.dirty || !hasOverrides} onClick={() => run(() => resetValues(scope, []))}>
          Reset All
        </Button>
        <Button variant="ghost" size="sm" disabled={busy || draft.dirty || !hasOverrides} onClick={() => run(() => persistValues(scope, []))}>
          Persist All
        </Button>
      </div>
      {showGitHistory ? (
        <Banner.Root tone="warning">
          {GIT_HISTORY_MESSAGE}
          <Banner.Actions>
            <Button variant="ghost" size="sm" onClick={() => setDismissedWarnings(table.warnings)}>
              Dismiss
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      {stale ? <Banner.Root tone="info">{STALE_MESSAGE}</Banner.Root> : null}
      <div className={styles.scroller}>
        <table aria-label={label} className={styles.table}>
          {TABLE_HEAD}
          <tbody>
            {draft.rows.length === 0 ? (
              <tr>
                <td colSpan={7} className={styles.empty}>
                  No variables yet. Add one to use it as {'{{name}}'}.
                </td>
              </tr>
            ) : (
              draft.rows.map((row, index) => (
                <VariableTableRow
                  key={row.id}
                  scope={scope}
                  row={row}
                  index={index}
                  problems={keyProblems(draft.rows, row)}
                  savedActionsDisabled={busy || draft.dirty}
                  actions={actions}
                />
              ))
            )}
          </tbody>
        </table>
      </div>
      {confirming ? (
        <TypeSwitchConfirm
          switches={confirming}
          busy={busy}
          onCancel={() => setConfirming(null)}
          onConfirm={() => commit(confirming.reveal)}
        />
      ) : null}
      {failure ? (
        <Banner.Root tone="error">
          <Banner.Message>
            <ErrorText error={failure} nextStep={SAVE_NEXT_STEP} />
          </Banner.Message>
        </Banner.Root>
      ) : null}
      <div className={styles.footer}>
        <Button variant="secondary" size="sm" disabled={!draft.dirty || busy} onClick={discard}>
          Discard Changes
        </Button>
        <Button variant="primary" size="sm" disabled={!draft.dirty || busy || confirming !== null} onClick={requestSave}>
          {busy ? 'Saving…' : 'Save Variables'}
        </Button>
      </div>
    </div>
  )
}
