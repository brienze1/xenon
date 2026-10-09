import styles from './Illustration.module.css'
import { Spray } from './Spray'

export function WorkspaceArt() {
  return (
    <svg className={styles.art} viewBox="0 0 120 96" aria-hidden="true">
      <path className={styles.cyan} d="M18 30h28l8 9h46v41H18z" />
      <path className={`${styles.cyan} ${styles.drip}`} d="M34 80v8M70 80v5" />
      <path className={styles.magenta} d="M60 47l4 10 11 1-8 7 3 11-10-6-10 6 3-11-8-7 11-1z" />
      <Spray />
    </svg>
  )
}
