import type { LoadProfile, LoadProfileKind, PerformanceSettings } from '@/platform/mappers/runner'

type ProfileParam = keyof Omit<LoadProfile, 'kind'>

const PARAMS_BY_KIND: Record<LoadProfileKind, readonly ProfileParam[]> = {
  fixed: [],
  ramp_up: ['initialUsers', 'rampSeconds'],
  spike: ['baseUsers', 'spikeSeconds'],
  peak: ['baseUsers', 'peakSeconds'],
}

/** The parameters a load profile takes. */
export function profileParams(kind: LoadProfileKind): readonly ProfileParam[] {
  return PARAMS_BY_KIND[kind]
}

/** FR-011 defaults of a profile parameter for the given user count and duration. */
export function defaultParam(param: ProfileParam, virtualUsers: number, durationSeconds: number): number {
  const tenthOfUsers = Math.max(1, Math.floor(virtualUsers / 10))
  switch (param) {
    case 'initialUsers':
      return 1
    case 'rampSeconds':
      return Math.max(1, Math.floor(durationSeconds / 4))
    case 'baseUsers':
      return tenthOfUsers
    case 'spikeSeconds':
      return Math.max(1, Math.floor(durationSeconds / 10))
    case 'peakSeconds':
      return Math.max(1, Math.floor(durationSeconds / 2))
  }
}

/** Switching the profile fills the parameters the new one takes that are still unset. */
export function withProfileKind(settings: PerformanceSettings, kind: LoadProfileKind): PerformanceSettings {
  const profile: LoadProfile = { ...settings.profile, kind }
  for (const param of PARAMS_BY_KIND[kind]) {
    if (profile[param] < 1) {
      profile[param] = defaultParam(param, settings.virtualUsers, settings.durationSeconds)
    }
  }
  return { ...settings, profile }
}

/**
 * New users or duration: each parameter of the profile that still holds the default of the old
 * values is recomputed; a parameter the user edited keeps its value (FR-011).
 */
export function withUsersAndDuration(settings: PerformanceSettings, virtualUsers: number, durationSeconds: number): PerformanceSettings {
  const { kind } = settings.profile
  const profile: LoadProfile = { ...settings.profile }
  for (const param of PARAMS_BY_KIND[kind]) {
    if (profile[param] === defaultParam(param, settings.virtualUsers, settings.durationSeconds)) {
      profile[param] = defaultParam(param, virtualUsers, durationSeconds)
    }
  }
  return { ...settings, virtualUsers, durationSeconds, profile }
}

/** The active users at `second` of the run (FR-043), for the curve preview. */
export function usersAt(settings: PerformanceSettings, second: number): number {
  const { virtualUsers: max, durationSeconds: duration, profile } = settings
  switch (profile.kind) {
    case 'fixed':
      return max
    case 'ramp_up': {
      const ramp = Math.max(1, profile.rampSeconds)
      return second >= ramp ? max : profile.initialUsers + ((max - profile.initialUsers) * second) / ramp
    }
    case 'spike': {
      const start = (duration - profile.spikeSeconds) / 2
      return second >= start && second < start + profile.spikeSeconds ? max : profile.baseUsers
    }
    case 'peak': {
      const ramp = Math.max(0, (duration - profile.peakSeconds) / 2)
      if (second < ramp) {
        return profile.baseUsers + ((max - profile.baseUsers) * second) / ramp
      }
      if (second < ramp + profile.peakSeconds) {
        return max
      }
      return ramp === 0 ? max : max - ((max - profile.baseUsers) * (second - ramp - profile.peakSeconds)) / ramp
    }
  }
}

/** The user curve as SVG polyline points in a `width` × `height` box. */
export function curvePoints(settings: PerformanceSettings, width: number, height: number, samples = 60): string {
  const duration = Math.max(1, settings.durationSeconds)
  const max = Math.max(1, settings.virtualUsers)
  const points: string[] = []
  for (let step = 0; step <= samples; step++) {
    const second = (duration * step) / samples
    const users = Math.min(max, Math.max(0, usersAt(settings, Math.min(second, duration - 1e-9))))
    points.push(`${round((width * step) / samples)},${round(height - (height * users) / max)}`)
  }
  return points.join(' ')
}

function round(value: number): number {
  return Math.round(value * 10) / 10
}
