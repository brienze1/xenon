import type { ReactNode } from 'react'
import styles from './RightPanel.module.css'

export function RightPanel({ children }: { children?: ReactNode }) {
  return (
    <aside aria-label="Side panel" className={styles.panel}>
      {children}
    </aside>
  )
}
