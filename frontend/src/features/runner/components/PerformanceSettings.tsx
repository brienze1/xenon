import { useId } from 'react'
import type { LoadProfile, LoadProfileKind, PerformanceSettings as Settings } from '@/platform/mappers/runner'
import { curvePoints, profileParams, withProfileKind, withUsersAndDuration } from '@/features/runner/lib/load-profile'
import { useRunnerStore } from '@/features/runner/store'
import { DataFileField } from './DataFileField'
import { NumberSetting } from './NumberSetting'
import fields from './SettingFields.module.css'
import styles from './PerformanceSettings.module.css'

const MAX_USERS = 100
const MAX_DURATION_SECONDS = 3600
const CURVE_WIDTH = 240
const CURVE_HEIGHT = 48

const PROFILES: ReadonlyArray<{ kind: LoadProfileKind; label: string }> = [
  { kind: 'fixed', label: 'Fixed' },
  { kind: 'ramp_up', label: 'Ramp up' },
  { kind: 'spike', label: 'Spike' },
  { kind: 'peak', label: 'Peak' },
]

type ProfileParam = keyof Omit<LoadProfile, 'kind'>

const PARAM_FIELDS: Record<ProfileParam, { label: string; name: string; unit: string; users: boolean }> = {
  initialUsers: { label: 'Initial users', name: 'profile.initial_users', unit: 'users', users: true },
  rampSeconds: { label: 'Ramp duration', name: 'profile.ramp_seconds', unit: 's', users: false },
  baseUsers: { label: 'Base users', name: 'profile.base_users', unit: 'users', users: true },
  spikeSeconds: { label: 'Spike duration', name: 'profile.spike_seconds', unit: 's', users: false },
  peakSeconds: { label: 'Peak duration', name: 'profile.peak_seconds', unit: 's', users: false },
}

interface PerformanceSettingsProps {
  itemId: string
}

/** Virtual users, duration, load profile with its parameters, iteration data and a preview of the user curve. */
export function PerformanceSettings({ itemId }: PerformanceSettingsProps) {
  const profileId = useId()
  const settings = useRunnerStore((state) => state.targets[itemId]?.config?.performance ?? null)
  const fieldErrors = useRunnerStore((state) => state.targets[itemId]?.fieldErrors)
  const updateConfig = useRunnerStore((state) => state.updateConfig)
  if (settings === null) {
    return null
  }
  const minutes = Math.floor(settings.durationSeconds / 60)
  const seconds = settings.durationSeconds % 60

  function edit(update: (current: Settings) => Settings) {
    updateConfig(itemId, (config) => ({ ...config, performance: update(config.performance) }))
  }

  function editProfile(param: ProfileParam, value: number) {
    edit((current) => ({ ...current, profile: { ...current.profile, [param]: value } }))
  }

  return (
    <div className={fields.group}>
      <div className={fields.row}>
        <NumberSetting
          label="Virtual users"
          name="virtual_users"
          value={settings.virtualUsers}
          min={1}
          max={MAX_USERS}
          error={fieldErrors?.virtual_users}
          onChange={(users) => edit((current) => withUsersAndDuration(current, users, current.durationSeconds))}
        />
        <NumberSetting
          label="Duration"
          name="duration_minutes"
          unit="min"
          value={minutes}
          min={0}
          max={MAX_DURATION_SECONDS / 60}
          onChange={(value) => edit((current) => withUsersAndDuration(current, current.virtualUsers, value * 60 + (current.durationSeconds % 60)))}
        />
        <NumberSetting
          label="Seconds"
          name="duration_seconds"
          unit="s"
          value={seconds}
          min={0}
          max={59}
          error={fieldErrors?.duration_seconds}
          onChange={(value) => edit((current) => withUsersAndDuration(current, current.virtualUsers, Math.floor(current.durationSeconds / 60) * 60 + value))}
        />
      </div>
      <div className={fields.field}>
        <label htmlFor={profileId} className={fields.label}>
          Load profile
        </label>
        <select
          id={profileId}
          className={fields.select}
          value={settings.profile.kind}
          onChange={(event) => edit((current) => withProfileKind(current, event.target.value as LoadProfileKind))}
        >
          {PROFILES.map((profile) => (
            <option key={profile.kind} value={profile.kind}>
              {profile.label}
            </option>
          ))}
        </select>
      </div>
      {profileParams(settings.profile.kind).length > 0 ? (
        <div className={fields.row}>
          {profileParams(settings.profile.kind).map((param) => {
            const field = PARAM_FIELDS[param]
            return (
              <NumberSetting
                key={param}
                label={field.label}
                name={field.name}
                unit={field.unit}
                value={settings.profile[param]}
                min={1}
                max={field.users ? settings.virtualUsers : settings.durationSeconds}
                error={fieldErrors?.[field.name]}
                onChange={(value) => editProfile(param, value)}
              />
            )
          })}
        </div>
      ) : null}
      <figure className={styles.preview}>
        <svg
          viewBox={`0 0 ${CURVE_WIDTH} ${CURVE_HEIGHT}`}
          width={CURVE_WIDTH}
          height={CURVE_HEIGHT}
          role="img"
          aria-label={`Active users over ${settings.durationSeconds} seconds, up to ${settings.virtualUsers}`}
          className={styles.curve}
        >
          <polyline points={curvePoints(settings, CURVE_WIDTH, CURVE_HEIGHT)} className={styles.line} />
        </svg>
        <figcaption className={styles.caption}>
          Up to {settings.virtualUsers} users for {minutes} min {seconds} s
        </figcaption>
      </figure>
      <DataFileField itemId={itemId} path={settings.dataFile} />
    </div>
  )
}
