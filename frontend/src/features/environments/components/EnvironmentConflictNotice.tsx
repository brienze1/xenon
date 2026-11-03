import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { useEnvironmentsStore } from '@/features/environments/store'

const CHANGED_MESSAGE = <Banner.Message>This environment changed on disk since it was opened.</Banner.Message>

interface EnvironmentConflictNoticeProps {
  scope: string
}

/** Shown when saving the scope's variables was refused because the environment file changed on disk meanwhile. */
export function EnvironmentConflictNotice({ scope }: EnvironmentConflictNoticeProps) {
  const conflict = useEnvironmentsStore((state) => state.conflicts[scope] === true)
  const keepMine = useEnvironmentsStore((state) => state.keepMine)
  const loadFromDisk = useEnvironmentsStore((state) => state.loadFromDisk)

  if (!conflict) {
    return null
  }
  return (
    <Banner.Root tone="warning">
      {CHANGED_MESSAGE}
      <Banner.Actions>
        <Button variant="secondary" size="sm" onClick={() => void keepMine(scope)}>
          Keep mine
        </Button>
        <Button variant="ghost" size="sm" onClick={() => void loadFromDisk(scope)}>
          Load from disk
        </Button>
      </Banner.Actions>
    </Banner.Root>
  )
}
