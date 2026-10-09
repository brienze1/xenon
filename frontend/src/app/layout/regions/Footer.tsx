import type { ReactNode } from 'react'
import styles from './Footer.module.css'

export function Footer({ children }: { children?: ReactNode }) {
  return <footer className={styles.footer}>{children}</footer>
}
