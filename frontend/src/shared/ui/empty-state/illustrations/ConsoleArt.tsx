import styles from './Illustration.module.css'
import { Spray } from './Spray'

export function ConsoleArt() {
  return (
    <svg className={styles.art} viewBox="0 0 120 96" aria-hidden="true">
      <path className={styles.magenta} d="M14 18h92v58H14z" />
      <path className={styles.green} d="M28 38l12 10-12 10M50 60h22" />
      <path className={`${styles.magenta} ${styles.drip}`} d="M36 76v9M90 76v6" />
      <Spray />
    </svg>
  )
}
