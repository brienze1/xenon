import type { ReactNode } from 'react'
import styles from './Toast.module.css'

const CLOSE_GLYPH = <span aria-hidden="true">×</span>

interface RootProps {
  /** The close (×) button's accessible name. */
  closeLabel?: string
  onClose: () => void
  children: ReactNode
}

/** A notice fixed to the bottom-left, above the footer. It never dismisses itself. */
function Root({ closeLabel = 'Close', onClose, children }: RootProps) {
  return (
    <div role="status" aria-live="polite" className={styles.root}>
      <div className={styles.body}>{children}</div>
      <button type="button" aria-label={closeLabel} title={closeLabel} className={styles.close} onClick={onClose}>
        {CLOSE_GLYPH}
      </button>
    </div>
  )
}

interface MessageProps {
  children: ReactNode
}

function Message({ children }: MessageProps) {
  return <p className={styles.message}>{children}</p>
}

interface ActionsProps {
  children: ReactNode
}

function Actions({ children }: ActionsProps) {
  return <div className={styles.actions}>{children}</div>
}

export const Toast = { Root, Message, Actions }
