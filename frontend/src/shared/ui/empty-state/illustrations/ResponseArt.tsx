import styles from './Illustration.module.css'
import { Spray } from './Spray'

const RESPONSE_ART = (
  <svg className={styles.art} viewBox="0 0 120 96" aria-hidden="true">
    <circle className={styles.cyan} cx="60" cy="46" r="30" />
    <path className={styles.green} d="M64 22L48 50h14l-6 24 18-30H60z" />
    <path className={`${styles.cyan} ${styles.drip}`} d="M44 74v10M78 72v6" />
    <Spray />
  </svg>
)

export function ResponseArt() {
  return RESPONSE_ART
}
