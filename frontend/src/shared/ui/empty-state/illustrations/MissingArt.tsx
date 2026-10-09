import styles from './Illustration.module.css'
import { Spray } from './Spray'

const MISSING_ART = (
  <svg className={styles.art} viewBox="0 0 120 96" aria-hidden="true">
    <path className={styles.yellow} d="M18 30h28l8 9h46v41H18z" />
    <path className={styles.red} d="M50 52a10 10 0 1 1 14 9c-3 1.5-4 3-4 6M60 74v1" />
    <path className={`${styles.yellow} ${styles.drip}`} d="M30 80v8M88 80v5" />
    <Spray />
  </svg>
)

export function MissingArt() {
  return MISSING_ART
}
