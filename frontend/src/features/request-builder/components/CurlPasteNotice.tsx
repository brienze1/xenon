import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import styles from './CurlPasteNotice.module.css'

interface CurlPasteNoticeProps {
  tabId: string
}

/** Tells that a pasted cURL command replaced the request, lists what was ignored and offers Undo. */
export function CurlPasteNotice({ tabId }: CurlPasteNoticeProps) {
  const notices = useRequestBuilderStore((state) => state.tabs[tabId]?.curlPaste?.notices ?? null)
  const undo = useRequestBuilderStore((state) => state.undoCurlPaste)
  const dismiss = useRequestBuilderStore((state) => state.dismissCurlPaste)
  if (notices === null) {
    return null
  }

  return (
    <Banner.Root tone={notices.length > 0 ? 'warning' : 'info'} className={styles.notice}>
      <Banner.Message>
        <p className={styles.title}>Replaced the request with the pasted cURL command.</p>
        {notices.length > 0 ? (
          <ul className={styles.list}>
            {notices.map((notice) => (
              <li key={notice}>
                <NoTranslateText text={notice} terms={[notice.split(' ').at(-1) ?? '']} />
              </li>
            ))}
          </ul>
        ) : null}
      </Banner.Message>
      <Banner.Actions>
        <Button variant="secondary" size="sm" onClick={() => undo(tabId)}>
          Undo
        </Button>
        <Button variant="ghost" size="sm" onClick={() => dismiss(tabId)}>
          Dismiss
        </Button>
      </Banner.Actions>
    </Banner.Root>
  )
}
