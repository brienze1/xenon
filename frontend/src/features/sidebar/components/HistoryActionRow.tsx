import { useEffect, useState } from 'react'
import type { AppError } from '@/platform/errors'
import type { HistoryActivity, HistoryEntry } from '@/platform/mappers/history'
import { cn } from '@/shared/lib/cn'
import { formatDuration } from '@/shared/lib/format-duration'
import { sentTime } from '@/features/sidebar/lib/history-rows'
import { activityName, activityOutcome, aiTag } from '@/features/sidebar/lib/history-labels'
import { useSidebarStore } from '@/features/sidebar/store'
import { SidebarErrorText } from './SidebarErrorText'
import styles from './HistoryView.module.css'

interface HistoryActionRowProps {
  entry: HistoryEntry
  expanded: boolean
  onToggle: (id: string) => void
}

/** An AI tool call other than a send: what it did, to what, how it ended; expands to its masked arguments. */
export function HistoryActionRow({ entry, expanded, onToggle }: HistoryActionRowProps) {
  const activity = entry.activity
  const name = activityName(activity)
  const result = activityOutcome(activity)
  const time = sentTime(entry.sentAt)
  const target = activity?.target?.name ?? null
  const durationMs = activity?.durationMs ?? entry.totalMs
  const detailsId = `history-args-${entry.id}`
  return (
    <div className={styles.action}>
      <button
        type="button"
        className={styles.entry}
        aria-expanded={expanded}
        aria-controls={expanded ? detailsId : undefined}
        aria-label={`${aiTag(entry.client)} — ${name}${target ? ` ${target}` : ''}, ${result.label}, ${formatDuration(durationMs)}, ${time}`}
        onClick={() => onToggle(entry.id)}
      >
        <span className={styles.line}>
          <span className={styles.aiTag} translate="no">
            {aiTag(entry.client)}
          </span>
          <span className={styles.tool} translate="no">
            {name}
          </span>
          {target ? (
            <span className={styles.target} title={activity?.target?.path ?? target} translate="no">
              {target}
            </span>
          ) : null}
        </span>
        <span className={cn(styles.meta, styles.actionMeta)}>
          <span className={cn(styles.status, styles[result.tone])} translate="no">
            {result.label}
          </span>
          <span>{formatDuration(durationMs)}</span>
          <time dateTime={entry.sentAt}>{time}</time>
        </span>
      </button>
      {expanded ? <ActivityArguments id={detailsId} entryId={entry.id} /> : null}
    </div>
  )
}

interface ActivityArgumentsProps {
  id: string
  entryId: string
}

type ArgumentsState = { status: 'loading' } | { status: 'ready'; activity: HistoryActivity | null } | { status: 'failed'; error: AppError }

function ActivityArguments({ id, entryId }: ActivityArgumentsProps) {
  const loadActivity = useSidebarStore((state) => state.loadHistoryActivity)
  const [state, setState] = useState<ArgumentsState>({ status: 'loading' })

  useEffect(() => {
    let current = true
    void loadActivity(entryId).then((loaded) => {
      if (current) {
        setState(loaded.ok ? { status: 'ready', activity: loaded.value } : { status: 'failed', error: loaded.error })
      }
    })
    return () => {
      current = false
    }
  }, [entryId, loadActivity])

  return (
    <div id={id} className={styles.args} aria-label="Arguments (masked)" role="region">
      {state.status === 'loading' ? (
        <p className={styles.argsNote}>Loading…</p>
      ) : state.status === 'failed' ? (
        <p role="alert" className={styles.argsNote}>
          <SidebarErrorText error={state.error} nextStep="Try again." />
        </p>
      ) : state.activity?.args ? (
        <pre className={styles.argsText} translate="no">
          {state.activity.args}
        </pre>
      ) : (
        <p className={styles.argsNote}>No arguments.</p>
      )}
    </div>
  )
}
