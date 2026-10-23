import { useState } from 'react'
import { useShallow } from 'zustand/react/shallow'
import type { SavedAlias } from '@/platform/mappers/request'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { Swatch } from '@/shared/ui/swatch/Swatch'
import { useEnvironmentsStore } from '@/features/environments/store'
import { aliasEnvironmentLabels, isAliasAvailable } from '@/features/request-builder/lib/alias-availability'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useTabsStore } from '@/features/tabs/store'
import { AliasEnvironmentsDialog } from './AliasEnvironmentsDialog'
import { CreateAliasDialog } from './CreateAliasDialog'
import { DeleteAliasDialog } from './DeleteAliasDialog'
import styles from './AliasMenu.module.css'

const DEFAULT_LABEL = 'Default'
const NO_ENVIRONMENT = 'no environment'
const NO_ALIASES: SavedAlias[] = []

interface AliasMenuProps {
  tabId: string
}

/** The request's versions: Default, its aliases and "Create new alias…"; picking one shows it in this tab. */
export function AliasMenu({ tabId }: AliasMenuProps) {
  const aliasId = useRequestBuilderStore((state) => state.tabs[tabId]?.aliasId ?? null)
  const aliases = useRequestBuilderStore(useShallow((state) => state.tabs[tabId]?.aliases ?? NO_ALIASES))
  const requestName = useRequestBuilderStore((state) => state.tabs[tabId]?.savedDefault?.name ?? '')
  const selectVersion = useRequestBuilderStore((state) => state.selectVersion)
  const environments = useEnvironmentsStore(useShallow((state) => state.environments))
  const activeId = useEnvironmentsStore((state) => state.activeId)
  const activeName = environments.find((environment) => environment.id === activeId)?.name ?? NO_ENVIRONMENT
  const pin = useTabsStore((state) => state.pin)
  const [open, setOpen] = useState(false)
  const [actionsFor, setActionsFor] = useState<SavedAlias | null>(null)
  const [creating, setCreating] = useState(false)
  const [linking, setLinking] = useState<SavedAlias | null>(null)
  const [deleting, setDeleting] = useState<SavedAlias | null>(null)
  const shown = aliases.find((alias) => alias.id === aliasId)?.name ?? DEFAULT_LABEL

  function choose(next: string | null) {
    selectVersion(tabId, next)
    pin(tabId)
  }

  function onOpenChange(next: boolean) {
    setOpen(next)
    if (!next) {
      setActionsFor(null)
    }
  }

  function showActions(alias: SavedAlias) {
    setActionsFor(alias)
    setOpen(true)
  }

  return (
    <>
      <Dropdown.Root className={styles.root} open={open} onOpenChange={onOpenChange}>
        <Dropdown.Trigger label={`Aliases: ${shown}`} className={styles.trigger}>
          <span className={styles.caption}>Alias</span>
          <span className={styles.shown}>{shown}</span>
        </Dropdown.Trigger>
        {actionsFor ? (
          <Dropdown.Menu label={`Actions for alias ${actionsFor.name}`} className={styles.menu}>
            <div className={styles.heading} translate="no">
              {actionsFor.name}
            </div>
            <Dropdown.Item onSelect={() => setLinking(actionsFor)}>Environments…</Dropdown.Item>
            <Dropdown.Item onSelect={() => setDeleting(actionsFor)}>Delete…</Dropdown.Item>
            <div role="separator" className={styles.separator} />
            <Dropdown.Item onSelect={() => setOpen(true)}>Back to Aliases</Dropdown.Item>
          </Dropdown.Menu>
        ) : (
          <Dropdown.Menu label="Aliases" className={styles.menu}>
            <Dropdown.RadioItem checked={aliasId === null} onSelect={() => choose(null)}>
              {DEFAULT_LABEL}
            </Dropdown.RadioItem>
            {aliases.map((alias) => {
              const available = isAliasAvailable(alias, activeId)
              return (
                <div key={alias.id} className={styles.row} data-unavailable={available ? undefined : true}>
                  <Dropdown.RadioItem checked={alias.id === aliasId} disabled={!available} onSelect={() => choose(alias.id)} className={styles.name}>
                    <span translate="no">{alias.name}</span>
                    {aliasEnvironmentLabels(alias, environments).map((label) => (
                      <span key={label.id} className={styles.environment}>
                        <Swatch color={label.color} />
                        <span translate="no">{label.name}</span>
                      </span>
                    ))}
                    {available ? null : <span className="visually-hidden">, not available in {activeName}</span>}
                  </Dropdown.RadioItem>
                  <Dropdown.Item onSelect={() => showActions(alias)} className={styles.more}>
                    <span aria-hidden="true">⋯</span>
                    <span className="visually-hidden">Actions for alias {alias.name}…</span>
                  </Dropdown.Item>
                </div>
              )
            })}
            <div role="separator" className={styles.separator} />
            <Dropdown.Item onSelect={() => setCreating(true)}>Create new alias…</Dropdown.Item>
          </Dropdown.Menu>
        )}
      </Dropdown.Root>
      <CreateAliasDialog tabId={tabId} open={creating} onClose={() => setCreating(false)} />
      <AliasEnvironmentsDialog tabId={tabId} alias={linking} onClose={() => setLinking(null)} />
      <DeleteAliasDialog tabId={tabId} alias={deleting} requestName={requestName} onClose={() => setDeleting(null)} />
    </>
  )
}
