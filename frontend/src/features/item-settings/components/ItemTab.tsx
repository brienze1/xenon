import { useEffect, type ReactNode } from 'react'
import type { AppError } from '@/platform/errors'
import type { ContainerKind, ItemSettings } from '@/platform/mappers/tree'
import { collectionScope } from '@/platform/mappers/variable'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import type { ItemAuthSlot, ItemScriptsSlot, ItemSection } from '@/features/item-settings/model'
import { isDocumentationDirty, useDocumentationStore } from '@/features/documentation/store'
import { isScopeDirty, useEnvironmentsStore } from '@/features/environments/store'
import { isItemDirty, useItemSettingsStore } from '@/features/item-settings/store'
import { ItemHeaders } from './ItemHeaders'
import { ItemOverview } from './ItemOverview'
import styles from './ItemTab.module.css'

const LOAD_NEXT_STEP = 'Check that the workspace folder exists, then try again.'

const LOADING = (
  <p role="status" className={styles.note}>
    Loading…
  </p>
)

const STALE_MESSAGE = <Banner.Message>These settings changed on disk while you were editing them.</Banner.Message>

const OVERVIEW_TAB = <Tabs.Tab value="overview">Overview</Tabs.Tab>
const AUTH_TAB = <Tabs.Tab value="auth">Authorization</Tabs.Tab>
const HEADERS_TAB = <Tabs.Tab value="headers">Headers</Tabs.Tab>
const SCRIPTS_TAB = <Tabs.Tab value="scripts">Scripts</Tabs.Tab>
const VARIABLES_TAB = <Tabs.Tab value="variables">Variables</Tabs.Tab>
const DOCUMENTATION_TAB = <Tabs.Tab value="documentation">Documentation</Tabs.Tab>

const KIND_LABEL: Record<ContainerKind, string> = { collection: 'Collection', folder: 'Folder' }

interface ErrorTextProps {
  error: AppError
  nextStep: string
}

function ErrorText({ error, nextStep }: ErrorTextProps) {
  return <NoTranslateText text={withNextStep(error.description || error.code, nextStep)} terms={[error.code]} />
}

interface ItemTabProps {
  itemId: string
  kind: ContainerKind
  /** Saves the tab (settings and, for a collection, its variables); the same as Cmd/Ctrl+S. */
  onSave: () => void
  /** Renders the auth editor (inherit, none or any auth type); composed by the app. */
  authEditor?: (slot: ItemAuthSlot) => ReactNode
  /** Renders the pre-request and post-response scripts editor; composed by the app. */
  scriptsEditor?: (slot: ItemScriptsSlot) => ReactNode
  /** Renders a collection's variables table; composed by the app. Folders have no variables. */
  collectionVariables?: (collectionId: string) => ReactNode
  /** Renders a collection's OpenAPI documentation; composed by the app. Folders have none. */
  documentation?: (collectionId: string) => ReactNode
}

/**
 * The main-area tab of a collection or folder, like Postman's: an Overview (name and markdown
 * description) and the defaults its requests inherit — authorization, headers, scripts and, for
 * a collection, variables.
 */
export function ItemTab({ itemId, kind, onSave, authEditor, scriptsEditor, collectionVariables, documentation }: ItemTabProps) {
  const open = useItemSettingsStore((state) => state.open)
  const load = useItemSettingsStore((state) => state.load)
  const status = useItemSettingsStore((state) => state.items[itemId]?.status ?? 'loading')
  const loadError = useItemSettingsStore((state) => state.items[itemId]?.loadError ?? null)
  const settings = useItemSettingsStore((state) => state.items[itemId]?.draft ?? null)

  useEffect(() => {
    void open(itemId)
  }, [itemId, open])

  if (settings === null) {
    if (status === 'ready' && loadError) {
      return (
        <div className={styles.tab}>
          <Banner.Root tone="error">
            <Banner.Message>
              <ErrorText error={loadError} nextStep={LOAD_NEXT_STEP} />
            </Banner.Message>
            <Banner.Actions>
              <Button variant="secondary" size="sm" onClick={() => void load(itemId)}>
                Try Again
              </Button>
            </Banner.Actions>
          </Banner.Root>
        </div>
      )
    }
    return <div className={styles.tab}>{LOADING}</div>
  }
  return (
    <ItemEditor
      settings={settings}
      kind={kind}
      onSave={onSave}
      authEditor={authEditor}
      scriptsEditor={scriptsEditor}
      collectionVariables={collectionVariables}
      documentation={documentation}
    />
  )
}

