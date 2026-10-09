import { describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/dialogservice', () => ({
  PickFolder: vi.fn(),
}))

const DialogService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/dialogservice')
const { pickFolder } = await import('./dialogs')

describe('pickFolder', () => {
  it('returns the chosen folder', async () => {
    vi.mocked(DialogService.PickFolder).mockResolvedValue({ path: '/work/acme' } as never)
    await expect(pickFolder('Choose')).resolves.toBe('/work/acme')
    expect(DialogService.PickFolder).toHaveBeenCalledWith({ title: 'Choose' })
  })

  it('returns null when the picker is cancelled', async () => {
    vi.mocked(DialogService.PickFolder).mockResolvedValue({ path: '' } as never)
    await expect(pickFolder('Choose')).resolves.toBeNull()
  })

  it('converts failures into AppError', async () => {
    vi.mocked(DialogService.PickFolder).mockRejectedValue(new Error('no window'))
    await expect(pickFolder('Choose')).rejects.toMatchObject({ code: 'UNK-00500' })
  })
})
