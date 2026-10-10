import { useShallow } from 'zustand/react/shallow'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { useEnvironmentsStore } from '@/features/environments/store'
import styles from './EnvironmentSwitcher.module.css'

export const NO_ENVIRONMENT_LABEL = 'No environment'

const CHEVRON = (
  <span className={styles.chevron} aria-hidden="true">
    ▾
  </span>
)

const MENU_SEPARATOR = <hr className={styles.separator} />

export function EnvironmentSwitcher() {
  const { environments, activeId, switcherOpen, setActive, setSwitcherOpen } = useEnvironmentsStore(
    useShallow((state) => ({
      environments: state.environments,
      activeId: state.activeId,
      switcherOpen: state.switcherOpen,
      setActive: state.setActive,
      setSwitcherOpen: state.setSwitcherOpen,
    })),
  )
  const error = useEnvironmentsStore((state) => state.error)
  const clearError = useEnvironmentsStore((state) => state.clearError)
  const activeName = environments.find((environment) => environment.id === activeId)?.name ?? NO_ENVIRONMENT_LABEL

  return (
    <div className={styles.container}>
      <Dropdown.Root className={styles.switcher} open={switcherOpen} onOpenChange={setSwitcherOpen}>
        <Dropdown.Trigger label={`Environment: ${activeName}`}>
          <span className={styles.name} translate="no">
            {activeName}
          </span>
          {CHEVRON}
        </Dropdown.Trigger>
        <Dropdown.Menu label="Environments" className={styles.menu}>
          <Dropdown.RadioItem checked={activeId === null} onSelect={() => void setActive(null)}>
            {NO_ENVIRONMENT_LABEL}
          </Dropdown.RadioItem>
          {environments.length > 0 ? MENU_SEPARATOR : null}
          {environments.map((environment) =>
            environment.error === null ? (
              <Dropdown.RadioItem
                key={environment.id}
                checked={environment.id === activeId}
                onSelect={() => void setActive(environment.id)}
              >
                <span translate="no">{environment.name}</span>
              </Dropdown.RadioItem>
            ) : (
              <Dropdown.Item key={environment.id} disabled onSelect={() => undefined}>
                <span className={styles.broken} title={environment.error}>
                  <span translate="no">{environment.name}</span>
                  <span className={styles.brokenMark}>{environment.error}</span>
                </span>
              </Dropdown.Item>
            ),
          )}
        </Dropdown.Menu>
      </Dropdown.Root>
      {error ? (
        <Banner.Root tone="error" className={styles.error}>
          <Banner.Message>Could not change the environment: {error.description || error.code}</Banner.Message>
          <Banner.Actions>
            <Button variant="ghost" size="sm" onClick={clearError}>
              Dismiss Error
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
    </div>
  )
}
