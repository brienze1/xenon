import { useShallow } from 'zustand/react/shallow'
import { cn } from '@/shared/lib/cn'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useEnvironmentsStore } from '@/features/environments/store'
import styles from './EnvironmentSwitcher.module.css'

export const NO_ENVIRONMENT_LABEL = 'No environment'

const CHEVRON = (
  <span className={styles.chevron} aria-hidden="true">
    ▾
  </span>
)

const SWITCH_NEXT_STEP = 'Check that the workspace folder is writable, then pick the environment again.'

const BROKEN_NEXT_STEP = 'Fix or remove the file in the workspace folder.'

const MENU_SEPARATOR = <hr className={styles.separator} />

const EYE_GLYPH = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.icon}>
    <path d="M2.06 12.35a1 1 0 0 1 0-.7 10.75 10.75 0 0 1 19.88 0 1 1 0 0 1 0 .7 10.75 10.75 0 0 1-19.88 0" />
    <circle cx="12" cy="12" r="3" />
  </svg>
)

const OPEN_ENVIRONMENT_LABEL = 'Open environment'

const OPEN_GLOBALS_LABEL = 'Open globals'

interface EnvironmentSwitcherProps {
  /** Opens the globals in a main-area tab. */
  onOpenGlobals: () => void
  /** Opens one environment in a main-area tab. */
  onOpenEnvironment: (environmentId: string) => void
  /** Shows the environments in the sidebar, where they are created, renamed and opened. */
  onManage: () => void
}

export function EnvironmentSwitcher({ onOpenGlobals, onOpenEnvironment, onManage }: EnvironmentSwitcherProps) {
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
          <Dropdown.Item onSelect={onOpenGlobals}>Globals</Dropdown.Item>
          <Dropdown.Item onSelect={onManage}>Manage Environments…</Dropdown.Item>
        </Dropdown.Menu>
      </Dropdown.Root>
      <IconButton
        label={active ? OPEN_ENVIRONMENT_LABEL : OPEN_GLOBALS_LABEL}
        onClick={() => (active ? onOpenEnvironment(active.id) : onOpenGlobals())}
      >
        {EYE_GLYPH}
      </IconButton>
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
    </div>
  )
}
