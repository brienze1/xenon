import { useState } from 'react'
import { useSyncStore } from '@/features/sync/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'

interface UnlinkRepositoryProps {
  clonePath: string
  disabled: boolean
}

export function UnlinkRepository({ clonePath, disabled }: UnlinkRepositoryProps) {
  const unlink = useSyncStore((state) => state.unlink)
  const chooseFolder = useWorkspacesStore((state) => state.chooseFolder)
  const [target, setTarget] = useState<string | null>(null)

  async function chooseTarget() {
    const folder = await chooseFolder('Choose an empty folder for the workspace files')
    if (folder) {
      setTarget(folder)
    }
  }

  async function confirm(folder: string) {
    setTarget(null)
    await unlink(folder)
  }

  return target === null ? (
    <Button size="sm" variant="ghost" disabled={disabled} onClick={() => void chooseTarget()}>
      Unlink Repository…
    </Button>
  ) : (
    <Banner.Root tone="warning">
      <Banner.Message>
        Copy the workspace files to <span translate="no">{target}</span>, keep working there without Git and delete Xenon&apos;s clone at{' '}
        <span translate="no">{clonePath}</span>? The repository is not changed.
      </Banner.Message>
      <Banner.Actions>
        <Button size="sm" variant="secondary" onClick={() => setTarget(null)}>
          Cancel
        </Button>
        <Button size="sm" variant="danger" disabled={disabled} onClick={() => void confirm(target)}>
          Unlink and Delete Clone
        </Button>
      </Banner.Actions>
    </Banner.Root>
  )
}
