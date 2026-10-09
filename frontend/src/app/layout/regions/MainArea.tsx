import type { ReactNode } from 'react'
import { MAIN_CONTENT_ID } from '@/app/layout/shell-layout'
import styles from './MainArea.module.css'

interface MainAreaProps {
  children?: ReactNode
}

export function MainArea({ children }: MainAreaProps) {
  return (
    <main id={MAIN_CONTENT_ID} tabIndex={-1} className={styles.main}>
      {children}
    </main>
  )
}
