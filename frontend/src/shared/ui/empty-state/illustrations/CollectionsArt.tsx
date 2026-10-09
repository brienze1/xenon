import styles from './Illustration.module.css'
import { Spray } from './Spray'

const COLLECTIONS_ART = (
  <svg className={styles.art} viewBox="0 0 120 96" aria-hidden="true">
    <path className={styles.magenta} d="M30 20h56v14H30z" />
    <path className={styles.cyan} d="M24 40h68v14H24z" />
    <path className={styles.green} d="M18 60h80v14H18z" />
    <path className={`${styles.green} ${styles.drip}`} d="M40 74v10M76 74v6" />
    <Spray />
  </svg>
)

export function CollectionsArt() {
  return COLLECTIONS_ART
}
