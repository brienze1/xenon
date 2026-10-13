import type { ReactNode } from 'react'
import { useShellLayout } from '@/app/layout/use-shell-layout'
import styles from './MainArea.module.css'

interface MainAreaProps {
  children?: ReactNode
}

export function MainArea({ children }: MainAreaProps) {
  const { meta } = useShellLayout()
  return (
    <main id={meta.mainContentId} tabIndex={-1} className={styles.main}>
      {children}
    </main>
  )
}
