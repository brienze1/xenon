import { useId, useState, useTransition } from 'react'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import type { AppError } from '@/platform/errors'
import type { EditableScope } from '@/platform/mappers/inspection'
import { withNextStep } from '@/shared/lib/next-step'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import styles from './AddVariableCard.module.css'

interface AddVariableCardProps {
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

const ADD_NEXT_STEP = 'Check the name and value, then try again.'

/** The hover card of an undefined `{{variable}}`: a value and the scope to add it to, like Postman's. */
export function AddVariableCard({ tabId, name, close }: AddVariableCardProps) {
  const scopes = useRequestBuilderStore((state) => state.tabs[tabId]?.inspection?.scopes)
  const inspect = useRequestBuilderStore((state) => state.inspect)
  const addVariable = useEnvironmentsStore((state) => state.addVariable)
  const environmentName = useEnvironmentsStore((state) => state.environments.find((environment) => environment.id === state.activeId)?.name ?? null)
  const [value, setValue] = useState('')
  const [failure, setFailure] = useState<AppError | null>(null)
  const [adding, startAdd] = useTransition()
  const headingId = useId()

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
        close()
      }
    })
  }

  return (
    <section aria-labelledby={headingId} className={styles.card}>
      <p id={headingId} className={styles.heading}>
        <code translate="no">{`{{${name}}}`}</code> <span className={styles.muted}>is not defined</span>
      </p>
      <input
        name="new-variable-value"
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
          const label = target.scope === 'environment' && environmentName ? `${target.label}: ${environmentName}` : target.label
          return (
            <li key={target.scope}>
              <button type="button" className={styles.target} disabled={!available || adding} onClick={() => add(target.scope)}>
                <span aria-hidden="true" className={styles[target.scope]}>
                  {target.badge}
                </span>
                <span translate="no">{available ? label : target.unavailable}</span>
              </button>
            </li>
          )
        })}
      </ul>
      {failure ? (
        <p role="alert" className={styles.error}>
          <NoTranslateText text={withNextStep(failure.description || failure.code, ADD_NEXT_STEP)} terms={[failure.code]} />
        </p>
      ) : null}
    </section>
  )
}
