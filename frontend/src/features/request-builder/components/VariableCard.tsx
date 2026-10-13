import { useEffect, useId, useRef, useState, useTransition, type KeyboardEvent } from 'react'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import type { AppError } from '@/platform/errors'
import type { EditableScope, VariableUsage } from '@/platform/mappers/inspection'
import { withNextStep } from '@/shared/lib/next-step'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import styles from './VariableCard.module.css'

interface VariableCardProps {
  tabId: string
  name: string
  close: () => void
}

interface Target {
  scope: EditableScope
  badge: string
  label: string
  /** Shown instead when the request has no such scope. */
  unavailable: string
}

const TARGETS: readonly Target[] = [
  { scope: 'environment', badge: 'E', label: 'Environment', unavailable: 'No active environment' },
  { scope: 'collection', badge: 'C', label: 'Collection', unavailable: 'No parent collection' },
  { scope: 'globals', badge: 'G', label: 'Globals', unavailable: 'No globals' },
]

const NEXT_STEP = 'Check the name and value, then try again.'

const CHARS_PER_ROW = 36
const MAX_ROWS = 6

/** Grows the value box with its text, like Postman's, up to a few lines. */
function rowsFor(text: string): number {
  const rows = text.split('\n').reduce((total, line) => total + Math.max(1, Math.ceil(line.length / CHARS_PER_ROW)), 0)
  return Math.min(MAX_ROWS, rows)
}

function isEditableScope(scope: string): scope is EditableScope {
  return scope === 'environment' || scope === 'collection' || scope === 'globals'
}

function targetOf(scope: EditableScope): Target {
  return TARGETS.find((target) => target.scope === scope) as Target
}

function useEnvironmentName(): string | null {
  return useEnvironmentsStore((state) => state.environments.find((environment) => environment.id === state.activeId)?.name ?? null)
}

function scopeName(target: Target, environmentName: string | null): string {
  return target.scope === 'environment' && environmentName ? `${target.label}: ${environmentName}` : target.label
}

function showVariablesInRequest(close: () => void) {
  useWorkspacesStore.getState().setLayout({ rightPanelVisible: true, rightPanelView: 'variables' })
  close()
}

function Failure({ error }: { error: AppError }) {
  return (
    <p role="alert" className={styles.error}>
      <NoTranslateText text={withNextStep(error.description || error.code, NEXT_STEP)} terms={[error.code]} />
    </p>
  )
}

/**
 * The hover card of a `{{variable}}`, like Postman's: a defined one shows its current value to edit in place and where
 * it comes from; an undefined one takes a value and the scope to add it to.
 */
export function VariableCard({ tabId, name, close }: VariableCardProps) {
  const usage = useRequestBuilderStore((state) => state.tabs[tabId]?.inspection?.variables.find((variable) => variable.name === name))
  const headingId = useId()
  const defined = usage !== undefined && usage.status !== 'unresolved' && isEditableScope(usage.scope)
  return (
    <section aria-labelledby={headingId} className={styles.card}>
      <p id={headingId} className={styles.heading}>
        <code translate="no">{`{{${name}}}`}</code>
        {defined ? null : (
          <>
            {' '}
            <span className={styles.muted}>is not defined</span>
          </>
        )}
      </p>
      {defined ? <DefinedVariable key={`${usage.scope}:${usage.displayValue}`} tabId={tabId} usage={usage} close={close} /> : <NewVariable tabId={tabId} name={name} />}
    </section>
  )
}

interface DefinedVariableProps {
  tabId: string
  usage: VariableUsage
  close: () => void
}

