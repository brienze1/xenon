import { useState } from 'react'
import type { AppError } from '@/platform/errors'
import type { SecretState } from '@/platform/mappers/variable'
import { withNextStep } from '@/shared/lib/next-step'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useEnvironmentsStore } from '@/features/environments/store'
import type { DraftRow } from '@/features/environments/lib/variable-draft'
import styles from './SecretValueCell.module.css'

export const MASK = '•••••'

const PLACEHOLDERS: Record<SecretState | 'new', string> = {
  set: MASK,
  missing: 'No value on this machine',
  locked: 'Vault locked',
  new: 'Secret value…',
}

const BADGE_TEXT: Partial<Record<SecretState, string>> = {
  missing: 'Missing',
  locked: 'Locked',
}

const BADGE_HINT: Partial<Record<SecretState, string>> = {
  missing: 'This secret has no value on this machine. Type one to store it in the vault.',
  locked: 'The vault is locked. Unlock it to use or reveal this secret.',
}

const SECRET_NEXT_STEP = 'Unlock the vault if it is locked, then try again.'

const EYE_GLYPH = <span aria-hidden="true">◉</span>
const COPY_GLYPH = <span aria-hidden="true">⧉</span>

interface SecretValueCellProps {
  scope: string
  row: DraftRow
  name: string
  onChange: (secretValue: string) => void
}

export function SecretValueCell({ scope, row, name, onChange }: SecretValueCellProps) {
  const revealSecret = useEnvironmentsStore((state) => state.revealSecret)
  const copySecret = useEnvironmentsStore((state) => state.copySecret)
  const [revealed, setRevealed] = useState<string | null>(null)
  const [visible, setVisible] = useState(false)
  const [failure, setFailure] = useState<AppError | null>(null)
  const storedKey = row.saved?.secret ? row.saved.key : null
  const state = row.saved?.secret ? row.saved.secretState : null
  const stored = storedKey !== null && state === 'set'
  const typed = row.secretValue !== null
  const value = row.secretValue ?? (visible ? (revealed ?? '') : '')
  const badge = state ? BADGE_TEXT[state] : undefined

  async function toggleReveal() {
    setFailure(null)
    if (visible) {
      setVisible(false)
      setRevealed(null)
      return
    }
    if (!typed && storedKey !== null) {
      const result = await revealSecret(scope, storedKey)
      if (!result.ok) {
        setFailure(result.error)
        return
      }
      setRevealed(result.value)
    }
    setVisible(true)
  }

  async function copy() {
    if (storedKey === null) {
      return
    }
    setFailure(await copySecret(scope, storedKey))
  }

  return (
    <div className={styles.cell}>
      <input
        name="secret-value"
        type={visible ? 'text' : 'password'}
        autoComplete="off"
        aria-label={`Value of ${name}`}
        aria-invalid={failure ? true : undefined}
        className={styles.input}
        spellCheck={false}
        placeholder={PLACEHOLDERS[state ?? 'new']}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
      {badge && state ? (
        <span className={styles.badge} data-tone={state} title={BADGE_HINT[state]}>
          {badge}
        </span>
      ) : null}
      <IconButton label={visible ? `Hide ${name}` : `Reveal ${name}`} aria-pressed={visible} disabled={!stored && !typed} onClick={() => void toggleReveal()}>
        {EYE_GLYPH}
      </IconButton>
      <IconButton label={`Copy ${name}`} disabled={!stored || typed} onClick={() => void copy()}>
        {COPY_GLYPH}
      </IconButton>
      {failure ? (
        <span role="alert" className={styles.error}>
          <NoTranslateText text={withNextStep(failure.description || failure.code, SECRET_NEXT_STEP)} terms={[failure.code]} />
        </span>
      ) : null}
    </div>
  )
}
