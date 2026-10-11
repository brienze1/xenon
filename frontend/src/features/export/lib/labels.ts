import type { ExportCounts, ExportFormat, ExportSecrets, ExportShape } from '@/platform/mappers/export'
import { formatNumber } from '@/shared/lib/number-format'
import type { PassphraseProblem } from './choices'

const FORMAT_LABELS: Readonly<Record<ExportFormat, string>> = {
  xenon_yaml: 'Xenon YAML',
  postman: 'Postman',
}

const SHAPE_LABELS: Readonly<Record<ExportShape, string>> = {
  zip: 'Zip of the folder',
  bundle: 'Single YAML file',
  file: 'Single file',
}

const SECRETS_LABELS: Readonly<Record<ExportSecrets, string>> = {
  included: 'Secret values are included, encrypted with your passphrase.',
  omitted: 'Secret values are not included; they arrive as missing.',
  names_only: 'Secret variables are exported by name with empty values.',
}

const PASSPHRASE_PROBLEMS: Readonly<Record<PassphraseProblem, string>> = {
  too_short: 'Use at least 8 characters.',
  mismatch: 'The passphrases do not match.',
}

const COUNT_LABELS: ReadonlyArray<[keyof ExportCounts, string, string]> = [
  ['collections', 'collection', 'collections'],
  ['folders', 'folder', 'folders'],
  ['requests', 'request', 'requests'],
  ['examples', 'example', 'examples'],
  ['environments', 'environment', 'environments'],
  ['variables', 'variable', 'variables'],
]

export function formatLabel(format: ExportFormat): string {
  return FORMAT_LABELS[format]
}

export function shapeLabel(shape: ExportShape): string {
  return SHAPE_LABELS[shape]
}

export function secretsLabel(secrets: ExportSecrets): string {
  return SECRETS_LABELS[secrets]
}

export function passphraseProblemLabel(problem: PassphraseProblem): string {
  return PASSPHRASE_PROBLEMS[problem]
}

export function countsLabel(counts: ExportCounts): string {
  const parts = COUNT_LABELS.flatMap(([key, one, many]) => (counts[key] > 0 ? [`${formatNumber(counts[key], 0)} ${counts[key] === 1 ? one : many}`] : []))
  return parts.length > 0 ? parts.join(' · ') : 'nothing'
}
