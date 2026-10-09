import type {
  ItemSettings as ItemSettingsDto,
  TreeNode as TreeNodeDto,
  TreeSearch as TreeSearchDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import { toAuth, toAuthDto, toKeyValueDto, toKeyValueRow, type KeyValueRow, type RequestAuth } from './request'

export type TreeNodeKind = 'collection' | 'folder' | 'request' | 'error'
export type ContainerKind = 'collection' | 'folder'

export interface TreeNode {
  id: string
  kind: TreeNodeKind
  name: string
  method: string | null
  url: string | null
  error: string | null
  warning: string | null
  children: TreeNode[]
}

export interface TreeSearchResult {
  matchIds: string[]
  ancestorIds: string[]
}

export interface ItemSettings {
  id: string
  kind: ContainerKind
  name: string
  description: string
  auth: RequestAuth
  headers: KeyValueRow[]
}

const NODE_KINDS: readonly TreeNodeKind[] = ['collection', 'folder', 'request', 'error']

export function toTreeNode(dto: TreeNodeDto): TreeNode {
  return {
    id: dto.id,
    kind: NODE_KINDS.includes(dto.kind as TreeNodeKind) ? (dto.kind as TreeNodeKind) : 'error',
    name: dto.name,
    method: dto.method ?? null,
    url: dto.url ?? null,
    error: dto.error ? dto.error : null,
    warning: dto.warning ? dto.warning : null,
    children: (dto.children ?? []).map(toTreeNode),
  }
}

export function toTreeSearch(dto: TreeSearchDto): TreeSearchResult {
  return { matchIds: dto.match_ids ?? [], ancestorIds: dto.ancestor_ids ?? [] }
}

export function toItemSettings(dto: ItemSettingsDto): ItemSettings {
  return {
    id: dto.id,
    kind: dto.kind === 'folder' ? 'folder' : 'collection',
    name: dto.name,
    description: dto.description,
    auth: toAuth(dto.auth),
    headers: (dto.headers ?? []).map(toKeyValueRow),
  }
}

export function toSaveItemSettingsDto(localId: string, settings: ItemSettings) {
  return {
    local_id: localId,
    id: settings.id,
    description: settings.description,
    auth: toAuthDto(settings.auth),
    headers: settings.headers.map(toKeyValueDto),
  }
}
