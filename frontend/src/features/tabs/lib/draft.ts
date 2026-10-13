import type { RequestDraft } from '@/platform/mappers/request'

type Comparable = unknown

function isRecord(value: Comparable): value is Record<string, Comparable> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function keysOf(value: Record<string, Comparable>, ignoreId: boolean): string[] {
  return Object.keys(value).filter((key) => !(ignoreId && key === 'id'))
}

function sameValue(left: Comparable, right: Comparable, isRow: boolean): boolean {
  if (Object.is(left, right)) {
    return true
  }
  if (Array.isArray(left) || Array.isArray(right)) {
    return (
      Array.isArray(left) &&
      Array.isArray(right) &&
      left.length === right.length &&
      left.every((item, index) => sameValue(item, right[index], true))
    )
  }
  if (!isRecord(left) || !isRecord(right)) {
    return false
  }
  const leftKeys = keysOf(left, isRow)
  const rightKeys = keysOf(right, isRow)
  return leftKeys.length === rightKeys.length && leftKeys.every((key) => Object.hasOwn(right, key) && sameValue(left[key], right[key], false))
}

// Compares two drafts by content. Table rows (params, headers, form fields…) carry a
// UI-only id that never reaches disk, so rows match on their fields alone.
export function sameDraftContent(left: RequestDraft, right: RequestDraft): boolean {
  return sameValue(left, right, false)
}
