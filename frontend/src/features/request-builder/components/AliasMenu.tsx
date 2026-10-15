import { useState } from 'react'
import { useShallow } from 'zustand/react/shallow'
import type { SavedAlias } from '@/platform/mappers/request'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useTabsStore } from '@/features/tabs/store'
import { CreateAliasDialog } from './CreateAliasDialog'
import { DeleteAliasDialog } from './DeleteAliasDialog'
import styles from './AliasMenu.module.css'

const DEFAULT_LABEL = 'Default'
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
  const pin = useTabsStore((state) => state.pin)
  const [creating, setCreating] = useState(false)
  const [deleting, setDeleting] = useState<SavedAlias | null>(null)
  const shown = aliases.find((alias) => alias.id === aliasId)?.name ?? DEFAULT_LABEL

  function choose(next: string | null) {
    selectVersion(tabId, next)
    pin(tabId)
  }

  return (
    <>
      <Dropdown.Root className={styles.root}>
        <Dropdown.Trigger label={`Aliases: ${shown}`} className={styles.trigger}>
          <span className={styles.caption}>Alias</span>
          <span className={styles.shown}>{shown}</span>
        </Dropdown.Trigger>
        <Dropdown.Menu label="Aliases" className={styles.menu}>
          <Dropdown.RadioItem checked={aliasId === null} onSelect={() => choose(null)}>
            {DEFAULT_LABEL}
          </Dropdown.RadioItem>
          {aliases.map((alias) => (
            <div key={alias.id} className={styles.row}>
              <Dropdown.RadioItem checked={alias.id === aliasId} onSelect={() => choose(alias.id)} className={styles.name}>
                {alias.name}
              </Dropdown.RadioItem>
              <Dropdown.Item onSelect={() => setDeleting(alias)} className={styles.more}>
                <span aria-hidden="true">⋯</span>
                <span className="visually-hidden">Delete alias {alias.name}…</span>
              </Dropdown.Item>
            </div>
          ))}
          <div role="separator" className={styles.separator} />
          <Dropdown.Item onSelect={() => setCreating(true)}>Create new alias…</Dropdown.Item>
        </Dropdown.Menu>
      </Dropdown.Root>
      <CreateAliasDialog tabId={tabId} open={creating} onClose={() => setCreating(false)} />
      <DeleteAliasDialog tabId={tabId} alias={deleting} requestName={requestName} onClose={() => setDeleting(null)} />
    </>
  )
}
