import type { ReactNode } from 'react'
import styles from './TopBar.module.css'

interface TopBarProps {
  children?: ReactNode
}

export function TopBar({ children }: TopBarProps) {
  return (
    <header className={styles.topBar}>
      <span className={styles.wordmark} aria-hidden="true">
        Xenon
      </span>
      {children}
    </header>
  )
}
