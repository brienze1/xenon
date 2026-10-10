import { useShallow } from 'zustand/react/shallow'
import { cn } from '@/shared/lib/cn'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useEnvironmentsStore } from '@/features/environments/store'
import { EnvironmentsManager } from './EnvironmentsManager'
import styles from './EnvironmentSwitcher.module.css'
import { GlobalsDialog } from './GlobalsDialog'
import { QuickLook } from './QuickLook'

export const NO_ENVIRONMENT_LABEL = 'No environment'

const CHEVRON = (
  <span className={styles.chevron} aria-hidden="true">
    ▾
  </span>
)

const SWITCH_NEXT_STEP = 'Check that the workspace folder is writable, then pick the environment again.'

const BROKEN_NEXT_STEP = 'Fix or remove the file in the workspace folder.'

const MENU_SEPARATOR = <hr className={styles.separator} />

const EDITORS = (
  <>
    <EnvironmentsManager />
    <GlobalsDialog />
  </>
)

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
  const openEditor = useEnvironmentsStore((state) => state.openEditor)
  const active = environments.find((environment) => environment.id === activeId)
  const activeName = active?.name ?? NO_ENVIRONMENT_LABEL
  const activeColor = active?.color ?? 'none'

  return (
    <div className={styles.container}>
      <Dropdown.Root className={styles.switcher} open={switcherOpen} onOpenChange={setSwitcherOpen}>
        <Dropdown.Trigger label={`Environment: ${activeName}`} className={cn(styles.trigger, styles[activeColor])}>
          <span className={styles.swatch} data-color={activeColor} aria-hidden="true" />
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
                <span className={styles.option}>
                  <span className={styles.swatch} data-color={environment.color} aria-hidden="true" />
                  <span className={styles.optionName} translate="no" title={environment.name}>
                    {environment.name}
                  </span>
                </span>
              </Dropdown.RadioItem>
            ) : (
              <Dropdown.Item key={environment.id} disabled onSelect={() => undefined}>
                <span className={styles.broken} title={withNextStep(environment.error, BROKEN_NEXT_STEP)}>
                  <span translate="no">{environment.name}</span>
                  <span className={styles.brokenMark}>
                    <NoTranslateText text={withNextStep(environment.error, BROKEN_NEXT_STEP)} terms={[environment.error]} />
                  </span>
                </span>
              </Dropdown.Item>
            ),
          )}
          {MENU_SEPARATOR}
          <Dropdown.Item onSelect={() => openEditor('globals')}>Globals</Dropdown.Item>
          <Dropdown.Item onSelect={() => openEditor('manager')}>Manage Environments…</Dropdown.Item>
        </Dropdown.Menu>
      </Dropdown.Root>
      <QuickLook />
      {error ? (
        <Banner.Root tone="error" className={styles.error}>
          <Banner.Message>
            <NoTranslateText
              text={withNextStep(`Could not change the environment: ${error.description || error.code}`, SWITCH_NEXT_STEP)}
              terms={[error.code]}
            />
          </Banner.Message>
          <Banner.Actions>
            <Button variant="ghost" size="sm" onClick={clearError}>
              Dismiss Error
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      {EDITORS}
    </div>
  )
}
