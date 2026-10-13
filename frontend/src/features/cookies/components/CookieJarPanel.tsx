import { useId, useState, type FormEvent } from 'react'
import type { Cookie } from '@/platform/mappers/cookie'
import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button/Button'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useCookiesStore } from '@/features/cookies/store'
import { CookieEditor } from './CookieEditor'
import styles from './CookiesDialog.module.css'

const NEW_COOKIE = 'Cookie_1=value; Path=/'

const NO_DOMAINS = <p className={styles.empty}>No cookies yet. Responses add them, or add a domain to write one.</p>

const CROSS = (
  <svg viewBox="0 0 16 16" aria-hidden="true" className={styles.icon}>
    <path d="M4 4l8 8M12 4l-8 8" />
  </svg>
)

type Editing = { kind: 'new' } | { kind: 'cookie'; cookie: Cookie } | null

export function CookieJarPanel() {
  const domains = useCookiesStore((state) => state.domains)
  const addedDomains = useCookiesStore((state) => state.addedDomains)
  const selectedDomain = useCookiesStore((state) => state.selectedDomain)
  const selectDomain = useCookiesStore((state) => state.selectDomain)
  const addDomain = useCookiesStore((state) => state.addDomain)
  const removeDomain = useCookiesStore((state) => state.removeDomain)
  const domainFieldId = useId()
  const [domain, setDomain] = useState('')
  const names = [...domains.map((group) => group.domain), ...addedDomains]
  const cookies = domains.find((group) => group.domain === selectedDomain)?.cookies ?? []

  function submitDomain(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    addDomain(domain)
    setDomain('')
  }

  return (
    <div className={styles.jar}>
      <form className={styles.addDomain} onSubmit={submitDomain}>
        <label htmlFor={domainFieldId} className="visually-hidden">
          Domain
        </label>
        <input
          id={domainFieldId}
          className={styles.input}
          placeholder="Type a domain name, e.g. api.example.com…"
          value={domain}
          autoComplete="off"
          spellCheck={false}
          onChange={(event) => setDomain(event.target.value)}
        />
        <Button type="submit" size="sm" disabled={domain.trim() === ''}>
          Add Domain
        </Button>
      </form>
      {names.length === 0 ? (
        NO_DOMAINS
      ) : (
        <ul aria-label="Domains" className={styles.domains}>
          {names.map((name) => (
            <li key={name} className={styles.domainRow}>
              <button
                type="button"
                aria-pressed={name === selectedDomain}
                className={cn(styles.domain, name === selectedDomain && styles.selected)}
                onClick={() => selectDomain(name)}
              >
                <span translate="no">{name}</span>{' '}
                <span className={styles.count}>{countLabel(domains.find((group) => group.domain === name)?.cookies.length ?? 0)}</span>
              </button>
              <IconButton label={`Delete domain ${name}`} onClick={() => void removeDomain(name)}>
                {CROSS}
              </IconButton>
            </li>
          ))}
        </ul>
      )}
      {selectedDomain ? <DomainCookies key={selectedDomain} domain={selectedDomain} cookies={cookies} /> : null}
    </div>
  )
}

function countLabel(count: number): string {
  return count === 1 ? '1 cookie' : `${count} cookies`
}

interface DomainCookiesProps {
  domain: string
  cookies: Cookie[]
}

function DomainCookies({ domain, cookies }: DomainCookiesProps) {
  const saveCookie = useCookiesStore((state) => state.saveCookie)
  const removeCookie = useCookiesStore((state) => state.removeCookie)
  const [editing, setEditing] = useState<Editing>(null)

  async function save(raw: string) {
    const replaces = editing?.kind === 'cookie' ? editing.cookie : undefined
    const failure = await saveCookie(domain, raw, replaces)
    if (!failure) {
      setEditing(null)
    }
    return failure
  }

  return (
    <section aria-label={`Cookies of ${domain}`} className={styles.cookies}>
      <ul className={styles.chips}>
        {cookies.map((cookie) => (
          <li key={`${cookie.path};${cookie.name}`} className={styles.chip}>
            <button type="button" className={styles.chipName} onClick={() => setEditing({ kind: 'cookie', cookie })}>
              <span translate="no">{cookie.name}</span>
            </button>
            <IconButton label={`Delete cookie ${cookie.name}`} onClick={() => void removeCookie(cookie)}>
              {CROSS}
            </IconButton>
          </li>
        ))}
      </ul>
      {editing ? (
        <CookieEditor
          key={editing.kind === 'cookie' ? editing.cookie.raw : 'new'}
          initial={editing.kind === 'cookie' ? editing.cookie.raw : NEW_COOKIE}
          onSave={save}
          onCancel={() => setEditing(null)}
        />
      ) : (
        <Button size="sm" variant="ghost" onClick={() => setEditing({ kind: 'new' })}>
          Add Cookie
        </Button>
      )}
    </section>
  )
}
