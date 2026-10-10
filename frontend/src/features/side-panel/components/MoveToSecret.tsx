import { useState, type FormEvent } from 'react'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import type { AppError } from '@/platform/errors'
import type { EditableScope } from '@/platform/mappers/inspection'
import { Button } from '@/shared/ui/button/Button'
import { addableScopes, isValidVariableName, SCOPE_NAMES } from '@/features/side-panel/lib/variable-items'
import styles from './MoveToSecret.module.css'

interface MoveToSecretProps {
  tabId: string
  scopes: Partial<Record<EditableScope, string>>
}

const PREVIEW_LENGTH = 24

/** "Move to secret variable": the text selected in a request field becomes a secret and is replaced by `{{name}}`. */
export function MoveToSecret({ tabId, scopes }: MoveToSecretProps) {
  const selection = useRequestBuilderStore((state) => state.tabs[tabId]?.selection ?? null)
  const replaceSelection = useRequestBuilderStore((state) => state.replaceSelection)
  const addVariable = useEnvironmentsStore((state) => state.addVariable)
  const [name, setName] = useState('')
  const [target, setTarget] = useState<EditableScope | ''>('')
  const [failure, setFailure] = useState<AppError | null>(null)
  const choices = addableScopes({ scopes })
  const scope = target === '' ? choices[0] : target
  if (selection === null || scope === undefined) {
    return null
  }
  const preview = selection.text.length > PREVIEW_LENGTH ? `${selection.text.slice(0, PREVIEW_LENGTH)}…` : selection.text

  async function move(event: FormEvent) {
    event.preventDefault()
    if (selection === null || scope === undefined || !isValidVariableName(name)) {
      return
    }
    const error = await addVariable(scopes[scope] as string, name, selection.text, true)
    setFailure(error)
    if (error === null) {
      replaceSelection(tabId, `{{${name}}}`)
      setName('')
    }
  }

  return (
    <form className={styles.root} aria-label="Move to secret variable" onSubmit={(event) => void move(event)}>
      <p className={styles.selected}>
        Selected in <code translate="no">{selection.field}</code>: <code translate="no">{preview}</code>
      </p>
      <div className={styles.fields}>
        <input
          name="secret-name"
          autoComplete="off"
          spellCheck={false}
          translate="no"
          aria-label="Secret variable name"
          placeholder="api_token…"
          className={styles.input}
          value={name}
          onChange={(event) => setName(event.target.value)}
        />
        <select
          name="secret-scope"
          aria-label="Secret variable scope"
          className={styles.input}
          value={scope}
          onChange={(event) => setTarget(event.target.value as EditableScope)}
        >
          {choices.map((choice) => (
            <option key={choice} value={choice}>
              {SCOPE_NAMES[choice]}
            </option>
          ))}
        </select>
      </div>
      <Button type="submit" size="sm" variant="primary" disabled={!isValidVariableName(name)}>
        Move to Secret Variable
      </Button>
      {failure ? (
        <span role="alert" className={styles.error}>
          {failure.description || failure.code}
        </span>
      ) : null}
    </form>
  )
}
