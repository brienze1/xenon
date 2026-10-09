import { useState } from 'react'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { FolderQuestionArt } from '@/shared/ui/empty-state/illustrations/FolderQuestionArt'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { RemoveWorkspaceDialog } from './RemoveWorkspaceDialog'
import { WorkspaceErrorText } from './WorkspaceErrorText'

const MISSING_FOLDER_INTRO = (
  <>
    <EmptyState.Art>
      <FolderQuestionArt />
    </EmptyState.Art>
    <EmptyState.Title>Folder Missing</EmptyState.Title>
  </>
)

interface MissingFolderNoticeProps {
  localId: string
}

export function MissingFolderNotice({ localId }: MissingFolderNoticeProps) {
  const workspace = useWorkspacesStore((state) => state.workspaces.find((candidate) => candidate.localId === localId))
  const error = useWorkspacesStore((state) => state.error)
  const locate = useWorkspacesStore((state) => state.locate)
  const [removing, setRemoving] = useState(false)

  return (
    <EmptyState.Root label="Workspace folder missing">
      {MISSING_FOLDER_INTRO}
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
        <Button variant="secondary" disabled={!workspace} onClick={() => setRemoving(true)}>
          Remove from <span translate="no">Xenon</span>…
        </Button>
      </EmptyState.Actions>
      {error && error.code !== 'WKS-06404' ? (
        <Banner.Root tone="error">
          <Banner.Message><WorkspaceErrorText error={error} /></Banner.Message>
        </Banner.Root>
      ) : null}
      {workspace ? <RemoveWorkspaceDialog key={`remove-${workspace.localId}`} workspace={workspace} open={removing} onClose={() => setRemoving(false)} /> : null}
    </EmptyState.Root>
  )
}
