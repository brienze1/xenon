import styles from './Illustration.module.css'
import { Spray } from './Spray'

export function TabsArt() {
  return (
    <svg className={styles.art} viewBox="0 0 120 96" aria-hidden="true">
      <path className={styles.cyan} d="M16 34h26l4-10h26l4 10h28v44H16z" />
      <path className={styles.magenta} d="M46 52h28M60 38v28" />
      <path className={`${styles.cyan} ${styles.drip}`} d="M30 78v8M88 78v5" />
      <Spray />
    </svg>
  )
}
