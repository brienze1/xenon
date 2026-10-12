import type { DocumentationPlan } from '@/platform/mappers/documentation'

/** Whether saving would change anything at all. */
export function isEmptyPlan(plan: DocumentationPlan): boolean {
  return !plan.collectionChanged && plan.updated.length === 0 && plan.created.length === 0 && plan.deleted.length === 0
}

function count(amount: number, noun: string): string {
  return `${amount} ${noun}`
}

/** "2 updated, 1 created, 1 deleted". */
export function planSummary(plan: DocumentationPlan): string {
  return [count(plan.updated.length, 'updated'), count(plan.created.length, 'created'), count(plan.deleted.length, 'deleted')].join(', ')
}