/** The current value (this machine's, never written to the workspace files), saved on Enter or when focus leaves. */
function DefinedVariable({ tabId, usage, close }: DefinedVariableProps) {
  const table = useRequestBuilderStore((state) => (isEditableScope(usage.scope) ? state.tabs[tabId]?.inspection?.scopes[usage.scope] : undefined))
  const inspect = useRequestBuilderStore((state) => state.inspect)
  const setCurrentValue = useEnvironmentsStore((state) => state.setCurrentValue)
  const openPassphrase = useEnvironmentsStore((state) => state.openPassphrase)
  const environmentName = useEnvironmentName()
  const shown = usage.secret ? '' : usage.displayValue
  const [draft, setDraft] = useState(shown)
  const [failure, setFailure] = useState<AppError | null>(null)
  const [saving, startSave] = useTransition()
  const target = targetOf(usage.scope as EditableScope)
  const locked = usage.status === 'locked_secret'

  // The value the variable has now: an edit equal to it needs no save. A secret's is never known here, so it starts empty.
  const committed = useRef(shown)
  const unsaved = useRef<{ table: string | undefined; draft: string }>({ table, draft })

  useEffect(() => {
    unsaved.current = { table, draft }
  }, [table, draft])

  // The card is removed when the pointer leaves or a press lands elsewhere, without a blur: keep what was typed.
  useEffect(
    () => () => {
      const pending = unsaved.current
      if (pending.table !== undefined && pending.draft !== committed.current && !(usage.secret && pending.draft === '')) {
        void setCurrentValue(pending.table, usage.name, pending.draft).then((error) => (error === null ? inspect(tabId) : undefined))
      }
    },
    [inspect, setCurrentValue, tabId, usage.name, usage.secret],
  )

  function save() {
    if (table === undefined || draft === committed.current || (usage.secret && draft === '')) {
      return
    }
    const value = draft
    committed.current = value
    startSave(async () => {
      const error = await setCurrentValue(table, usage.name, value)
      setFailure(error)
      if (error === null) {
        await inspect(tabId)
      } else {
        committed.current = shown
      }
    })
  }

  function onKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      save()
    }
  }

  return (
    <>
      {locked ? (
        <button type="button" className={styles.link} onClick={() => openPassphrase('unlock')}>
          Unlock the vault to use this secret
        </button>
      ) : (
        <textarea
          name="variable-value"
          rows={rowsFor(draft)}
          autoComplete="off"
          spellCheck={false}
          translate="no"
          aria-label={`Value of ${usage.name}`}
          className={styles.value}
          placeholder={usage.secret ? (usage.status === 'missing_secret' ? 'No value on this machine — enter one…' : 'Secret — type to replace…') : 'Enter value…'}
          value={draft}
          disabled={saving}
          onChange={(event) => setDraft(event.target.value)}
          onKeyDown={onKeyDown}
          onBlur={save}
        />
      )}
      {failure ? <Failure error={failure} /> : null}
      <div className={styles.footer}>
        <span className={styles.source}>
          <span aria-hidden="true" className={styles[target.scope]}>
            {target.badge}
          </span>
          <span translate="no">{scopeName(target, environmentName)}</span>
        </span>
        <button type="button" className={styles.link} onClick={() => showVariablesInRequest(close)}>
          Variables in request →
        </button>
      </div>
    </>
  )
}

interface NewVariableProps {
  tabId: string
  name: string
}

/** A value and the scopes it can be added to; once added the card turns into the defined variable's. */
function NewVariable({ tabId, name }: NewVariableProps) {
  const scopes = useRequestBuilderStore((state) => state.tabs[tabId]?.inspection?.scopes)
  const inspect = useRequestBuilderStore((state) => state.inspect)
  const addVariable = useEnvironmentsStore((state) => state.addVariable)
  const environmentName = useEnvironmentName()
  const [value, setValue] = useState('')
  const [failure, setFailure] = useState<AppError | null>(null)
  const [adding, startAdd] = useTransition()

  function add(scope: EditableScope) {
    const table = scopes?.[scope]
    if (table === undefined) {
      return
    }
    startAdd(async () => {
      const error = await addVariable(table, name, value, false)
      setFailure(error)
      if (error === null) {
        await inspect(tabId)
      }
    })
  }

  return (
    <>
      <textarea
        name="new-variable-value"
        rows={rowsFor(value)}
        autoComplete="off"
        spellCheck={false}
        translate="no"
        aria-label={`Value of ${name}`}
        className={styles.value}
        placeholder="Enter value…"
        value={value}
        onChange={(event) => setValue(event.target.value)}
      />
      <p className={styles.muted}>Add to</p>
      <ul className={styles.targets} aria-label={`Add ${name} to`}>
        {TARGETS.map((target) => {
          const available = scopes?.[target.scope] !== undefined
          return (
            <li key={target.scope}>
              <button type="button" className={styles.target} disabled={!available || adding} onClick={() => add(target.scope)}>
                <span aria-hidden="true" className={styles[target.scope]}>
                  {target.badge}
                </span>
                <span translate="no">{available ? scopeName(target, environmentName) : target.unavailable}</span>
              </button>
            </li>
          )
        })}
      </ul>
      {failure ? <Failure error={failure} /> : null}
    </>
  )
}
