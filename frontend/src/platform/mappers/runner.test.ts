import { describe, expect, it } from 'vitest'
import {
  runResponseKey,
  toPerformanceReport,
  toRun,
  toRunConfig,
  toRunConfigDto,
  toRunPlan,
  toRunSummaryView,
  type RunConfig,
} from './runner'

const configDto = {
  type: 'functional',
  environment_id: '',
  rows: [{ request_id: 'r1', alias_id: 'a1', selected: true }],
  functional: {
    iterations: 2,
    data_file: '/data/users.csv',
    delay_ms: 50,
    persist_responses: true,
    disable_logs: false,
    stop_on_error: true,
    keep_variable_values: true,
    skip_stored_cookies: false,
    save_cookies: false,
  },
  performance: {
    virtual_users: 10,
    duration_seconds: 600,
    data_file: '',
    profile: { kind: 'ramp_up', initial_users: 1, ramp_seconds: 150, base_users: 0, spike_seconds: 0, peak_seconds: 0 },
  },
}

const progressDto = { current_row: 1, current_name: 'Login', iteration: 0, completed: 1, planned: 2, tests_passed: 1, tests_failed: 0, elapsed_ms: 100, metrics: null }
const summaryDto = { iterations: 1, total_requests: 2, tests_passed: 1, tests_failed: 1, tests_skipped: 0, avg_response_ms: 50, duration_ms: 200, passed: 1, failed: 1, skipped: 0, errors: 0, cancelled: 0 }

const summaryViewDto = {
  id: 'R1',
  item_id: 'c1',
  target_kind: 'folder',
  target_name: 'Auth',
  type: 'performance',
  status: 'running',
  environment_name: 'dev',
  started_at: '2026-10-23T10:00:00Z',
  finished_at: '',
  progress: { ...progressDto, metrics: { active_users: 4, requests: 40, avg_ms: 12.5, throughput: 8, error_rate: 0.1 } },
  summary: summaryDto,
}

