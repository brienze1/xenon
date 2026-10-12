import { useId, useState } from 'react'
import type { AppError } from '@/platform/errors'
import type { DocumentationItem, DocumentationPlan } from '@/platform/mappers/documentation'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { planSummary } from '@/features/documentation/lib/plan-summary'
import styles from './SavePlanDialog.module.css'

const SAVE_NEXT_STEP = 'Check that the workspace folder exists and is writable, then try again.'

interface PlanListProps {
  title: string
  items: DocumentationItem[]
  withFolder?: boolean
}

function PlanList({ title, items, withFolder = false }: PlanListProps) {
  if (items.length === 0) {
    return null
  }
  return (
    <section aria-label={title} className={styles.section}>
      <h3 className={styles.heading}>{title}</h3>
      <ul className={styles.list}>
        {items.map((item, index) => (
          <li key={item.id || `${item.name}-${index}`} className={styles.item}>
            <MethodBadge method={item.method} />
            <span translate="no">{item.name}</span>
            {withFolder ? <span className={styles.folder}>{item.folder ? `in ${item.folder}` : 'in the collection'}</span> : null}
          </li>
        ))}
      </ul>
    </section>
  )
}

interface PlanContentProps {
  plan: DocumentationPlan
  saving: boolean
  failure: AppError | null
  onCancel: () => void
  onConfirm: (confirmDeletions: boolean) => void
}

function PlanContent({ plan, saving, failure, onCancel, onConfirm }: PlanContentProps) {
  const [deleteConfirmed, setDeleteConfirmed] = useState(false)
  const confirmId = useId()
  const deletions = plan.deleted.length
  const blocked = saving || (deletions > 0 && !deleteConfirmed)
  return (
    <>
      <Dialog.Title>Save Documentation Changes?</Dialog.Title>
      <Dialog.Body>
        <p className={styles.summary}>{planSummary(plan)}.</p>
        {plan.collectionChanged ? <p>The collection’s name or description changes too.</p> : null}
        <PlanList title="Updated" items={plan.updated} />
        <PlanList title="Created" items={plan.created} withFolder />
        <PlanList title="Deleted" items={plan.deleted} />
        {deletions > 0 ? (
          <label htmlFor={confirmId} className={styles.confirm}>
            <input id={confirmId} type="checkbox" checked={deleteConfirmed} onChange={(event) => setDeleteConfirmed(event.target.checked)} />
            Delete {deletions === 1 ? 'this request' : `these ${deletions} requests`} from disk
          </label>
        ) : null}
        {failure ? (
          <Banner.Root tone="error">
            <Banner.Message>
              <NoTranslateText text={withNextStep(failure.description || failure.code, SAVE_NEXT_STEP)} terms={[failure.code]} />
            </Banner.Message>
          </Banner.Root>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
        <Button variant={deletions > 0 ? 'danger' : 'primary'} disabled={blocked} onClick={() => onConfirm(deletions > 0 && deleteConfirmed)}>
          {saving ? 'Saving…' : 'Save Changes'}
        </Button>
      </Dialog.Actions>
    </>
  )
}

interface SavePlanDialogProps {
  plan: DocumentationPlan | null
  saving: boolean
  failure: AppError | null
  onCancel: () => void
  onConfirm: (confirmDeletions: boolean) => void
}

/** Lists what saving the documentation changes and asks before writing it; deletions need an explicit tick. */
export function SavePlanDialog({ plan, saving, failure, onCancel, onConfirm }: SavePlanDialogProps) {
  return (
    <Dialog.Root open={plan !== null} onOpenChange={(open) => (open ? undefined : onCancel())}>
      {plan ? <PlanContent plan={plan} saving={saving} failure={failure} onCancel={onCancel} onConfirm={onConfirm} /> : null}
    </Dialog.Root>
  )
}
