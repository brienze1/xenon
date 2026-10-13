import { lazy, Suspense, useId, useState, type KeyboardEvent } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { selectLocation, useSidebarStore } from '@/features/sidebar/store'
import type { AppError } from '@/platform/errors'
import { withNextStep } from '@/shared/lib/next-step'
import { Button } from '@/shared/ui/button/Button'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import styles from './InfoPanel.module.css'

const MarkdownView = lazy(() => import('@/shared/ui/markdown/MarkdownView'))

const INFO_HEADING = <h2 className={styles.title}>Info</h2>

const NO_DESCRIPTION = <span className={styles.muted}>No description</span>

const RENAME_NEXT_STEP = 'Choose another name, then press Enter.'

const LOADING_DESCRIPTION = (
  <span role="status" className={styles.muted}>
    Loading…
  </span>
)

const LOCATION_LABEL = <span className={styles.label}>Location</span>

const DESCRIPTION_LABEL = <span className={styles.label}>Description</span>

interface FailedRename {
  tabId: string
  error: AppError
}

interface InfoPanelProps {
  tabId: string
}

export function InfoPanel({ tabId }: InfoPanelProps) {
  const requestId = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.id ?? '')
  const name = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.name ?? '')
  const description = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.description ?? '')
  const setName = useRequestBuilderStore((state) => state.setName)
  const setDescription = useRequestBuilderStore((state) => state.setDescription)
  const adoptRename = useRequestBuilderStore((state) => state.adoptRename)
  const location = useSidebarStore(useShallow((state) => (requestId ? selectLocation(state.tree, requestId).slice(0, -1) : [])))
  const rename = useSidebarStore((state) => state.rename)
  const [editing, setEditing] = useState(false)
  const [failedRename, setFailedRename] = useState<FailedRename | null>(null)
  const nameError = failedRename?.tabId === tabId ? failedRename.error : null
  const nameErrorId = useId()
  const descriptionId = useId()

  async function commitName(value: string) {
    const trimmed = value.trim()
    if (trimmed === '' || trimmed === name) {
      return
    }
    if (requestId === '') {
      setName(tabId, trimmed)
      return
    }
    const renamed = await rename(requestId, trimmed)
    if (!renamed.ok) {
      setFailedRename({ tabId, error: renamed.error })
      return
    }
    setFailedRename(null)
    if (renamed.value) {
      await adoptRename(requestId)
    }
  }

  function onNameKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'Enter') {
      event.currentTarget.blur()
    }
  }

  return (
    <section aria-label="Info" className={styles.info}>
      {INFO_HEADING}
      <div className={styles.field}>
        <label className={styles.field}>
          Name
          <input
            name="request-name"
            autoComplete="off"
            key={`${requestId}:${name}`}
            className={styles.name}
            defaultValue={name}
            spellCheck={false}
            aria-invalid={nameError !== null}
            aria-describedby={nameError ? nameErrorId : undefined}
            onChange={() => setFailedRename(null)}
            onBlur={(event) => void commitName(event.target.value)}
            onKeyDown={onNameKeyDown}
          />
        </label>
        {nameError ? (
          <p id={nameErrorId} role="alert" className={styles.fieldError}>
            <NoTranslateText text={withNextStep(nameError.description || nameError.code, RENAME_NEXT_STEP)} terms={[nameError.code]} />
          </p>
        ) : null}
      </div>
      <div className={styles.field}>
        {LOCATION_LABEL}
        {location.length > 0 ? (
          <nav aria-label="Location" className={styles.breadcrumb}>
            {location.join(' / ')}
          </nav>
        ) : (
          <span className={styles.muted}>{requestId ? 'Unknown' : 'Not saved yet'}</span>
        )}
      </div>
      <div className={styles.field}>
        <div className={styles.descriptionHeader}>
          {editing ? (
            <label htmlFor={descriptionId} className={styles.label}>
              Description
            </label>
          ) : (
            DESCRIPTION_LABEL
          )}
          <Button variant="ghost" size="sm" onClick={() => setEditing((current) => !current)}>
            {editing ? 'Done Editing' : 'Edit Description'}
          </Button>
        </div>
        {editing ? (
          <textarea
            id={descriptionId}
            name="description"
            autoComplete="off"
            className={styles.textarea}
            value={description}
            onChange={(event) => setDescription(tabId, event.target.value)}
          />
        ) : description.trim() === '' ? (
          NO_DESCRIPTION
        ) : (
          <Suspense fallback={LOADING_DESCRIPTION}>
            <MarkdownView text={description} />
          </Suspense>
        )}
      </div>
    </section>
  )
}