describe('runner mappers', () => {
  it('round-trips a run configuration', () => {
    const config = toRunConfig(configDto)
    expect(config).toEqual<RunConfig>({
      type: 'functional',
      environmentId: '',
      rows: [{ requestId: 'r1', aliasId: 'a1', selected: true }],
      functional: {
        iterations: 2,
        dataFile: '/data/users.csv',
        delayMs: 50,
        persistResponses: true,
        disableLogs: false,
        stopOnError: true,
        keepVariableValues: true,
        skipStoredCookies: false,
        saveCookies: false,
      },
      performance: {
        virtualUsers: 10,
        durationSeconds: 600,
        dataFile: '',
        profile: { kind: 'ramp_up', initialUsers: 1, rampSeconds: 150, baseUsers: 0, spikeSeconds: 0, peakSeconds: 0 },
      },
    })
    expect(toRunConfigDto(config)).toEqual(configDto)
  })

  it('falls back to known values for unknown run types and profiles', () => {
    const config = toRunConfig({ ...configDto, type: 'weird', performance: { ...configDto.performance, profile: { ...configDto.performance.profile, kind: '??' } } })
    expect(config.type).toBe('functional')
    expect(config.performance.profile.kind).toBe('fixed')
  })

  it('maps a plan with its rows, environments and active run', () => {
    const plan = toRunPlan({
      target: { item_id: 'c1', kind: 'collection', name: 'Orders API' },
      rows: [{ number: 1, request_id: 'r1', alias_id: '', selected: true, name: 'Login', method: 'POST', folder_path: null, aliases: [{ id: 'a1', name: 'mx' }] }],
      config: configDto,
      active_environment_id: '',
      environments: [{ id: 'e1', name: 'prod', color: 'red', confirm_before_send: true }],
      active_run: summaryViewDto,
      data_file_missing: true,
    })
    expect(plan.target).toEqual({ itemId: 'c1', kind: 'collection', name: 'Orders API' })
    expect(plan.rows).toEqual([{ number: 1, requestId: 'r1', aliasId: '', selected: true, name: 'Login', method: 'POST', folderPath: [], aliases: [{ id: 'a1', name: 'mx' }] }])
    expect(plan.activeEnvironmentId).toBeNull()
    expect(plan.environments).toEqual([{ id: 'e1', name: 'prod', color: 'red', confirmBeforeSend: true, error: null }])
    expect(plan.activeRun).toMatchObject({ id: 'R1', targetKind: 'folder', type: 'performance', status: 'running' })
    expect(plan.dataFileMissing).toBe(true)
  })

  it('maps the live metrics of a performance run', () => {
    expect(toRunSummaryView(summaryViewDto).progress.metrics).toEqual({ activeUsers: 4, requests: 40, avgMs: 12.5, throughput: 8, errorRate: 0.1 })
  })

  it('maps a stored run with its rows, tests and sent requests', () => {
    const run = toRun({
      id: 'R1',
      target: { item_id: 'c1', kind: 'collection', name: 'Orders API' },
      config: configDto,
      environment_name: 'dev',
      data_file_name: 'users.csv',
      status: 'completed',
      started_at: 's',
      finished_at: 'f',
      progress: progressDto,
      summary: summaryDto,
      performance: null,
      notices: null,
      rows: [
        {
          index: 0,
          iteration: 0,
          request_id: 'r1',
          alias_id: '',
          alias_name: '',
          name: 'Login',
          method: 'POST',
          folder_path: ['Auth'],
          url: 'https://api/login',
          status: 200,
          duration_ms: 98,
          size_bytes: 512,
          outcome: 'failed',
          tests: [
            { name: 'status is 200', status: 'passed', error: '' },
            { name: 'has token', status: 'failed', error: 'expected token' },
          ],
          script_errors: null,
          console: null,
          sent_request: { method: 'POST', url: 'https://api/login', headers: [{ key: 'Authorization', value: '•••••' }], body: '{}' },
          error: '',
          response_body_url: '',
        },
      ],
    })
    expect(run.notices).toEqual([])
    expect(run.performance).toBeNull()
    expect(run.rows[0]).toMatchObject({
      outcome: 'failed',
      folderPath: ['Auth'],
      testSummary: { total: 2, passed: 1, failed: 1, skipped: 0 },
      tests: [
        { index: 0, name: 'status is 200', status: 'passed', errorMessage: '' },
        { index: 1, name: 'has token', status: 'failed', errorMessage: 'expected token' },
      ],
      sentRequest: { headers: [{ key: 'Authorization', value: '•••••' }], body: '{}' },
    })
  })

  it('maps a performance report', () => {
    expect(toPerformanceReport(null)).toBeNull()
    const stats = { request_id: '', name: 'All', count: 10, avg_ms: 5, p90_ms: 9, p95_ms: 9.5, p99_ms: 10, min_ms: 1, max_ms: 10, throughput: 2, error_rate: 0.1, status_counts: { '200': 9, '500': 1 } }
    const report = toPerformanceReport({
      overall: stats,
      per_request: [{ ...stats, request_id: 'r1', name: 'Login', status_counts: null }],
      series: [{ second: 1, active_users: 2, requests: 3, avg_ms: 4, errors: 0 }],
    })
    expect(report?.overall.statusCounts).toEqual([
      { status: '200', count: 9 },
      { status: '500', count: 1 },
    ])
    expect(report?.overall).toMatchObject({ count: 10, p95Ms: 9.5 })
    expect(report?.perRequest[0]).toMatchObject({ requestId: 'r1', name: 'Login', statusCounts: [] })
    expect(report?.series).toEqual([{ second: 1, activeUsers: 2, requests: 3, avgMs: 4, errors: 0 }])
  })

  it('builds the response key of a run row', () => {
    expect(runResponseKey('R1', 3)).toBe('run:R1:3')
  })
})
