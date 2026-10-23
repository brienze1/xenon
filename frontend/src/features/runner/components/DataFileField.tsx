import { useId } from 'react'
import { Button } from '@/shared/ui/button/Button'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useRunnerStore } from '@/features/runner/store'
import styles from './SettingFields.module.css'

const CLEAR_GLYPH = <span aria-hidden="true">×</span>

function fileName(path: string): string {
  return path.split(/[\\/]/).pop() ?? path
}

interface DataFileFieldProps {
  itemId: string
  /** The data file of the shown run type. */
  path: string
}

/** Iteration data: Select file… (CSV or JSON), the chosen file's name, and × to clear it. */
export function DataFileField({ itemId, path }: DataFileFieldProps) {
  const labelId = useId()
  const errorId = useId()
  const error = useRunnerStore((state) => state.targets[itemId]?.fieldErrors.data_file)
  const pickDataFile = useRunnerStore((state) => state.pickDataFile)
  const clearDataFile = useRunnerStore((state) => state.clearDataFile)

  return (
    <div className={styles.field} role="group" aria-labelledby={labelId} aria-describedby={error ? errorId : undefined}>
      <span id={labelId} className={styles.label}>
        Iteration data
      </span>
      <div className={styles.file}>
        <Button variant="secondary" size="sm" onClick={() => void pickDataFile(itemId)}>
          Select file…
        </Button>
        {path ? (
          <>
            <span className={styles.fileName} title={path} translate="no">
              {fileName(path)}
            </span>
            <IconButton label="Clear iteration data" onClick={() => clearDataFile(itemId)}>
              {CLEAR_GLYPH}
            </IconButton>
          </>
        ) : null}
      </div>
      {error ? (
        <p id={errorId} role="alert" className={styles.error}>
          {error}
        </p>
      ) : null}
    </div>
  )
}
