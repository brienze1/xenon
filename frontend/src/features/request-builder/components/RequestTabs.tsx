import { Tabs } from '@/shared/ui/tabs/Tabs'
import type { SubTab } from '@/features/request-builder/model'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { AuthPanel } from './AuthPanel'
import { BodyPanel } from './BodyPanel'
import { HeadersPanel } from './HeadersPanel'
import { ParamsPanel } from './ParamsPanel'
import styles from './RequestTabs.module.css'
import { ScriptsPanel } from './ScriptsPanel'
import { SettingsPanel } from './SettingsPanel'

interface RequestTabsProps {
  tabId: string
}

export function RequestTabs({ tabId }: RequestTabsProps) {
  const subTab = useRequestBuilderStore((state) => state.tabs[tabId]?.subTab ?? 'params')
  const headerCount = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.headers.filter((row) => row.enabled).length ?? 0)
  const setSubTab = useRequestBuilderStore((state) => state.setSubTab)

  return (
    <Tabs.Root value={subTab} onValueChange={(value) => setSubTab(tabId, value as SubTab)} className={styles.tabs}>
      <Tabs.List label="Request">
        <Tabs.Tab value="params">Params</Tabs.Tab>
        <Tabs.Tab value="auth">Authorization</Tabs.Tab>
        <Tabs.Tab value="headers">Headers{headerCount > 0 ? ` (${headerCount})` : ''}</Tabs.Tab>
        <Tabs.Tab value="body">Body</Tabs.Tab>
        <Tabs.Tab value="scripts">Scripts</Tabs.Tab>
        <Tabs.Tab value="settings">Settings</Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="params" className={styles.panel}>
        <ParamsPanel tabId={tabId} />
      </Tabs.Panel>
      <Tabs.Panel value="auth" className={styles.panel}>
        <AuthPanel tabId={tabId} />
      </Tabs.Panel>
      <Tabs.Panel value="headers" className={styles.panel}>
        <HeadersPanel tabId={tabId} />
      </Tabs.Panel>
      <Tabs.Panel value="body" className={styles.panel}>
        <BodyPanel tabId={tabId} />
      </Tabs.Panel>
      <Tabs.Panel value="scripts" className={styles.panel}>
        <ScriptsPanel />
      </Tabs.Panel>
      <Tabs.Panel value="settings" className={styles.panel}>
        <SettingsPanel tabId={tabId} />
      </Tabs.Panel>
    </Tabs.Root>
  )
}
