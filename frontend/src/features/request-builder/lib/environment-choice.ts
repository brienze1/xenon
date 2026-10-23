import type { EnvironmentSummary } from '@/platform/mappers/environment'

/** An alias's Environments setting while it is being edited: all environments, or only the ticked ones. */
export interface EnvironmentChoice {
  mode: 'all' | 'only'
  ids: string[]
}

export function choiceOf(environmentIds: string[]): EnvironmentChoice {
  return environmentIds.length === 0 ? { mode: 'all', ids: [] } : { mode: 'only', ids: environmentIds }
}

/** The environments an alias can be limited to: every readable one, in the switcher's order. */
export function pickableEnvironments(environments: EnvironmentSummary[]): EnvironmentSummary[] {
  return environments.filter((environment) => environment.error === null)
}

/** The ids to save: none for all environments, else the ticked environments that still exist, in the switcher's order. */
export function chosenIds(choice: EnvironmentChoice, environments: EnvironmentSummary[]): string[] {
  if (choice.mode === 'all') {
    return []
  }
  return pickableEnvironments(environments)
    .filter((environment) => choice.ids.includes(environment.id))
    .map((environment) => environment.id)
}

export function isChoiceComplete(choice: EnvironmentChoice, environments: EnvironmentSummary[]): boolean {
  return choice.mode === 'all' || chosenIds(choice, environments).length > 0
}
