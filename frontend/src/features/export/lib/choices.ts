import type { ExportFormat, ExportFormatOption, ExportOptions, ExportRequest, ExportShape } from '@/platform/mappers/export'
import type { ExportChoices, ExportProgressState, ExportTarget } from '@/features/export/model'

export const MIN_PASSPHRASE_LENGTH = 8
const PROGRESS_THRESHOLD = 200

export type PassphraseProblem = 'too_short' | 'mismatch'

export function formatOption(options: ExportOptions, format: ExportFormat): ExportFormatOption | undefined {
  return options.formats.find((option) => option.format === format)
}

export function defaultChoices(options: ExportOptions): ExportChoices {
  const [first] = options.formats
  return {
    format: first?.format ?? 'xenon_yaml',
    shape: first?.shapes[0] ?? 'file',
    includeSecrets: false,
    passphrase: '',
    passphraseConfirmation: '',
    currentValuesAsInitial: false,
  }
}

/** Switching format keeps the shape when the new format offers it and drops "Include secrets" where it is not allowed. */
export function withFormat(options: ExportOptions, choices: ExportChoices, format: ExportFormat): ExportChoices {
  const option = formatOption(options, format)
  if (!option) {
    return choices
  }
  const shape: ExportShape = option.shapes.includes(choices.shape) ? choices.shape : (option.shapes[0] ?? 'file')
  return { ...choices, format, shape, includeSecrets: option.includeSecretsAllowed && choices.includeSecrets }
}

export function includesSecrets(options: ExportOptions, choices: ExportChoices): boolean {
  return choices.includeSecrets && formatOption(options, choices.format)?.includeSecretsAllowed === true
}

export function passphraseProblem(choices: ExportChoices): PassphraseProblem | null {
  if (choices.passphrase.length < MIN_PASSPHRASE_LENGTH) {
    return 'too_short'
  }
  return choices.passphrase === choices.passphraseConfirmation ? null : 'mismatch'
}

export function canExport(options: ExportOptions, choices: ExportChoices): boolean {
  const option = formatOption(options, choices.format)
  if (!option || !option.shapes.includes(choices.shape)) {
    return false
  }
  return !includesSecrets(options, choices) || passphraseProblem(choices) === null
}

export function suggestedFileName(options: ExportOptions, choices: ExportChoices): string {
  return formatOption(options, choices.format)?.suggestedFileName ?? options.suggestedFileName
}

export function exportRequest(target: ExportTarget, options: ExportOptions, choices: ExportChoices, targetPath: string): ExportRequest {
  const secrets = includesSecrets(options, choices)
  return {
    localId: target.localId,
    kind: target.kind,
    itemId: target.itemId,
    format: choices.format,
    shape: choices.shape,
    includeSecrets: secrets,
    passphrase: secrets ? choices.passphrase : '',
    passphraseConfirmation: secrets ? choices.passphraseConfirmation : '',
    currentValuesAsInitial: choices.currentValuesAsInitial,
    targetPath,
  }
}

/** Only exports of more than 200 files report progress (FR-018); smaller ones just show that they are running. */
export function showsProgress(progress: ExportProgressState | null): progress is ExportProgressState {
  return progress !== null && progress.total > PROGRESS_THRESHOLD
}
