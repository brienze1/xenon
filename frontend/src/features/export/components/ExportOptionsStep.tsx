import { useId } from 'react'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import type { ExportFormatOption, ExportOptions } from '@/platform/mappers/export'
import { canExport, formatOption, passphraseProblem } from '@/features/export/lib/choices'
import { formatLabel, passphraseProblemLabel, shapeLabel } from '@/features/export/lib/labels'
import type { ExportChoices } from '@/features/export/model'
import { useExportStore } from '@/features/export/store'
import { ExportErrorBanner } from './ExportErrorBanner'
import styles from './ExportDialog.module.css'

const POSTMAN_SECRETS =
  'Postman exports never carry secret values: secret variables keep their names with empty values. Export as Xenon YAML to include them, encrypted with a passphrase.'

const CURRENT_VALUES_HINT = 'Plain variables’ current values replace their initial values in the exported files only; the workspace is not changed.'

const SECRETS_HINT = 'Anyone with the file and the passphrase can read the secret values. Xenon cannot recover a forgotten passphrase.'

interface FormatChoiceProps {
  options: ExportOptions
  choices: ExportChoices
}

function FormatChoice({ options, choices }: FormatChoiceProps) {
  const setFormat = useExportStore((state) => state.setFormat)
  return (
    <fieldset className={styles.group}>
      <legend className={styles.legend}>Format</legend>
      <div className={styles.options}>
        {options.formats.map((option) => (
          <label key={option.format} className={styles.option}>
            <input type="radio" name="export-format" value={option.format} checked={choices.format === option.format} onChange={() => setFormat(option.format)} />
            {formatLabel(option.format)}
          </label>
        ))}
      </div>
    </fieldset>
  )
}

interface ShapeChoiceProps {
  option: ExportFormatOption
  choices: ExportChoices
}

function ShapeChoice({ option, choices }: ShapeChoiceProps) {
  const setShape = useExportStore((state) => state.setShape)
  return (
    <fieldset className={styles.group}>
      <legend className={styles.legend}>Shape</legend>
      <div className={styles.options}>
        {option.shapes.map((shape) => (
          <label key={shape} className={styles.option}>
            <input type="radio" name="export-shape" value={shape} checked={choices.shape === shape} onChange={() => setShape(shape)} />
            {shapeLabel(shape)}
          </label>
        ))}
      </div>
    </fieldset>
  )
}

interface PassphraseFieldsProps {
  choices: ExportChoices
}

function PassphraseFields({ choices }: PassphraseFieldsProps) {
  const setChoice = useExportStore((state) => state.setChoice)
  const passphraseId = useId()
  const confirmationId = useId()
  const problemId = useId()
  const hintId = useId()
  const problem = passphraseProblem(choices)
  const shortShown = problem === 'too_short' && choices.passphrase !== ''
  const mismatchShown = problem === 'mismatch' && choices.passphraseConfirmation !== ''
  return (
    <div className={styles.passphrase}>
      <p id={hintId} className={styles.hint}>
        {SECRETS_HINT}
      </p>
      <div className={styles.field}>
        <label htmlFor={passphraseId} className={styles.label}>
          Passphrase
        </label>
        <input
          id={passphraseId}
          name="export-passphrase"
          type="password"
          autoComplete="new-password"
          spellCheck={false}
          className={styles.input}
          aria-describedby={shortShown ? `${hintId} ${problemId}` : hintId}
          aria-invalid={shortShown ? true : undefined}
          value={choices.passphrase}
          onChange={(event) => setChoice({ passphrase: event.target.value })}
        />
      </div>
      <div className={styles.field}>
        <label htmlFor={confirmationId} className={styles.label}>
          Confirm Passphrase
        </label>
        <input
          id={confirmationId}
          name="export-passphrase-confirmation"
          type="password"
          autoComplete="new-password"
          spellCheck={false}
          className={styles.input}
          aria-describedby={mismatchShown ? problemId : undefined}
          aria-invalid={mismatchShown ? true : undefined}
          value={choices.passphraseConfirmation}
          onChange={(event) => setChoice({ passphraseConfirmation: event.target.value })}
        />
      </div>
      {problem !== null && (shortShown || mismatchShown) ? (
        <p id={problemId} className={styles.fieldError} role="alert">
          {passphraseProblemLabel(problem)}
        </p>
      ) : null}
    </div>
  )
}

interface ValueChoicesProps {
  options: ExportOptions
  option: ExportFormatOption
  choices: ExportChoices
}

function ValueChoices({ options, option, choices }: ValueChoicesProps) {
  const setChoice = useExportStore((state) => state.setChoice)
  const currentValuesHintId = useId()
  return (
    <div className={styles.group}>
      {option.includeSecretsAllowed && options.hasSecrets ? (
        <label className={styles.option}>
          <input
            type="checkbox"
            name="export-include-secrets"
            checked={choices.includeSecrets}
            onChange={(event) => setChoice({ includeSecrets: event.target.checked })}
          />
          Include secrets (encrypted with a passphrase)
        </label>
      ) : null}
      {option.includeSecretsAllowed ? null : (
        <p className={styles.hint}>
          {POSTMAN_SECRETS}
        </p>
      )}
      {option.includeSecretsAllowed && choices.includeSecrets ? <PassphraseFields choices={choices} /> : null}
      <label className={styles.option}>
        <input
          type="checkbox"
          name="export-current-values"
          aria-describedby={currentValuesHintId}
          checked={choices.currentValuesAsInitial}
          onChange={(event) => setChoice({ currentValuesAsInitial: event.target.checked })}
        />
        Export current values as initial values
      </label>
      <p id={currentValuesHintId} className={styles.hint}>
        {CURRENT_VALUES_HINT}
      </p>
    </div>
  )
}

export function ExportOptionsStep() {
  const options = useExportStore((state) => state.options)
  const choices = useExportStore((state) => state.choices)
  const close = useExportStore((state) => state.close)
  const start = useExportStore((state) => state.start)
  const option = options && choices ? formatOption(options, choices.format) : undefined

  return (
    <>
      <Dialog.Body>
        {options && choices && option ? (
          <form
            className={styles.form}
            noValidate
            onSubmit={(event) => {
              event.preventDefault()
              void start()
            }}
          >
            <FormatChoice options={options} choices={choices} />
            {option.shapes.length > 1 ? <ShapeChoice option={option} choices={choices} /> : null}
            <ValueChoices options={options} option={option} choices={choices} />
          </form>
        ) : null}
        <ExportErrorBanner />
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={close}>
          Cancel
        </Button>
        <Button variant="primary" disabled={!options || !choices || !canExport(options, choices)} onClick={() => void start()}>
          Export…
        </Button>
      </Dialog.Actions>
    </>
  )
}
