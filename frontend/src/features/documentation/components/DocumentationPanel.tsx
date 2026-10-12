import { useEffect, useMemo } from 'react'
import type { AppError } from '@/platform/errors'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { CodeEditor } from '@/shared/ui/code-editor/CodeEditor'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { problemMessage, problemPosition } from '@/features/documentation/lib/problem-position'
import type { DocumentationView } from '@/features/documentation/model'
import { useDocumentationStore } from '@/features/documentation/store'
import { ApiDocView } from './ApiDocView'
import { SavePlanDialog } from './SavePlanDialog'
import styles from './DocumentationPanel.module.css'

const LOAD_NEXT_STEP = 'Check that the workspace folder exists, then try again.'
const NO_DIAGNOSTICS: never[] = []

const LOADING = (
  <p role="status" className={styles.note}>
    Loading…
  </p>
)

const STALE_MESSAGE = <Banner.Message>The collection changed on disk while you were editing its documentation.</Banner.Message>

const HINT = (
  <p className={styles.note}>
    Built from the requests of this collection. Edit the YAML and save to change the requests too — names, descriptions, URLs,
    parameters, headers, bodies and examples; operations without an x-xenon-id become new requests.
  </p>
)

interface ErrorTextProps {
  error: AppError
  nextStep: string
}

function ErrorText({ error, nextStep }: ErrorTextProps) {
  return <NoTranslateText text={withNextStep(error.description || error.code, nextStep)} terms={[error.code]} />
}

interface ViewToggleProps {
  view: DocumentationView
  onChange: (view: DocumentationView) => void
}

function ViewToggle({ view, onChange }: ViewToggleProps) {
  return (
    <div role="group" aria-label="Documentation view" className={styles.toggle}>
      <Button variant={view === 'preview' ? 'primary' : 'ghost'} size="sm" aria-pressed={view === 'preview'} onClick={() => onChange('preview')}>
        Preview
      </Button>
      <Button variant={view === 'yaml' ? 'primary' : 'ghost'} size="sm" aria-pressed={view === 'yaml'} onClick={() => onChange('yaml')}>
        YAML
      </Button>
    </div>
  )
}

interface DocumentationPanelProps {
  collectionId: string
}

/** A collection's OpenAPI documentation: rendered, or as YAML that saves back onto its requests. */
export function DocumentationPanel({ collectionId }: DocumentationPanelProps) {
  const open = useDocumentationStore((state) => state.open)
  const load = useDocumentationStore((state) => state.load)
  const entry = useDocumentationStore((state) => state.entries[collectionId])
  const edit = useDocumentationStore((state) => state.edit)
  const setView = useDocumentationStore((state) => state.setView)
  const discard = useDocumentationStore((state) => state.discard)
  const keepDraft = useDocumentationStore((state) => state.keepDraft)
  const cancelPlan = useDocumentationStore((state) => state.cancelPlan)
  const confirm = useDocumentationStore((state) => state.confirm)
  const problem = entry?.problem ?? null
  const position = problemPosition(problem)
  const diagnostics = useMemo(() => {
    const at = problemPosition(problem)
    return problem && at ? [{ line: at.line, column: at.column, message: problemMessage(problem) }] : NO_DIAGNOSTICS
  }, [problem])

  useEffect(() => {
    if (useDocumentationStore.getState().entries[collectionId]) {
      void load(collectionId)
    } else {
      void open(collectionId)
    }
  }, [collectionId, open, load])

  if (!entry?.base) {
    if (entry?.status === 'ready' && entry.loadError) {
      return (
        <Banner.Root tone="error">
          <Banner.Message>
            <ErrorText error={entry.loadError} nextStep={LOAD_NEXT_STEP} />
          </Banner.Message>
          <Banner.Actions>
            <Button variant="secondary" size="sm" onClick={() => void load(collectionId)}>
              Try Again
            </Button>
          </Banner.Actions>
        </Banner.Root>
      )
    }
    return LOADING
  }

  const document = entry.dirty ? (entry.preview ?? entry.base.document) : entry.base.document

  return (
    <div className={styles.panel}>
      <div className={styles.toolbar}>
        <ViewToggle view={entry.view} onChange={(view) => setView(collectionId, view)} />
        {entry.dirty ? (
          <Button variant="ghost" size="sm" onClick={() => discard(collectionId)}>
            Discard Changes
          </Button>
        ) : null}
      </div>
      {HINT}
      {entry.stale ? (
        <Banner.Root tone="warning">
          {STALE_MESSAGE}
          <Banner.Actions>
            <Button variant="secondary" size="sm" onClick={() => keepDraft(collectionId)}>
              Keep Your Version
            </Button>
            <Button variant="ghost" size="sm" onClick={() => discard(collectionId)}>
              Load from Collection
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      {entry.problem && position ? (
        <Banner.Root tone="error">
          <Banner.Message>
            The YAML cannot be read at line {position.line}, column {position.column}: {problemMessage(entry.problem)}
          </Banner.Message>
          {entry.view === 'preview' ? (
            <Banner.Actions>
              <Button variant="secondary" size="sm" onClick={() => setView(collectionId, 'yaml')}>
                Show YAML
              </Button>
            </Banner.Actions>
          ) : null}
        </Banner.Root>
      ) : null}
      {entry.failure && !entry.plan && !(entry.problem && position) ? (
        <Banner.Root tone="error">
          <Banner.Message>
            <ErrorText error={entry.failure} nextStep="Fix the YAML or try again." />
          </Banner.Message>
        </Banner.Root>
      ) : null}
      {entry.view === 'yaml' ? (
        <CodeEditor
          value={entry.draft}
          language="yaml"
          label="OpenAPI document"
          onChange={(yaml) => edit(collectionId, yaml)}
          diagnostics={diagnostics}
          className={styles.editor}
        />
      ) : (
        <ApiDocView document={document} />
      )}
      <SavePlanDialog
        plan={entry.plan}
        saving={entry.saving}
        failure={entry.plan ? entry.failure : null}
        onCancel={() => cancelPlan(collectionId)}
        onConfirm={(confirmDeletions) => void confirm(collectionId, confirmDeletions)}
      />
    </div>
  )
}
