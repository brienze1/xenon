import { useEffect } from 'react'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import type { RunnerSlots, RunnerView } from '@/features/runner/model'
import { useRunnerStore } from '@/features/runner/store'
import { ConfirmRunDialog } from './ConfirmRunDialog'
import { RunMissingSecretsDialog } from './RunMissingSecretsDialog'
import { RunOrderList } from './RunOrderList'
import { RunProgress } from './RunProgress'
import { RunResults } from './RunResults'
import { RunSettings } from './RunSettings'
import { RunsList } from './RunsList'
import styles from './RunnerSection.module.css'

const LOAD_NEXT_STEP = 'Check that the workspace folder exists, then try again.'

const LOADING = (
  <p role="status" className={styles.note}>
    Loading…
  </p>
)

const LOADING_RUN = (
  <p role="status" className={styles.note}>
    Loading results…
  </p>
)

const NO_RUN = (
  <EmptyState.Root label="No run yet" size="compact">
    <EmptyState.Title>No Run Yet</EmptyState.Title>
    <EmptyState.Description>Start a run from Setup, or open a stored run from Runs.</EmptyState.Description>
  </EmptyState.Root>
)

const SETUP_TAB = <Tabs.Tab value="setup">Setup</Tabs.Tab>
const RESULTS_TAB = <Tabs.Tab value="results">Results</Tabs.Tab>
const RUNS_TAB = <Tabs.Tab value="runs">Runs</Tabs.Tab>

interface RunnerSectionProps {
  itemId: string
  /** The response viewer parts a result row shows; composed by the app. */
  slots: RunnerSlots
}

/** The Runner section of a collection or folder tab: Setup, Results and Runs (FR-001). */
export function RunnerSection({ itemId, slots }: RunnerSectionProps) {
  const ensure = useRunnerStore((state) => state.ensure)
  const load = useRunnerStore((state) => state.load)
  const status = useRunnerStore((state) => state.targets[itemId]?.status ?? 'idle')
  const loadError = useRunnerStore((state) => state.targets[itemId]?.loadError ?? null)
  const hasConfig = useRunnerStore((state) => state.targets[itemId]?.config != null)
  const view = useRunnerStore((state) => state.targets[itemId]?.view ?? 'setup')
  const failure = useRunnerStore((state) => state.targets[itemId]?.failure ?? null)
  const showView = useRunnerStore((state) => state.showView)
  const dismissFailure = useRunnerStore((state) => state.dismissFailure)

  useEffect(() => {
    void ensure(itemId)
  }, [itemId, ensure])

  if (!hasConfig) {
    if (status === 'ready' && loadError) {
      return (
        <Banner.Root tone="error">
          <Banner.Message>
            <NoTranslateText text={withNextStep(loadError.description || loadError.code, LOAD_NEXT_STEP)} terms={[loadError.code]} />
          </Banner.Message>
          <Banner.Actions>
            <Button variant="secondary" size="sm" onClick={() => void load(itemId)}>
              Try Again
            </Button>
          </Banner.Actions>
        </Banner.Root>
      )
    }
    return LOADING
  }

  return (
    <div className={styles.runner}>
      {failure ? (
        <Banner.Root tone="error">
          <Banner.Message>
            <NoTranslateText text={withNextStep(failure.error.description || failure.error.code, failure.nextStep)} terms={[failure.error.code]} />
          </Banner.Message>
          <Banner.Actions>
            <Button variant="ghost" size="sm" onClick={() => dismissFailure(itemId)}>
              Dismiss
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      <Tabs.Root value={view} onValueChange={(next) => showView(itemId, next as RunnerView)} className={styles.views}>
        <Tabs.List label="Runner">
          {SETUP_TAB}
          {RESULTS_TAB}
          {RUNS_TAB}
        </Tabs.List>
        <Tabs.Panel value="setup" className={styles.setup}>
          <RunOrderList itemId={itemId} />
          <RunSettings itemId={itemId} />
        </Tabs.Panel>
        <Tabs.Panel value="results" className={styles.panel}>
          <ResultsView itemId={itemId} slots={slots} />
        </Tabs.Panel>
        <Tabs.Panel value="runs" className={styles.panel}>
          <RunsList itemId={itemId} />
        </Tabs.Panel>
      </Tabs.Root>
      <ConfirmRunDialog itemId={itemId} />
      <RunMissingSecretsDialog itemId={itemId} />
    </div>
  )
}

interface ResultsViewProps {
  itemId: string
  slots: RunnerSlots
}

function ResultsView({ itemId, slots }: ResultsViewProps) {
  const runId = useRunnerStore((state) => state.targets[itemId]?.runId ?? null)
  const run = useRunnerStore((state) => state.targets[itemId]?.run ?? null)
  const active = useRunnerStore((state) => {
    const current = state.active[itemId]
    return current !== undefined && current.id === state.targets[itemId]?.runId ? current : null
  })

  if (runId === null) {
    return NO_RUN
  }
  return (
    <>
      {active ? <RunProgress itemId={itemId} run={active} /> : null}
      {run?.id === runId ? <RunResults itemId={itemId} run={run} slots={slots} /> : active ? null : LOADING_RUN}
    </>
  )
}
