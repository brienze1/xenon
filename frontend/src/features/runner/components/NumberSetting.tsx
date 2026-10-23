import { useId, useState } from 'react'
import styles from './SettingFields.module.css'

interface NumberSettingProps {
  label: string
  name: string
  value: number
  min: number
  max: number
  unit?: string
  /** The inline error of a refused start (RUN-04400) for this setting. */
  error?: string
  onChange: (value: number) => void
}

/** A whole-number run setting; while typing, an empty or partial value is kept as text. */
export function NumberSetting({ label, name, value, min, max, unit, error, onChange }: NumberSettingProps) {
  const errorId = useId()
  const [draft, setDraft] = useState<string | null>(null)

  function edit(text: string) {
    setDraft(text)
    const parsed = Number(text)
    if (text.trim() !== '' && Number.isInteger(parsed)) {
      onChange(parsed)
    }
  }

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
            className={styles.number}
            aria-invalid={error ? true : undefined}
            aria-describedby={error ? errorId : undefined}
            value={draft ?? String(value)}
            onChange={(event) => edit(event.target.value)}
            onBlur={() => setDraft(null)}
          />
          {unit ? <span className={styles.unit}>{unit}</span> : null}
        </span>
      </label>
      {error ? (
        <p id={errorId} role="alert" className={styles.error}>
          {error}
        </p>
      ) : null}
    </div>
  )
}
