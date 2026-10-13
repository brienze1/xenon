import type { Cookie } from '@/platform/mappers/cookie'
import { formatDateTime } from '@/shared/lib/date-format'
import styles from './ResponseHeaders.module.css'

const NO_COOKIES = <p className={styles.empty}>The response set no cookies</p>

const EXPIRES_FORMAT: Intl.DateTimeFormatOptions = { dateStyle: 'medium', timeStyle: 'medium', timeZone: 'UTC' }

const COOKIES_HEAD = (
  <thead>
    <tr>
      <th>Name</th>
      <th>Value</th>
      <th>Domain</th>
      <th>Path</th>
      <th>Expires</th>
      <th>HttpOnly</th>
      <th>Secure</th>
    </tr>
  </thead>
)

function expiry(expires: string | null): string {
  return expires ? `${formatDateTime(new Date(expires), EXPIRES_FORMAT)} UTC` : 'Session'
}

interface ResponseCookiesTabProps {
  cookies: Cookie[]
}

export function ResponseCookiesTab({ cookies }: ResponseCookiesTabProps) {
  if (cookies.length === 0) {
    return NO_COOKIES
  }
  return (
    <table aria-label="Response cookies" className={styles.table}>
      {COOKIES_HEAD}
      <tbody>
        {cookies.map((cookie) => (
          <tr key={`${cookie.domain};${cookie.path};${cookie.name}`}>
            <td translate="no">{cookie.name}</td>
            <td translate="no">{cookie.value}</td>
            <td translate="no">{cookie.domain}</td>
            <td translate="no">{cookie.path}</td>
            <td>{expiry(cookie.expires)}</td>
            <td>{cookie.httpOnly ? 'Yes' : 'No'}</td>
            <td>{cookie.secure ? 'Yes' : 'No'}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
