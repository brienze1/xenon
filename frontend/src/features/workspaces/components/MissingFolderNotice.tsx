import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { WorkspaceErrorText } from './WorkspaceErrorText'

interface MissingFolderNoticeProps {
  localId: string
}

export function MissingFolderNotice({ localId }: MissingFolderNoticeProps) {
  const workspace = useWorkspacesStore((state) => state.workspaces.find((candidate) => candidate.localId === localId))
  const error = useWorkspacesStore((state) => state.error)
  const locate = useWorkspacesStore((state) => state.locate)
  const remove = useWorkspacesStore((state) => state.remove)

  return (
    <EmptyState.Root label="Workspace folder missing">
      <EmptyState.Art variant="missing" />
      <EmptyState.Title>Folder Missing</EmptyState.Title>
      <EmptyState.Description>
        Someone moved or deleted the folder of “{workspace?.name ?? 'this workspace'}” outside <span translate="no">Xenon</span>
        {workspace ? (
          <>
            {' ('}
            <span translate="no">{workspace.path}</span>)
          </>
        ) : null}
        . Locate it again, or remove the workspace from <span translate="no">Xenon</span>.
      </EmptyState.Description>
      <EmptyState.Actions>
        <Button variant="primary" onClick={() => void locate(localId)}>
          Locate Folder…
        </Button>
        <Button variant="secondary" onClick={() => void remove(localId, false)}>
          Remove from <span translate="no">Xenon</span>
        </Button>
      </EmptyState.Actions>
      {error && error.code !== 'WKS-06404' ? (
        <Banner.Root tone="error">
          <Banner.Message><WorkspaceErrorText error={error} /></Banner.Message>
        </Banner.Root>
      ) : null}
    </EmptyState.Root>
  )
}
