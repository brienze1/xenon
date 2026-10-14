import { SAVE_CONFLICT_CODE } from '@/platform/errors'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useItemSettingsStore } from '@/features/item-settings/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useSettingsStore } from '@/features/settings/store'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useTabTitle } from '@/features/tabs/hooks/use-tab-title'
import { environmentTabScope, useTabsStore } from '@/features/tabs/store'
import styles from './CloseTabPrompt.module.css'

const SAVE_CHANGES_TITLE = <Dialog.Title>Save Changes?</Dialog.Title>

const SAVE_NEXT_STEP = 'Try again, or choose Don’t Save to close the tab without your changes.'
const CONFLICT_NEXT_STEP = 'Choose Cancel, then keep your version or load the one from disk in the tab.'

export function CloseTabPrompt() {
  const tabId = useTabsStore((state) => state.closePrompt)
  const closeError = useTabsStore((state) => state.closeError)
  const resolveClose = useTabsStore((state) => state.resolveClose)
  const name = useTabTitle(tabId)
  const scope = useTabsStore((state) => environmentTabScope(state.tabs.find((tab) => tab.id === tabId)))
  const requestSaving = useRequestBuilderStore((state) => (tabId ? state.tabs[tabId]?.saving === true : false))
  const variablesSaving = useEnvironmentsStore((state) => (scope === null ? false : state.drafts[scope]?.saving === true))
  const itemId = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.itemId ?? null)
  const settingsSaving = useItemSettingsStore((state) => (itemId === null ? false : state.items[itemId]?.saving === true))
  const isSettingsTab = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.kind === 'settings')
  const appSettingsSaving = useSettingsStore((state) => isSettingsTab && state.saving)
  const saving = requestSaving || variablesSaving || settingsSaving || appSettingsSaving

  return (
    <Dialog.Root open={tabId !== null} onOpenChange={(open) => (open ? undefined : void resolveClose('cancel'))}>
      {SAVE_CHANGES_TITLE}
      <Dialog.Body>
        “<span translate="no">{name}</span>” has unsaved changes. Do you want to save them before closing?
        {closeError ? (
          <Banner.Root tone="error" className={styles.error}>
            <Banner.Message>
              <NoTranslateText text={withNextStep(closeError.description || closeError.code, closeError.code === SAVE_CONFLICT_CODE ? CONFLICT_NEXT_STEP : SAVE_NEXT_STEP)} terms={[closeError.code]} />
            </Banner.Message>
          </Banner.Root>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={() => void resolveClose('cancel')}>
          Cancel
        </Button>
        <Button variant="secondary" onClick={() => void resolveClose('discard')}>
          Don’t Save
        </Button>
        <Button variant="primary" disabled={saving} onClick={() => void resolveClose('save')}>
          {saving ? 'Saving…' : 'Save Changes'}
        </Button>
      </Dialog.Actions>
    </Dialog.Root>
  )
}
