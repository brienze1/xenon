import { DocumentationPanel } from '@/features/documentation/components/DocumentationPanel'
import { VariableTable } from '@/features/environments/components/VariableTable'
import { ItemTab } from '@/features/item-settings/components/ItemTab'
import type { ItemAuthSlot, ItemScriptsSlot } from '@/features/item-settings/model'
import { RunnerSection } from '@/features/runner/components/RunnerSection'
import type { ContainerKind } from '@/platform/mappers/tree'
import { collectionScope } from '@/platform/mappers/variable'
import { ItemAuth } from './ItemAuth'
import { ItemScripts } from './ItemScripts'
import { RUNNER_SLOTS } from './runner-slots'

function authEditor(slot: ItemAuthSlot) {
  return <ItemAuth slot={slot} />
}

function scriptsEditor(slot: ItemScriptsSlot) {
  return <ItemScripts slot={slot} />
}

function collectionVariables(collectionId: string) {
  return <VariableTable scope={collectionScope(collectionId)} label="Collection variables" />
}

function documentation(itemId: string, kind: ContainerKind) {
  return <DocumentationPanel itemId={itemId} kind={kind} />
}

function runner(itemId: string) {
  return <RunnerSection itemId={itemId} slots={RUNNER_SLOTS} />
}

interface ItemWorkspaceProps {
  itemId: string
  kind: ContainerKind
  onSave: () => void
}

/** A collection or folder tab with the editors it shares with requests (auth, scripts, variables), its documentation and its Runner. */
export function ItemWorkspace({ itemId, kind, onSave }: ItemWorkspaceProps) {
  return (
    <ItemTab
      itemId={itemId}
      kind={kind}
      onSave={onSave}
      authEditor={authEditor}
      scriptsEditor={scriptsEditor}
      collectionVariables={collectionVariables}
      documentation={documentation}
      runner={runner}
    />
  )
}
