import type { ReactNode } from 'react'
import styles from './MainArea.module.css'

export function MainArea({ children }: { children?: ReactNode }) {
  return <main className={styles.main}>{children}</main>
}
