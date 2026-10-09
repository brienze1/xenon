import styles from './Panel.module.css'

export function ScriptsPanel() {
  return (
    <section className={styles.section} aria-label="Scripts">
      <h3 className={styles.heading}>Pre-request and Post-response scripts</h3>
      <p className={styles.note}>Scripts arrive in a later stage of Xenon.</p>
    </section>
  )
}