interface ItemEditorProps extends Omit<ItemTabProps, 'itemId'> {
  settings: ItemSettings
}

function ItemEditor({ settings, kind, onSave, authEditor, scriptsEditor, collectionVariables, documentation }: ItemEditorProps) {
  const itemId = settings.id
  const section = useItemSettingsStore((state) => state.items[itemId]?.section ?? 'overview')
  const saving = useItemSettingsStore((state) => state.items[itemId]?.saving === true)
  const settingsDirty = useItemSettingsStore((state) => isItemDirty(state, itemId))
  const variablesDirty = useEnvironmentsStore((state) => kind === 'collection' && isScopeDirty(state, collectionScope(itemId)))
  const documentationDirty = useDocumentationStore((state) => kind === 'collection' && isDocumentationDirty(state, itemId))
  const dirty = settingsDirty || variablesDirty || documentationDirty
  const stale = useItemSettingsStore((state) => state.items[itemId]?.stale === true)
  const failure = useItemSettingsStore((state) => state.items[itemId]?.failure ?? null)
  const edit = useItemSettingsStore((state) => state.edit)
  const setSection = useItemSettingsStore((state) => state.setSection)
  const discard = useItemSettingsStore((state) => state.discard)
  const label = KIND_LABEL[kind]
  const variables = kind === 'collection' && collectionVariables ? collectionVariables(itemId) : null
  const docs = kind === 'collection' && documentation ? documentation(itemId) : null
  const shown: ItemSection = (section === 'variables' && variables === null) || (section === 'documentation' && docs === null) ? 'overview' : section

  return (
    <section className={styles.tab} aria-label={`${settings.name} ${label.toLowerCase()}`}>
      <header className={styles.header}>
        <h2 className={styles.title} translate="no" title={settings.name}>
          {settings.name}
        </h2>
        <span className={styles.kind}>{label}</span>
        <div className={styles.actions}>
          <Button variant="primary" size="sm" disabled={!dirty || saving} onClick={onSave}>
            {saving ? 'Saving…' : 'Save'}
          </Button>
        </div>
      </header>
      {stale ? (
        <Banner.Root tone="warning">
          {STALE_MESSAGE}
          <Banner.Actions>
            <Button variant="secondary" size="sm" onClick={onSave}>
              Keep Your Version
            </Button>
            <Button variant="ghost" size="sm" onClick={() => discard(itemId)}>
              Load from Disk
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      {failure ? (
        <Banner.Root tone="error">
          <Banner.Message>
            <ErrorText error={failure.error} nextStep={failure.nextStep} />
          </Banner.Message>
        </Banner.Root>
      ) : null}
      <Tabs.Root value={shown} onValueChange={(next) => setSection(itemId, next as ItemSection)} className={styles.sections}>
        <Tabs.List label={`${label} settings`}>
          {OVERVIEW_TAB}
          {AUTH_TAB}
          {HEADERS_TAB}
          {SCRIPTS_TAB}
          {variables ? VARIABLES_TAB : null}
          {docs ? DOCUMENTATION_TAB : null}
        </Tabs.List>
        <Tabs.Panel value="overview" className={styles.panel}>
          <ItemOverview settings={settings} label={label} />
        </Tabs.Panel>
        <Tabs.Panel value="auth" className={styles.panel}>
          <p className={styles.note}>
            Requests below this {label.toLowerCase()} that inherit their authorization use this one.
          </p>
          {authEditor?.({ itemId, kind: settings.kind, auth: settings.auth, onChange: (auth) => edit(itemId, { auth }) })}
        </Tabs.Panel>
        <Tabs.Panel value="headers" className={styles.panel}>
          <ItemHeaders settings={settings} />
        </Tabs.Panel>
        <Tabs.Panel value="scripts" className={styles.scriptsPanel}>
          {scriptsEditor?.({
            itemId,
            scripts: settings.scripts,
            inheritScripts: settings.inheritScripts,
            onChange: (scripts) => edit(itemId, { scripts }),
            onInheritChange: (inheritScripts) => edit(itemId, { inheritScripts }),
          })}
        </Tabs.Panel>
        {variables ? (
          <Tabs.Panel value="variables" className={styles.panel}>
            {variables}
          </Tabs.Panel>
        ) : null}
        {docs ? (
          <Tabs.Panel value="documentation" className={styles.documentationPanel}>
            {docs}
          </Tabs.Panel>
        ) : null}
      </Tabs.Root>
    </section>
  )
}
