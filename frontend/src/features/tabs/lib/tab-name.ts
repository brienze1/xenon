import type { TreeNode } from '@/platform/mappers/tree'
import type { TabMeta } from '@/features/tabs/model'
import type { NameSubject } from './name-draft'

/** The name a tab edits: its draft and the saved name the sidebar shows (null when nothing is saved yet). */
export interface TabName {
  draft: string | undefined
  saved: string | null
  subject: NameSubject
}

/** What the stores hold about one tab's name; each store fills in its own part. */
export interface TabNameParts {
  /** The request builder's draft name (request and example tabs). */
  builderName: string | undefined
  /** The name the builder loaded or last saved; null when it does not know it (a draft restored from the UI state). */
  builderSavedName: string | null
  /** The tree's name of the tab's request, collection or folder, or of its example: the fallback saved name. */
  treeName: string | null
  itemDraft: string | undefined
  itemBase: string | null
  /** The saved example's name, as the examples store loaded or last saved it. */
  exampleName: string | null
  environmentDraft: string | undefined
  environmentName: string | null
}

export function treeNodeName(nodes: readonly TreeNode[], id: string): string | null {
  for (const node of nodes) {
    if (node.id === id) {
      return node.name
    }
    const found = treeNodeName(node.children, id)
    if (found !== null) {
      return found
    }
  }
  return null
}

export function treeExampleName(nodes: readonly TreeNode[], exampleId: string): string | null {
  for (const node of nodes) {
    const example = node.examples?.find((candidate) => candidate.id === exampleId)
    if (example) {
      return example.name
    }
    const found = treeExampleName(node.children, exampleId)
    if (found !== null) {
      return found
    }
  }
  return null
}

/** The tree id whose name the tab shows: its example, collection, folder or request. */
export function namedTreeId(tab: TabMeta): string | null {
  if (tab.kind === 'example') {
    return tab.exampleId ?? null
  }
  if (tab.kind === 'collection' || tab.kind === 'folder') {
    return tab.itemId ?? null
  }
  return tab.kind === undefined || tab.kind === 'request' ? tab.requestId : null
}

/**
 * The tab's name draft and saved name; null for tabs without a name of their own (Globals, Settings).
 * The saved name comes from the store that loaded the item, so it moves together with the draft (a rename made
 * elsewhere reloads both); the tree only fills in while that store does not know it.
 */
export function tabNameOf(tab: TabMeta, parts: TabNameParts): TabName | null {
  if (tab.kind === 'collection' || tab.kind === 'folder') {
    return { draft: parts.itemDraft, saved: parts.itemBase ?? parts.treeName, subject: tab.kind }
  }
  if (tab.kind === 'environment') {
    return { draft: parts.environmentDraft, saved: parts.environmentName, subject: 'environment' }
  }
  if (tab.kind === 'example') {
    return { draft: parts.builderName, saved: parts.exampleName ?? parts.treeName, subject: 'example' }
  }
  if (tab.kind === 'globals' || tab.kind === 'settings') {
    return null
  }
  return { draft: parts.builderName, saved: tab.requestId === null ? null : (parts.builderSavedName ?? parts.treeName), subject: 'request' }
}
