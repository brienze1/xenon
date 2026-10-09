import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { RequestBar } from './RequestBar'
import styles from './RequestBuilder.module.css'
import { RequestTabs } from './RequestTabs'

interface RequestBuilderProps {
  tabId: string
  onSave?: () => void
}

export function RequestBuilder({ tabId, onSave }: RequestBuilderProps) {
  const error = useRequestBuilderStore((state) => state.tabs[tabId]?.error ?? null)
  const conflict = useRequestBuilderStore((state) => state.tabs[tabId]?.conflict === true)
  const save = useRequestBuilderStore((state) => state.save)
  const openSaved = useRequestBuilderStore((state) => state.openSaved)

  function loadFromDisk() {
    const requestId = useRequestBuilderStore.getState().tabs[tabId]?.draft.id ?? ''
    void openSaved(tabId, requestId)
  }

  return (
    <section aria-label="Request" className={styles.builder}>
      <RequestBar tabId={tabId} onSave={onSave} />
      {conflict ? (
        <Banner.Root tone="warning" className={styles.banner}>
          <Banner.Message>This request changed on disk since you opened it.</Banner.Message>
          <Banner.Actions>
            <Button variant="secondary" size="sm" onClick={() => void save(tabId, true)}>
              Keep mine
            </Button>
            <Button variant="ghost" size="sm" onClick={loadFromDisk}>
              Load from disk
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      {error ? (
        <Banner.Root tone="error" className={styles.banner}>
          <Banner.Message>{error.description || error.code}</Banner.Message>
        </Banner.Root>
      ) : null}
      <RequestTabs tabId={tabId} />
    </section>
  )
}
