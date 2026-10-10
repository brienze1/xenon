import type { ReactNode } from 'react'
import styles from './TopBar.module.css'

const WORDMARK = (
  <span className={styles.wordmark} aria-hidden="true" translate="no">
    Xenon
  </span>
)

interface TopBarProps {
  children?: ReactNode
  end?: ReactNode
}

export function TopBar({ children, end }: TopBarProps) {
  return (
    <header className={styles.topBar}>
      {WORDMARK}
      {children}
      {end ? <div className={styles.end}>{end}</div> : null}
    </header>
  )
}
