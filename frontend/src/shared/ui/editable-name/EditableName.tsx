import { useEffect, useId, useRef, useState, type KeyboardEvent } from 'react'
import styles from './EditableName.module.css'

/** The longest name any item takes (`maxItemNameLength` in Go). */
const NAME_MAX_LENGTH = 200

const CLICK_TO_RENAME = 'Click to rename'

interface EditableNameProps {
  /** The name shown: the tab's draft name. */
  value: string
  /** What is renamed, e.g. "Rename request"; the input's name. The text's name adds the name itself. */
  label: string
  onChange: (value: string) => void
  /** Escape: the owner drops the draft, so `value` turns back into the saved name. */
  onCancel: () => void
  error?: string | null
  /** Shown instead of an empty name (e.g. "Untitled Request"); never part of the value. */
  placeholder?: string
  maxLength?: number
}

/** A name shown as text that turns into an input in place; it never saves, the tab does (Mod+S). */
export function EditableName({ value, label, onChange, onCancel, error = null, placeholder = '', maxLength = NAME_MAX_LENGTH }: EditableNameProps) {
  const [editing, setEditing] = useState(false)
  const textRef = useRef<HTMLButtonElement>(null)
  const refocus = useRef(false)
  const errorId = useId()
  const invalid = error != null
  const described = invalid ? { 'aria-invalid': true, 'aria-describedby': errorId } : {}
  const shown = value === '' ? placeholder : value

  useEffect(() => {
    if (!editing && refocus.current) {
      refocus.current = false
      textRef.current?.focus()
    }
  }, [editing])

  /** Enter and Escape give the focus back to the name; a blur leaves it where the user moved it. */
  function close(focusText: boolean) {
    refocus.current = focusText
    setEditing(false)
  }

  function onKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'Enter') {
      event.preventDefault()
      event.stopPropagation()
      close(true)
      return
    }
    if (event.key === 'Escape') {
      event.preventDefault()
      event.stopPropagation()
      onCancel()
      close(true)
    }
  }

  return (
    <span className={styles.root}>
      {editing ? (
        <input
          name="item-name"
          autoComplete="off"
          aria-label={label}
          {...described}
          className={styles.input}
          translate="no"
          value={value}
          placeholder={placeholder || undefined}
          maxLength={maxLength}
          spellCheck={false}
          autoFocus
          onFocus={(event) => event.currentTarget.select()}
          onChange={(event) => onChange(event.target.value)}
          onKeyDown={onKeyDown}
          onBlur={() => close(false)}
        />
      ) : (
        <button
          ref={textRef}
          type="button"
          aria-label={`${label}: ${shown}`}
          {...described}
          title={shown === '' ? CLICK_TO_RENAME : `${shown}\n${CLICK_TO_RENAME}`}
          className={styles.text}
          translate="no"
          onClick={() => setEditing(true)}
        >
          {shown}
        </button>
      )}
      {invalid ? (
        <span id={errorId} role="alert" className={styles.error}>
          {error}
        </span>
      ) : null}
    </span>
  )
}
