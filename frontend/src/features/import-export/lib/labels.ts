import type { ImportCounts } from '@/platform/mappers/import'
import { formatNumber } from '@/shared/lib/number-format'

const KIND_LABELS: Readonly<Record<string, string>> = {
  postman_collection: 'Postman collection',
  postman_environment: 'Postman environment',
  postman_globals: 'Postman globals',
  postman_dump: 'Postman data dump',
  postman_request: 'Postman request',
  postman_workspace: 'Postman workspace',
  bruno_bru: 'Bruno collection (.bru)',
  bruno_yaml: 'Bruno collection (YAML)',
  curl: 'cURL command',
  mixed: 'Multiple sources',
}

const COUNT_LABELS: ReadonlyArray<[keyof ImportCounts, string, string]> = [
  ['collections', 'collection', 'collections'],
  ['folders', 'folder', 'folders'],
  ['requests', 'request', 'requests'],
  ['examples', 'example', 'examples'],
  ['variables', 'variable', 'variables'],
  ['scripts', 'script', 'scripts'],
  ['environments', 'environment', 'environments'],
]

const ITEM_KIND_LABELS: Readonly<Record<string, string>> = {
  collection: 'Collection',
  environment: 'Environment',
  globals: 'Globals',
  request: 'Request',
}

export function itemKindLabel(kind: string): string {
  return ITEM_KIND_LABELS[kind] ?? kind
}

export function sourceLabel(kind: string, version: string): string {
  const label = KIND_LABELS[kind] ?? kind
  return version ? `${label} v${version}` : label
}

export function countsLabel(counts: ImportCounts): string {
  const parts = COUNT_LABELS.flatMap(([key, one, many]) => (counts[key] > 0 ? [`${formatNumber(counts[key], 0)} ${counts[key] === 1 ? one : many}`] : []))
  return parts.length > 0 ? parts.join(' · ') : 'nothing'
}

export function clashLabel(kind: string, name: string): string {
  return `A ${kind} named “${name}” already exists`
}
