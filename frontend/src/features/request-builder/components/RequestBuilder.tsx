import { Banner } from '@/shared/ui/banner/Banner'
import { useRequestBuilderStore } from '../store'
import { RequestBar } from './RequestBar'
import styles from './RequestBuilder.module.css'
import { RequestTabs } from './RequestTabs'

interface RequestBuilderProps {
  tabId: string
}

export function RequestBuilder({ tabId }: RequestBuilderProps) {
  const error = useRequestBuilderStore((state) => state.tabs[tabId]?.error ?? null)
  return (
    <section aria-label="Request" className={styles.builder}>
      <RequestBar tabId={tabId} />
      {error ? (
        <Banner.Root tone="error" className={styles.banner}>
          <Banner.Message>{error.description || error.code}</Banner.Message>
        </Banner.Root>
      ) : null}
      <RequestTabs tabId={tabId} />
    </section>
  )
}
