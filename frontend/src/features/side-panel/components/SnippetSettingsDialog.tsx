import { useId } from 'react'
import type { SnippetLanguage, SnippetOption } from '@/platform/mappers/snippet'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { optionValue } from '@/features/side-panel/lib/snippet-options'
import styles from './SnippetSettingsDialog.module.css'

interface SnippetSettingsDialogProps {
  language: SnippetLanguage
  options: Readonly<Record<string, string>>
  onChange: (option: SnippetOption, value: string) => void
  onClose: () => void
}

/** Postman's "Settings for <language>": every generator option with its explanation, applied as it changes. */
export function SnippetSettingsDialog({ language, options, onChange, onClose }: SnippetSettingsDialogProps) {
  return (
    <Dialog.Root open onOpenChange={(open) => (open ? undefined : onClose())} className={styles.dialog}>
      <Dialog.Title>Settings for {language.label}</Dialog.Title>
      <Dialog.Body>
        <ul className={styles.list}>
          {language.options.map((option) => (
            <SettingRow key={option.id} option={option} value={optionValue(options, option.id, option.defaultValue)} onChange={(value) => onChange(option, value)} />
          ))}
        </ul>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onClose}>
          Done
        </Button>
      </Dialog.Actions>
    </Dialog.Root>
  )
}

interface SettingRowProps {
  option: SnippetOption
  value: string
  onChange: (value: string) => void
}

function SettingRow({ option, value, onChange }: SettingRowProps) {
  const labelId = useId()
  const descriptionId = useId()
  return (
    <li className={styles.row}>
      <div className={styles.text}>
        <span id={labelId} className={styles.label}>
          {option.label}
        </span>
        {option.description ? (
          <span id={descriptionId} className={styles.description}>
            {option.description}
          </span>
        ) : null}
      </div>
      <SettingControl option={option} value={value} labelId={labelId} descriptionId={option.description ? descriptionId : undefined} onChange={onChange} />
    </li>
  )
}

interface SettingControlProps extends SettingRowProps {
  labelId: string
  descriptionId: string | undefined
}

function SettingControl({ option, value, labelId, descriptionId, onChange }: SettingControlProps) {
  if (option.type === 'boolean') {
    return (
      <input
        type="checkbox"
        role="switch"
        name={option.id}
        aria-labelledby={labelId}
        aria-describedby={descriptionId}
        className={styles.switch}
        checked={value === 'true'}
        onChange={(event) => onChange(String(event.target.checked))}
      />
    )
  }
  if (option.type === 'enum') {
    return (
      <select
        name={option.id}
        aria-labelledby={labelId}
        aria-describedby={descriptionId}
        className={styles.control}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {option.values.map((choice) => (
          <option key={choice} value={choice}>
            {choice}
          </option>
        ))}
      </select>
    )
  }
  return (
    <input
      type="number"
      name={option.id}
      min={0}
      step={1}
      inputMode="numeric"
      autoComplete="off"
      aria-labelledby={labelId}
      aria-describedby={descriptionId}
      className={styles.control}
      value={value}
      onChange={(event) => onChange(event.target.value === '' ? '0' : event.target.value)}
    />
  )
}
