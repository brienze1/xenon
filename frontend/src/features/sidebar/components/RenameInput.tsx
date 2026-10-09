import { useId, useRef, useState, type KeyboardEvent } from 'react'
import type { AppError } from '@/platform/errors'
import { SidebarErrorText } from './SidebarErrorText'
import styles from './RenameInput.module.css'

const RENAME_NEXT_STEP = 'Choose another name, or press Escape to keep the current one.'

interface RenameInputProps {
  name: string
  onCommit: (name: string) => Promise<AppError | null>
  onCancel: () => void
}

export function RenameInput({ name, onCommit, onCancel }: RenameInputProps) {
  const [value, setValue] = useState(name)
  const [error, setError] = useState<AppError | null>(null)
  const committing = useRef(false)
  const errorId = useId()

  async function commit() {
    if (committing.current) {
      return
    }
    committing.current = true
    const failure = await onCommit(value)
    committing.current = false
    setError(failure)
  }

  function onKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    event.stopPropagation()
    if (event.key === 'Enter') {
      void commit()
    }
    if (event.key === 'Escape') {
      onCancel()
    }
  }

  return (
    <span className={styles.field}>
      <input
        name="item-name"
        autoComplete="off"
        aria-label="Name"
        aria-invalid={error !== null}
        aria-describedby={error ? errorId : undefined}
        className={styles.rename}
        value={value}
        autoFocus
        spellCheck={false}
        onFocus={(event) => event.currentTarget.select()}
        onChange={(event) => {
          setValue(event.target.value)
          setError(null)
        }}
        onKeyDown={onKeyDown}
        onClick={(event) => event.stopPropagation()}
        onBlur={() => void commit()}
      />
      {error ? (
        <span id={errorId} role="alert" className={styles.error}>
          <SidebarErrorText error={error} nextStep={RENAME_NEXT_STEP} />
        </span>
      ) : null}
    </span>
  )
}
