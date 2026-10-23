import type { RunConfig, RunPlanRow } from '@/platform/mappers/runner'
import type { TreeNode } from '@/platform/mappers/tree'

function renumber(rows: RunPlanRow[]): RunPlanRow[] {
  return rows.map((row, index) => (row.number === index + 1 ? row : { ...row, number: index + 1 }))
}

export function selectedCount(rows: readonly RunPlanRow[]): number {
  return rows.reduce((count, row) => (row.selected ? count + 1 : count), 0)
}

/** The select-all checkbox: checked, unchecked or mixed. */
export function selectionState(rows: readonly RunPlanRow[]): boolean | 'mixed' {
  const selected = selectedCount(rows)
  if (selected === 0) {
    return false
  }
  return selected === rows.length ? true : 'mixed'
}

export function toggleRow(rows: RunPlanRow[], requestId: string, selected: boolean): RunPlanRow[] {
  return rows.map((row) => (row.requestId === requestId ? { ...row, selected } : row))
}

export function setAllSelected(rows: RunPlanRow[], selected: boolean): RunPlanRow[] {
  return rows.map((row) => (row.selected === selected ? row : { ...row, selected }))
}

export function chooseAlias(rows: RunPlanRow[], requestId: string, aliasId: string): RunPlanRow[] {
  return rows.map((row) => (row.requestId === requestId ? { ...row, aliasId } : row))
}

/** Moves the row `fromId` to the position of `toId`, as a sortable list drop does. */
export function moveRow(rows: RunPlanRow[], fromId: string, toId: string): RunPlanRow[] {
  const from = rows.findIndex((row) => row.requestId === fromId)
  const to = rows.findIndex((row) => row.requestId === toId)
  if (from < 0 || to < 0 || from === to) {
    return rows
  }
  const next = [...rows]
  const [moved] = next.splice(from, 1)
  next.splice(to, 0, moved!)
  return renumber(next)
}

function findNode(nodes: readonly TreeNode[], id: string): TreeNode | null {
  for (const node of nodes) {
    if (node.id === id) {
      return node
    }
    const found = findNode(node.children, id)
    if (found) {
      return found
    }
  }
  return null
}

function requestIds(nodes: readonly TreeNode[], ids: string[]): string[] {
  for (const node of nodes) {
    if (node.kind === 'request') {
      ids.push(node.id)
    } else if (node.kind === 'collection' || node.kind === 'folder') {
      requestIds(node.children, ids)
    }
  }
  return ids
}

/**
 * Reset: every row checked, in sidebar order (depth first under the target). Rows the sidebar does
 * not know (yet) keep their relative order at the end; alias choices stay.
 */
export function resetRows(rows: RunPlanRow[], tree: readonly TreeNode[], itemId: string): RunPlanRow[] {
  const target = findNode(tree, itemId)
  const order = new Map(requestIds(target?.children ?? [], []).map((id, index) => [id, index]))
  const position = (row: RunPlanRow) => order.get(row.requestId) ?? order.size + row.number
  const sorted = [...rows].sort((left, right) => position(left) - position(right))
  return renumber(setAllSelected(sorted, true))
}

/** The configuration to remember or run: the settings with the rows as shown. */
export function configWithRows(config: RunConfig, rows: readonly RunPlanRow[]): RunConfig {
  return { ...config, rows: rows.map((row) => ({ requestId: row.requestId, aliasId: row.aliasId, selected: row.selected })) }
}
