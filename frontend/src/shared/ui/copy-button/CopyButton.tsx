import { useEffect, useRef, useState, type ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import styles from './CopyButton.module.css'

/** How long the check icon shows after a successful copy. */
export const COPIED_FOR_MS = 1500

type CopyState = 'idle' | 'copied' | 'failed'

const TITLES: Record<Exclude<CopyState, 'idle'>, string> = { copied: 'Copied', failed: 'Copy failed' }

const COPY_ICON = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.icon}>
    <rect x="8" y="8" width="14" height="14" rx="2" ry="2" />
    <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
  </svg>
)

const CHECK_ICON = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.icon}>
    <path d="M20 6 9 17l-5-5" />
  </svg>
)

const FAILED_ICON = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.icon}>
    <path d="M18 6 6 18" />
    <path d="m6 6 12 12" />
  </svg>
)

const ICONS: Record<CopyState, ReactNode> = { idle: COPY_ICON, copied: CHECK_ICON, failed: FAILED_ICON }

interface CopyButtonProps {
  /** The accessible name, e.g. "Copy body". The button never shows text. */
  label: string
  /** Copies; resolves true when the clipboard took the value. A rejection counts as a failure. */
  onCopy: () => Promise<boolean>
  disabled?: boolean
  copiedMessage?: ReactNode
  failedMessage?: ReactNode
  className?: string
}

export function CopyButton({
  label,
  onCopy,
  disabled,
  copiedMessage = 'Copied to the clipboard.',
  failedMessage = 'Couldn’t copy to the clipboard.',
  className,
}: CopyButtonProps) {
  const [state, setState] = useState<CopyState>('idle')
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  useEffect(() => () => clearTimeout(timer.current), [])

  async function copy() {
    clearTimeout(timer.current)
    let copied = false
    try {
      copied = await onCopy()
    } catch {
      copied = false
    }
    setState(copied ? 'copied' : 'failed')
    if (copied) {
      timer.current = setTimeout(() => setState('idle'), COPIED_FOR_MS)
    }
  }

  return (
    <>
      <IconButton
        label={label}
        title={state === 'idle' ? label : TITLES[state]}
        data-state={state}
        disabled={disabled}
        className={cn(styles.copyButton, className)}
        onClick={() => void copy()}
      >
        {ICONS[state]}
      </IconButton>
      <span aria-live="polite" className="visually-hidden">
        {state === 'copied' ? copiedMessage : state === 'failed' ? failedMessage : null}
      </span>
    </>
  )
}
