import type { ReactNode } from 'react'
import styles from './Sidebar.module.css'

export function Sidebar({ children }: { children?: ReactNode }) {
  return (
    <nav aria-label="Sidebar" className={styles.sidebar}>
      {children}
    </nav>
  )
}
