import type { ReactNode } from 'react'
import styles from './MainArea.module.css'

interface MainAreaProps {
  children?: ReactNode
}

export function MainArea({ children }: MainAreaProps) {
  return <main className={styles.main}>{children}</main>
}
