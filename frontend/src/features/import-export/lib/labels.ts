import type { ImportCounts } from '@/platform/mappers/import'

const KIND_LABELS: Readonly<Record<string, string>> = {
  postman_collection: 'Postman collection',
  postman_environment: 'Postman environment',
  postman_globals: 'Postman globals',
  postman_dump: 'Postman data dump',
  bruno_bru: 'Bruno collection (.bru)',
  bruno_yaml: 'Bruno collection (YAML)',
  curl: 'cURL command',
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

export function sourceLabel(kind: string, version: string): string {
  const label = KIND_LABELS[kind] ?? kind
  return version ? `${label} v${version}` : label
}

export function countsLabel(counts: ImportCounts): string {
  const parts = COUNT_LABELS.filter(([key]) => counts[key] > 0).map(([key, one, many]) => `${counts[key]} ${counts[key] === 1 ? one : many}`)
  return parts.length > 0 ? parts.join(' · ') : 'nothing'
}

export function clashLabel(kind: string, name: string): string {
  return `A ${kind} named “${name}” already exists`
}
