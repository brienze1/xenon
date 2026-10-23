import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '../errors'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/runnerservice', () => ({
  Plan: vi.fn(),
  SavePreferences: vi.fn(),
  Start: vi.fn(),
  Pause: vi.fn(),
  Resume: vi.fn(),
  Stop: vi.fn(),
  Get: vi.fn(),
  List: vi.fn(),
  Active: vi.fn(),
  Response: vi.fn(),
  Delete: vi.fn(),
  Clear: vi.fn(),
  Export: vi.fn(),
}))

const RunnerService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/runnerservice')
const runner = await import('./runner')
const { toRunConfig, toRunConfigDto } = await import('../mappers/runner')

const configDto = {
  type: 'functional',
  environment_id: 'none',
  rows: [{ request_id: 'r1', alias_id: '', selected: true }],
  functional: { iterations: 1, data_file: '', delay_ms: 0, persist_responses: false, disable_logs: false, stop_on_error: false, keep_variable_values: true, skip_stored_cookies: false, save_cookies: false },
  performance: { virtual_users: 10, duration_seconds: 600, data_file: '', profile: { kind: 'fixed', initial_users: 0, ramp_seconds: 0, base_users: 0, spike_seconds: 0, peak_seconds: 0 } },
}
const progress = { current_row: 0, current_name: '', iteration: 0, completed: 0, planned: 1, tests_passed: 0, tests_failed: 0, elapsed_ms: 0, metrics: null }
const summary = { iterations: 1, total_requests: 0, tests_passed: 0, tests_failed: 0, tests_skipped: 0, avg_response_ms: 0, duration_ms: 0, passed: 0, failed: 0, skipped: 0, errors: 0, cancelled: 0 }
const view = { id: 'R1', item_id: 'c1', target_kind: 'collection', target_name: 'Orders', type: 'functional', status: 'running', environment_name: '', started_at: 's', finished_at: '', progress, summary }

