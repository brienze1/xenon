import { useId } from 'react'
import styles from './SettingsSection.module.css'

interface NumberFieldProps {
  label: string
  /** The form control name; Save focuses the first invalid field by it. */
  name: string
  value: string
  min: number
  max: number
  unit?: string
  hint?: string
  error?: string
  onChange: (value: string) => void
}

export function NumberField({ label, name, value, min, max, unit, hint, error, onChange }: NumberFieldProps) {
  const hintId = useId()
  const errorId = useId()
  const describedBy = [error ? errorId : null, hint ? hintId : null].filter(Boolean).join(' ') || undefined
  return (
    <div className={styles.field}>
      <label className={styles.label}>
        {label}
        <span className={styles.control}>
          <input
            name={name}
            type="number"
            inputMode="numeric"
            min={min}
            max={max}
            step={1}
            autoComplete="off"
            className={styles.input}
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy}
            value={value}
            onChange={(event) => onChange(event.target.value)}
          />
          {unit ? <span className={styles.unit}>{unit}</span> : null}
        </span>
      </label>
      {error ? (
        <p id={errorId} role="alert" className={styles.fieldError}>
          {error}
        </p>
      ) : null}
      {hint ? (
        <p id={hintId} className={styles.hint}>
          {hint}
        </p>
      ) : null}
    </div>
  )
}
