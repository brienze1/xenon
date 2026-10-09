import type { ReactNode } from 'react'
import styles from './TopBar.module.css'

const WORDMARK = (
  <span className={styles.wordmark} aria-hidden="true" translate="no">
    Xenon
  </span>
)

interface TopBarProps {
  children?: ReactNode
}

export function TopBar({ children }: TopBarProps) {
  return (
    <header className={styles.topBar}>
      {WORDMARK}
      {children}
    </header>
  )
}
