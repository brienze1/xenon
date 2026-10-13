import { useId, useState, useTransition, type FormEvent } from 'react'
import type { AppError } from '@/platform/errors'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useCookiesStore } from '@/features/cookies/store'
import { CookieErrorText } from './CookieErrorText'
import styles from './CookiesDialog.module.css'

const ALLOWLIST_INTRO = (
  <p className={styles.intro}>
    Scripts can read and change the cookies of these domains and their subdomains through <code translate="no">pm.cookies.jar()</code>.
  </p>
)

const NO_ALLOWED_DOMAINS = <p className={styles.empty}>No domain is allowed yet.</p>

const SAVE_NEXT_STEP = 'Check that the app data folder is writable, then save again.'

const CROSS = (
  <svg viewBox="0 0 16 16" aria-hidden="true" className={styles.icon}>
    <path d="M4 4l8 8M12 4l-8 8" />
  </svg>
)

export function AllowlistPanel() {
  const saved = useCookiesStore((state) => state.allowlist)
  const saveAllowlist = useCookiesStore((state) => state.saveAllowlist)
  const fieldId = useId()
  const [domains, setDomains] = useState(saved)
  const [domain, setDomain] = useState('')
  const [error, setError] = useState<AppError | null>(null)
  const [saving, startSave] = useTransition()

  function add(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const added = domain.trim().toLowerCase()
    if (added !== '' && !domains.includes(added)) {
      setDomains([...domains, added])
    }
    setDomain('')
  }

  function save() {
    startSave(async () => {
      setError(await saveAllowlist(domains))
    })
  }

  return (
    <div className={styles.allowlist}>
      {ALLOWLIST_INTRO}
      <form className={styles.addDomain} onSubmit={add}>
        <label htmlFor={fieldId} className="visually-hidden">
          Allowed domain
        </label>
        <input
          id={fieldId}
          className={styles.input}
          placeholder="Domain scripts may access, e.g. example.com…"
          value={domain}
          autoComplete="off"
          spellCheck={false}
          onChange={(event) => setDomain(event.target.value)}
        />
        <Button type="submit" size="sm" disabled={domain.trim() === ''}>
          Add
        </Button>
      </form>
      {domains.length === 0 ? (
        NO_ALLOWED_DOMAINS
      ) : (
        <ul aria-label="Allowed domains" className={styles.domains}>
          {domains.map((allowed) => (
            <li key={allowed} className={styles.domainRow}>
              <span className={styles.allowed} translate="no">
                {allowed}
              </span>
              <IconButton label={`Remove ${allowed}`} onClick={() => setDomains(domains.filter((kept) => kept !== allowed))}>
                {CROSS}
              </IconButton>
            </li>
          ))}
        </ul>
      )}
      {error ? (
        <Banner.Root tone="error">
          <Banner.Message>
            <CookieErrorText error={error} nextStep={SAVE_NEXT_STEP} />
          </Banner.Message>
        </Banner.Root>
      ) : null}
      <div className={styles.editorActions}>
        <Button variant="primary" size="sm" disabled={saving} onClick={save}>
          {saving ? 'Saving…' : 'Save Allowlist'}
        </Button>
      </div>
    </div>
  )
}