describe('runner service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('plans a target', async () => {
    vi.mocked(RunnerService.Plan).mockResolvedValue({
      target: { item_id: 'c1', kind: 'collection', name: 'Orders' },
      rows: null,
      config: configDto,
      active_environment_id: 'e1',
      environments: null,
      active_run: null,
      data_file_missing: false,
    } as never)
    const plan = await runner.planRun('L1', 'c1')
    expect(RunnerService.Plan).toHaveBeenCalledWith({ local_id: 'L1', item_id: 'c1' })
    expect(plan).toMatchObject({ rows: [], environments: [], activeEnvironmentId: 'e1', activeRun: null })
  })

  it('saves preferences and starts a run with the acknowledgement', async () => {
    const config = toRunConfig(configDto)
    vi.mocked(RunnerService.SavePreferences).mockResolvedValue()
    vi.mocked(RunnerService.Start).mockResolvedValue(view as never)
    await runner.saveRunnerPreferences('L1', 'c1', config)
    expect(RunnerService.SavePreferences).toHaveBeenCalledWith({ local_id: 'L1', item_id: 'c1', config: toRunConfigDto(config) })
    await expect(runner.startRun('L1', 'c1', config, { confirmedEnvironmentId: 'e1', missingSecrets: true })).resolves.toMatchObject({ id: 'R1', status: 'running' })
    expect(RunnerService.Start).toHaveBeenCalledWith({ local_id: 'L1', item_id: 'c1', config: configDto, confirmed_environment_id: 'e1', acknowledge_missing_secrets: true })
  })

  it('controls, reads, lists and deletes runs', async () => {
    vi.mocked(RunnerService.Pause).mockResolvedValue({ ...view, status: 'paused' } as never)
    vi.mocked(RunnerService.Resume).mockResolvedValue(view as never)
    vi.mocked(RunnerService.Stop).mockResolvedValue({ ...view, status: 'stopped' } as never)
    vi.mocked(RunnerService.List).mockResolvedValue(null as never)
    vi.mocked(RunnerService.Active).mockResolvedValue([view] as never)
    vi.mocked(RunnerService.Delete).mockResolvedValue()
    vi.mocked(RunnerService.Clear).mockResolvedValue()
    vi.mocked(RunnerService.Export).mockResolvedValue()
    await expect(runner.pauseRun('L1', 'R1')).resolves.toMatchObject({ status: 'paused' })
    await expect(runner.resumeRun('L1', 'R1')).resolves.toMatchObject({ status: 'running' })
    await expect(runner.stopRun('L1', 'R1')).resolves.toMatchObject({ status: 'stopped' })
    expect(RunnerService.Stop).toHaveBeenCalledWith({ local_id: 'L1', run_id: 'R1' })
    await expect(runner.listRuns('L1', 'c1')).resolves.toEqual([])
    await expect(runner.activeRuns('L1')).resolves.toHaveLength(1)
    expect(RunnerService.Active).toHaveBeenCalledWith({ local_id: 'L1' })
    await runner.deleteRun('L1', 'R1')
    await runner.clearRuns('L1', 'c1')
    await runner.exportRun('L1', 'R1', 'junit', '/tmp/run.xml')
    expect(RunnerService.Clear).toHaveBeenCalledWith({ local_id: 'L1', item_id: 'c1' })
    expect(RunnerService.Export).toHaveBeenCalledWith({ local_id: 'L1', run_id: 'R1', format: 'junit', path: '/tmp/run.xml' })
  })

  it('reads a persisted response and rejects empty answers', async () => {
    vi.mocked(RunnerService.Response).mockResolvedValue({
      tab_id: '',
      send_id: 's1',
      outcome: 'sent',
      status_text: 'OK',
      proto: 'HTTP/1.1',
      content_type: 'application/json',
      body_url: '/body/run:R1:0',
      tests: null,
      script_errors: null,
      headers: null,
      cookies_set: null,
      timings: { dns_ms: 0, connect_ms: 0, tls_ms: 0, first_byte_ms: 0, download_ms: 0, total_ms: 5 },
      test_summary: { total: 0, passed: 0, failed: 0, skipped: 0 },
      status_code: 200,
      size_bytes: 2,
      ok: true,
      is_binary: false,
    } as never)
    await expect(runner.runResponse('L1', 'R1', 0)).resolves.toMatchObject({ statusCode: 200, bodyUrl: '/body/run:R1:0' })
    expect(RunnerService.Response).toHaveBeenCalledWith({ local_id: 'L1', run_id: 'R1', index: 0 })
    vi.mocked(RunnerService.Get).mockResolvedValue(null as never)
    await expect(runner.getRun('L1', 'R1')).rejects.toMatchObject({ code: 'UNK-00204' })
  })

  it('maps backend errors', async () => {
    vi.mocked(RunnerService.Start).mockRejectedValue({ cause: { error: { code: 'RUN-02409', description: 'busy' } } })
    await expect(runner.startRun('L1', 'c1', toRunConfig(configDto), { confirmedEnvironmentId: '', missingSecrets: false })).rejects.toMatchObject({ code: 'RUN-02409' })
  })

  it('reads the field errors of a refused start', () => {
    expect(runner.runFieldErrorsOf(new AppError('RUN-04400', 'bad', [{ attribute: 'iterations', messages: ['must be at least 1'] }]))).toEqual({ iterations: 'must be at least 1' })
    expect(runner.runFieldErrorsOf(new AppError('RUN-05400', 'bad file', [{ attribute: 'reason', messages: ['not an array'] }]))).toEqual({ data_file: 'not an array' })
    expect(runner.runFieldErrorsOf(new AppError('RUN-05400', 'bad file'))).toEqual({ data_file: 'bad file' })
    expect(runner.runFieldErrorsOf(new AppError('RUN-03400', 'empty'))).toEqual({})
  })
})
