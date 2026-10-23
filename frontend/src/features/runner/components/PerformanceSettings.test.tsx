import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/services/runner', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/platform/services/runner')>()),
  saveRunnerPreferences: vi.fn(async () => undefined),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const { useRunnerStore } = await import('@/features/runner/store')
const { readyTarget, runConfig, runPlan } = await import('@/features/runner/test-fixtures')
const { PerformanceSettings } = await import('./PerformanceSettings')

function performance() {
  return useRunnerStore.getState().targets.c1!.config!.performance
}

describe('PerformanceSettings', () => {
  beforeEach(() => {
    useRunnerStore.getState().reset()
    useRunnerStore.setState({ targets: { c1: readyTarget(runPlan({ config: runConfig({ type: 'performance' }) })) } })
  })

  it('shows users, duration in minutes and seconds, a fixed profile and the curve preview', () => {
    render(<PerformanceSettings itemId="c1" />)
    expect(screen.getByRole('spinbutton', { name: 'Virtual users' })).toHaveValue(10)
    expect(screen.getByRole('spinbutton', { name: /Duration/ })).toHaveValue(10)
    expect(screen.getByRole('spinbutton', { name: /Seconds/ })).toHaveValue(0)
    expect(screen.getByRole('combobox', { name: 'Load profile' })).toHaveValue('fixed')
    expect(screen.getByRole('img', { name: 'Active users over 600 seconds, up to 10' }).querySelector('polyline')).toHaveAttribute('points', expect.stringMatching(/^0,0 /))
    expect(screen.getByRole('group', { name: 'Iteration data' })).toBeInTheDocument()
  })

  it('fills the profile defaults and recomputes them with the duration until edited', async () => {
    render(<PerformanceSettings itemId="c1" />)
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Load profile' }), 'ramp_up')
    expect(screen.getByRole('spinbutton', { name: /Initial users/ })).toHaveValue(1)
    expect(screen.getByRole('spinbutton', { name: /Ramp duration/ })).toHaveValue(150)
    const minutes = screen.getByRole('spinbutton', { name: /Duration/ })
    await userEvent.clear(minutes)
    await userEvent.type(minutes, '2')
    expect(performance()).toMatchObject({ durationSeconds: 120, profile: { rampSeconds: 30 } })
    const ramp = screen.getByRole('spinbutton', { name: /Ramp duration/ })
    await userEvent.clear(ramp)
    await userEvent.type(ramp, '45')
    await userEvent.clear(minutes)
    await userEvent.type(minutes, '4')
    expect(performance()).toMatchObject({ durationSeconds: 240, profile: { rampSeconds: 45 } })
  })

  it('shows the spike and peak parameters with their defaults', async () => {
    render(<PerformanceSettings itemId="c1" />)
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Load profile' }), 'spike')
    expect(screen.getByRole('spinbutton', { name: /Spike duration/ })).toHaveValue(60)
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Load profile' }), 'peak')
    expect(screen.getByRole('spinbutton', { name: /Peak duration/ })).toHaveValue(300)
    expect(screen.getByRole('spinbutton', { name: /Base users/ })).toHaveValue(1)
  })

  it('shows a base users error inline', () => {
    useRunnerStore.setState({
      targets: {
        c1: readyTarget(runPlan({ config: runConfig({ type: 'performance', performance: { ...runConfig().performance, profile: { ...runConfig().performance.profile, kind: 'spike', baseUsers: 50, spikeSeconds: 10 } } }) }), {
          fieldErrors: { 'profile.base_users': 'must be between 1 and the virtual users' },
        }),
      },
    })
    render(<PerformanceSettings itemId="c1" />)
    expect(screen.getByRole('alert')).toHaveTextContent('must be between 1 and the virtual users')
    expect(screen.getByRole('spinbutton', { name: /Base users/ })).toHaveAttribute('aria-invalid', 'true')
  })
})
