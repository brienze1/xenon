import { useState, type KeyboardEvent } from 'react'
import styles from './RenameInput.module.css'

interface RenameInputProps {
  name: string
  onCommit: (name: string) => void
  onCancel: () => void
}

export function RenameInput({ name, onCommit, onCancel }: RenameInputProps) {
  const [value, setValue] = useState(name)

  function onKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    event.stopPropagation()
    if (event.key === 'Enter') {
      onCommit(value)
    }
    if (event.key === 'Escape') {
      onCancel()
    }
  }

  return (
    <input
      name="item-name"
      autoComplete="off"
      aria-label="Name"
      className={styles.rename}
      value={value}
      autoFocus
      spellCheck={false}
      onFocus={(event) => event.currentTarget.select()}
      onChange={(event) => setValue(event.target.value)}
      onKeyDown={onKeyDown}
      onClick={(event) => event.stopPropagation()}
      onBlur={() => onCommit(value)}
    />
  )
}
