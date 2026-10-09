import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { useTabsStore } from '@/features/tabs/store'
import styles from './TabNotice.module.css'

interface TabNoticeProps {
  tabId: string
}

export function TabNotice({ tabId }: TabNoticeProps) {
  const notice = useTabsStore((state) => state.notices[tabId])
  const keepMine = useTabsStore((state) => state.keepMine)
  const loadFromDisk = useTabsStore((state) => state.loadFromDisk)
  const keepAsNew = useTabsStore((state) => state.keepAsNew)
  const closeTab = useTabsStore((state) => state.closeTab)

  if (notice === undefined) {
    return null
  }
  return notice === 'changed' ? (
    <Banner.Root tone="warning" className={styles.notice}>
      <Banner.Message>This request changed on disk while you were editing it.</Banner.Message>
      <Banner.Actions>
        <Button variant="secondary" size="sm" onClick={() => void keepMine(tabId)}>
          Keep mine
        </Button>
        <Button variant="ghost" size="sm" onClick={() => void loadFromDisk(tabId)}>
          Load from disk
        </Button>
      </Banner.Actions>
    </Banner.Root>
  ) : (
    <Banner.Root tone="warning" className={styles.notice}>
      <Banner.Message>This request was deleted. Your unsaved changes are still here.</Banner.Message>
      <Banner.Actions>
        <Button variant="secondary" size="sm" onClick={() => keepAsNew(tabId)}>
          Keep as unsaved tab
        </Button>
        <Button variant="ghost" size="sm" onClick={() => closeTab(tabId)}>
          Close
        </Button>
      </Banner.Actions>
    </Banner.Root>
  )
}
