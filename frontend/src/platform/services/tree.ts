import * as TreeService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/treeservice'
import { AppError, toAppError } from '../errors'
import { toRequestDto, type RequestDraft } from '../mappers/request'
import {
  toItemSettings,
  toSaveItemSettingsDto,
  toTreeNode,
  toTreeSearch,
  type ItemSettings,
  type TreeNode,
  type TreeSearchResult,
} from '../mappers/tree'

const EMPTY_RESPONSE = 'UNK-00204'

function required<T>(value: T | null): T {
  if (value === null) {
    throw new AppError(EMPTY_RESPONSE, 'The backend returned no data')
  }
  return value
}

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

export function getTree(localId: string): Promise<TreeNode[]> {
  return call(async () => ((await TreeService.GetTree({ local_id: localId })) ?? []).map(toTreeNode))
}

export function searchTree(localId: string, query: string): Promise<TreeSearchResult> {
  return call(async () => toTreeSearch(required(await TreeService.Search({ local_id: localId, query }))))
}

export function createCollection(localId: string, name: string): Promise<TreeNode> {
  return call(async () => toTreeNode(required(await TreeService.CreateCollection({ local_id: localId, name }))))
}

export function createFolder(localId: string, parentId: string, name: string): Promise<TreeNode> {
  return call(async () => toTreeNode(required(await TreeService.CreateFolder({ local_id: localId, parent_id: parentId, name }))))
}

export function createRequest(localId: string, parentId: string, name: string, draft?: RequestDraft): Promise<TreeNode> {
  return call(async () =>
    toTreeNode(
      required(
        await TreeService.CreateRequest({
          local_id: localId,
          parent_id: parentId,
          name,
          request: draft ? toRequestDto(draft) : null,
        }),
      ),
    ),
  )
}

export function renameItem(localId: string, id: string, name: string): Promise<TreeNode> {
  return call(async () => toTreeNode(required(await TreeService.Rename({ local_id: localId, id, name }))))
}

export function duplicateItem(localId: string, id: string): Promise<TreeNode> {
  return call(async () => toTreeNode(required(await TreeService.Duplicate({ local_id: localId, id }))))
}

export function deleteItem(localId: string, id: string): Promise<string[]> {
  return call(async () => required(await TreeService.Delete({ local_id: localId, id })).deleted_request_ids ?? [])
}

/** Moves the item to `index` among the new parent's other children ("" for collections) and returns the whole tree. */
export function moveItem(localId: string, id: string, newParentId: string, index: number): Promise<TreeNode[]> {
  return call(async () => ((await TreeService.Move({ local_id: localId, id, new_parent_id: newParentId, index })) ?? []).map(toTreeNode))
}

export function getItemSettings(localId: string, id: string): Promise<ItemSettings> {
  return call(async () => toItemSettings(required(await TreeService.GetSettings({ local_id: localId, id }))))
}

export function saveItemSettings(localId: string, settings: ItemSettings): Promise<ItemSettings> {
  return call(async () => toItemSettings(required(await TreeService.SaveSettings(toSaveItemSettingsDto(localId, settings)))))
}
