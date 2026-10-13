import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { useTabsStore } from '@/features/tabs/store'
import styles from './TabNotice.module.css'

const CHANGED_MESSAGE = <Banner.Message>This request changed on disk while you were editing it.</Banner.Message>

const DELETED_MESSAGE = <Banner.Message>This request no longer exists on disk. Your unsaved changes are still here.</Banner.Message>

interface TabNoticeProps {
  tabId: string
}

export function TabNotice({ tabId }: TabNoticeProps) {
  const notice = useTabsStore((state) => state.notices[tabId])
  const keepMine = useTabsStore((state) => state.keepMine)
  const loadFromDisk = useTabsStore((state) => state.loadFromDisk)
  const keepAsNew = useTabsStore((state) => state.keepAsNew)
  const requestClose = useTabsStore((state) => state.requestClose)

  if (notice === undefined) {
    return null
  }
  return notice === 'changed' ? (
    <Banner.Root tone="warning" className={styles.notice}>
      {CHANGED_MESSAGE}
      <Banner.Actions>
        <Button variant="secondary" size="sm" onClick={() => void keepMine(tabId)}>
          Keep Your Version
        </Button>
        <Button variant="ghost" size="sm" onClick={() => void loadFromDisk(tabId)}>
          Load from Disk
        </Button>
      </Banner.Actions>
    </Banner.Root>
  ) : (
    <Banner.Root tone="warning" className={styles.notice}>
      {DELETED_MESSAGE}
      <Banner.Actions>
        <Button variant="secondary" size="sm" onClick={() => keepAsNew(tabId)}>
          Keep as Unsaved Tab
        </Button>
        <Button variant="ghost" size="sm" onClick={() => requestClose(tabId)}>
          Close Tab
        </Button>
      </Banner.Actions>
    </Banner.Root>
  )
}
