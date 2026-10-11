import { useCookiesStore } from '@/features/cookies/store'
import styles from './CookiesDialog.module.css'

const COOKIE_ICON = (
  <svg viewBox="0 0 16 16" aria-hidden="true" className={styles.icon}>
    <path d="M8 2a6 6 0 1 0 6 6 2 2 0 0 1-2-2 2 2 0 0 1-2-2 2 2 0 0 1-2-2z" />
    <circle cx="6" cy="7" r="0.75" />
    <circle cx="9" cy="10" r="0.75" />
    <circle cx="5.5" cy="10.5" r="0.75" />
  </svg>
)

export function CookiesButton() {
  const openDialog = useCookiesStore((state) => state.openDialog)
  return (
    <button type="button" className={styles.footerButton} onClick={() => void openDialog()}>
      {COOKIE_ICON}
      <span>Cookies</span>
    </button>
  )
}
