import type { ReactNode } from 'react'
import { formatNumber } from '@/shared/lib/number-format'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import type { RequestBuilderMode, SubTab } from '@/features/request-builder/model'
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
  mode?: RequestBuilderMode
  scripts?: ReactNode
}

const PARAMS_TAB = <Tabs.Tab value="params">Params</Tabs.Tab>

const AUTH_TAB = <Tabs.Tab value="auth">Authorization</Tabs.Tab>

const BODY_TAB = <Tabs.Tab value="body">Body</Tabs.Tab>

const SCRIPTS_TAB = <Tabs.Tab value="scripts">Scripts</Tabs.Tab>

const SETTINGS_TAB = <Tabs.Tab value="settings">Settings</Tabs.Tab>

export function RequestTabs({ tabId, mode = 'request', scripts }: RequestTabsProps) {
  const stored = useRequestBuilderStore((state) => state.tabs[tabId]?.subTab ?? 'params')
  const full = mode === 'request'
  const subTab = !full && (stored === 'scripts' || stored === 'settings') ? 'params' : stored
  const headerCount = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.headers.filter((row) => row.enabled).length ?? 0)
  const setSubTab = useRequestBuilderStore((state) => state.setSubTab)

  return (
    <Tabs.Root value={subTab} onValueChange={(value) => setSubTab(tabId, value as SubTab)} className={styles.tabs}>
      <Tabs.List label="Request">
        {PARAMS_TAB}
        {AUTH_TAB}
        <Tabs.Tab value="headers">Headers{headerCount > 0 ? ` (${formatNumber(headerCount, 0)})` : ''}</Tabs.Tab>
        {BODY_TAB}
        {full ? SCRIPTS_TAB : null}
        {full ? SETTINGS_TAB : null}
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
      {full ? (
        <Tabs.Panel value="scripts" className={styles.panel}>
          <ScriptsPanel>{scripts}</ScriptsPanel>
        </Tabs.Panel>
      ) : null}
      {full ? (
        <Tabs.Panel value="settings" className={styles.panel}>
          <SettingsPanel tabId={tabId} />
        </Tabs.Panel>
      ) : null}
    </Tabs.Root>
  )
}
