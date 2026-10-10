import { useState, type FormEvent } from 'react'
import { useEnvironmentsStore } from '@/features/environments/store'
import type { AppError } from '@/platform/errors'
import type { EditableScope, VariableUsage } from '@/platform/mappers/inspection'
import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button/Button'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { addableScopes, isEditableScope, SCOPE_NAMES, STATUS_LABELS } from '@/features/side-panel/lib/variable-items'
import styles from './VariableItem.module.css'

interface VariableItemProps {
  variable: VariableUsage
  scopes: Partial<Record<EditableScope, string>>
  scopeLabel: string
  onChanged: () => void
}

const EYE_GLYPH = <span aria-hidden="true">◉</span>
const COPY_GLYPH = <span aria-hidden="true">⧉</span>

/** One variable of the request: value and source, edited in place (current value) or added when unresolved. */
export function VariableItem({ variable, scopes, scopeLabel, onChanged }: VariableItemProps) {
  const setCurrentValue = useEnvironmentsStore((state) => state.setCurrentValue)
  const addVariable = useEnvironmentsStore((state) => state.addVariable)
  const revealSecret = useEnvironmentsStore((state) => state.revealSecret)
  const copySecret = useEnvironmentsStore((state) => state.copySecret)
  const openPassphrase = useEnvironmentsStore((state) => state.openPassphrase)
  const [draft, setDraft] = useState<string | null>(null)
  const [revealed, setRevealed] = useState<string | null>(null)
  const [target, setTarget] = useState<EditableScope | ''>('')
  const [failure, setFailure] = useState<AppError | null>(null)
  const scopeRef = isEditableScope(variable.scope) ? scopes[variable.scope] : undefined
  const choices = addableScopes({ scopes })
  const addTo = target === '' ? choices[0] : target

  async function finish(error: AppError | null) {
    setFailure(error)
    if (error === null) {
      setDraft(null)
      setRevealed(null)
      onChanged()
    }
  }

  async function save(event: FormEvent) {
    event.preventDefault()
    if (draft === null) {
      return
    }
    if (variable.status === 'unresolved' && addTo !== undefined) {
      await finish(await addVariable(scopes[addTo] as string, variable.name, draft, false))
    } else if (scopeRef !== undefined) {
      await finish(await setCurrentValue(scopeRef, variable.name, draft))
    }
  }

  async function toggleReveal() {
    if (revealed !== null || scopeRef === undefined) {
      setRevealed(null)
      return
    }
    const result = await revealSecret(scopeRef, variable.name)
    if (result.ok) {
      setRevealed(result.value)
      setFailure(null)
    } else {
      setFailure(result.error)
    }
  }

  async function copy() {
    if (scopeRef !== undefined) {
      setFailure(await copySecret(scopeRef, variable.name))
    }
  }

  const editable = scopeRef !== undefined && (variable.status === 'resolved' || variable.status === 'missing_secret')
  const addable = variable.status === 'unresolved' && variable.scope === 'none' && choices.length > 0
  const shown = revealed ?? variable.displayValue
  const valueLabel = `Value of ${variable.name}`

  return (
    <li className={styles.item}>
      <div className={styles.header}>
        <code className={styles.name} translate="no">
          {variable.name}
        </code>
        <span className={cn(styles.status, styles[variable.status])}>{STATUS_LABELS[variable.status]}</span>
      </div>
      <span className={styles.scope}>{scopeLabel}</span>
      {variable.status === 'dynamic' ? <span className={styles.muted}>Generated at send</span> : null}
      {variable.status === 'locked_secret' ? (
        <Button size="sm" variant="secondary" onClick={() => openPassphrase('unlock')}>
          Unlock Vault
        </Button>
      ) : null}
      {editable || addable ? (
        <form className={styles.form} onSubmit={(event) => void save(event)}>
          {addable ? (
            <select
              name="add-scope"
              aria-label={`Add ${variable.name} to`}
              className={styles.select}
              value={addTo}
              onChange={(event) => setTarget(event.target.value as EditableScope)}
            >
              {choices.map((scope) => (
                <option key={scope} value={scope}>
                  {`Add to ${SCOPE_NAMES[scope].toLowerCase()}`}
                </option>
              ))}
            </select>
          ) : null}
          <input
            name="variable-value"
            autoComplete="off"
            spellCheck={false}
            translate="no"
            aria-label={valueLabel}
            className={styles.value}
            type={variable.secret && revealed === null && draft === null ? 'password' : 'text'}
            placeholder={variable.status === 'missing_secret' ? 'No value on this machine' : 'Value…'}
            value={draft ?? (variable.status === 'resolved' ? shown : '')}
            onChange={(event) => setDraft(event.target.value)}
          />
          {variable.secret && variable.status === 'resolved' ? (
            <>
              <IconButton label={revealed === null ? `Reveal ${variable.name}` : `Hide ${variable.name}`} onClick={() => void toggleReveal()}>
                {EYE_GLYPH}
              </IconButton>
              <IconButton label={`Copy ${variable.name}`} onClick={() => void copy()}>
                {COPY_GLYPH}
              </IconButton>
            </>
          ) : null}
          {draft !== null ? (
            <Button type="submit" size="sm" variant="primary">
              {addable ? 'Add' : 'Save'}
            </Button>
          ) : null}
        </form>
      ) : null}
      {variable.usedIn.length > 0 ? (
        <span className={styles.usedIn} translate="no">
          {variable.usedIn.join(', ')}
        </span>
      ) : null}
      {failure ? (
        <span role="alert" className={styles.error}>
          {failure.description || failure.code}
        </span>
      ) : null}
    </li>
  )
}
