import { idleUpdateStatus, type UpdateStatus } from '@/platform/mappers/update-status'

export function updateStatus(patch: Partial<UpdateStatus> = {}): UpdateStatus {
  return {
    ...idleUpdateStatus(),
    currentVersion: '0.1.4',
    latestVersion: '0.1.7',
    releaseUrl: 'https://github.com/brienze1/xenon/releases/tag/v0.1.7',
    ...patch,
  }
}
