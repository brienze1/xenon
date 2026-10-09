import { describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/dialogservice', () => ({
  PickFolder: vi.fn(),
  PickFile: vi.fn(),
  PickSaveFile: vi.fn(),
}))

const DialogService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/dialogservice')
const { pickFile, pickFolder, pickSaveFile } = await import('./dialogs')

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

describe('pickFile and pickSaveFile', () => {
  it('return the chosen paths', async () => {
    vi.mocked(DialogService.PickFile).mockResolvedValue({ path: '/a.bin' } as never)
    vi.mocked(DialogService.PickSaveFile).mockResolvedValue({ path: '' } as never)
    await expect(pickFile('Choose a file')).resolves.toBe('/a.bin')
    await expect(pickSaveFile('Save', 'body.json')).resolves.toBeNull()
    expect(DialogService.PickSaveFile).toHaveBeenCalledWith({ title: 'Save', suggested_name: 'body.json' })
  })

  it('convert failures into AppError', async () => {
    vi.mocked(DialogService.PickFile).mockRejectedValue(new Error('x'))
    vi.mocked(DialogService.PickSaveFile).mockRejectedValue(new Error('x'))
    await expect(pickFile('Choose')).rejects.toMatchObject({ code: 'UNK-00500' })
    await expect(pickSaveFile('Save', 'a')).rejects.toMatchObject({ code: 'UNK-00500' })
  })
})
