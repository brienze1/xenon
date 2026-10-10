import { useId, useRef, useState, useTransition, type FormEvent } from 'react'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import type { AppError } from '@/platform/errors'
import type { EditableScope } from '@/platform/mappers/inspection'
import { Button } from '@/shared/ui/button/Button'
import { addableScopes, isValidVariableName, SCOPE_NAMES } from '@/features/side-panel/lib/variable-items'
import styles from './MoveToSecret.module.css'
import { VariableErrorText } from './VariableErrorText'

interface MoveToSecretProps {
  tabId: string
  scopes: Partial<Record<EditableScope, string>>
}

const PREVIEW_LENGTH = 24

const NAME_INVALID = 'Enter a name without {, } or line breaks.'

const MOVE_NEXT_STEP = 'Check the name and scope, then try again.'

/** "Move to secret variable": the text selected in a request field becomes a secret and is replaced by `{{name}}`. */
export function MoveToSecret({ tabId, scopes }: MoveToSecretProps) {
  const selection = useRequestBuilderStore((state) => state.tabs[tabId]?.selection ?? null)
  const replaceSelection = useRequestBuilderStore((state) => state.replaceSelection)
  const addVariable = useEnvironmentsStore((state) => state.addVariable)
  const [name, setName] = useState('')
  const [target, setTarget] = useState<EditableScope | ''>('')
  const [nameError, setNameError] = useState<string | null>(null)
  const [failure, setFailure] = useState<AppError | null>(null)
  const [moving, startMove] = useTransition()
  const nameRef = useRef<HTMLInputElement>(null)
  const nameErrorId = useId()
  const choices = addableScopes({ scopes })
  const scope = target === '' ? choices[0] : target
  if (selection === null || scope === undefined) {
    return null
  }
  const preview = selection.text.length > PREVIEW_LENGTH ? `${selection.text.slice(0, PREVIEW_LENGTH)}…` : selection.text

  function move(event: FormEvent) {
    event.preventDefault()
    if (selection === null || scope === undefined) {
      return
    }
    if (!isValidVariableName(name)) {
      setNameError(NAME_INVALID)
      nameRef.current?.focus()
      return
    }
    const text = selection.text
    startMove(async () => {
      const error = await addVariable(scopes[scope] as string, name, text, true)
      setFailure(error)
      if (error === null) {
        replaceSelection(tabId, `{{${name}}}`)
        setName('')
      }
    })
  }

  return (
    <form className={styles.root} aria-label="Move to secret variable" noValidate onSubmit={move}>
      <p className={styles.selected}>
        Selected in <code translate="no">{selection.field}</code>: <code translate="no">{preview}</code>
      </p>
      <div className={styles.fields}>
        <input
          ref={nameRef}
          name="secret-name"
          autoComplete="off"
          spellCheck={false}
          translate="no"
          aria-label="Secret variable name"
          placeholder="api_token…"
          className={styles.input}
          aria-invalid={nameError !== null}
          aria-describedby={nameError ? nameErrorId : undefined}
          value={name}
          onChange={(event) => {
            setName(event.target.value)
            setNameError(null)
          }}
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
      {nameError ? (
        <p id={nameErrorId} role="alert" className={styles.error}>
          {nameError}
        </p>
      ) : null}
      <Button type="submit" size="sm" variant="primary" disabled={moving}>
        {moving ? 'Moving…' : 'Move to Secret Variable'}
      </Button>
      {failure ? (
        <span role="alert" className={styles.error}>
          <VariableErrorText error={failure} nextStep={MOVE_NEXT_STEP} />
        </span>
      ) : null}
    </form>
  )
}
