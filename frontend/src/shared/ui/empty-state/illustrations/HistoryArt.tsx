import styles from './Illustration.module.css'
import { Spray } from './Spray'

const HISTORY_ART = (
  <svg className={styles.art} viewBox="0 0 120 96" aria-hidden="true">
    <path className={styles.cyan} d="M32 46a28 28 0 1 0 9-20.5" />
    <path className={styles.cyan} d="M40 14v12h12" />
    <path className={styles.magenta} d="M60 30v18l12 8" />
    <path className={`${styles.cyan} ${styles.drip}`} d="M48 72v12M82 70v6" />
    <Spray />
  </svg>
)

export function HistoryArt() {
  return HISTORY_ART
}
