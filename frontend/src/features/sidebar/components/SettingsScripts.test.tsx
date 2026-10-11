import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyAuth } from '@/platform/mappers/request'
import type { ItemSettings } from '@/platform/mappers/tree'
import { useSidebarStore } from '@/features/sidebar/store'
import { ItemSettingsDialog } from './ItemSettingsDialog'
import type { ItemScriptsSlot } from './SettingsForm'

const folder: ItemSettings = {
  id: 'f1',
  kind: 'folder',
  name: 'Invoices',
  description: '',
  auth: { ...emptyAuth(), type: 'inherit', bearer: { token: '' }, basic: { username: '', password: '' } },
  headers: [],
  scripts: { preRequest: 'console.log(1)', postResponse: '', disabled: [] },
  inheritScripts: true,
}

function scriptsEditor(slot: ItemScriptsSlot) {
  return (
    <div>
      <p>Editing {slot.itemId}: {slot.scripts.preRequest}</p>
      <button type="button" onClick={() => slot.onChange({ ...slot.scripts, preRequest: 'edited()' })}>
        Edit script
      </button>
      <button type="button" onClick={() => slot.onInheritChange(false)}>
        Stop inheriting
      </button>
    </div>
  )
}

describe('Collection and folder scripts', () => {
  const saveSettings = vi.fn()

  beforeEach(() => {
    saveSettings.mockReset().mockResolvedValue(null)
    useSidebarStore.setState({ loadSettings: vi.fn().mockResolvedValue({ ok: true, value: folder }), saveSettings })
  })

  it('has a Scripts tab whose edits and inherit toggle are saved with the settings', async () => {
    const user = userEvent.setup()
    render(<ItemSettingsDialog itemId="f1" onClose={vi.fn()} scriptsEditor={scriptsEditor} />)
    await user.click(await screen.findByRole('tab', { name: 'Scripts' }))
    expect(screen.getByText('Editing f1: console.log(1)')).toBeInTheDocument()
    expect(screen.getByRole('tablist', { name: 'Folder settings' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Edit script' }))
    await user.click(screen.getByRole('button', { name: 'Stop inheriting' }))
    await user.click(screen.getByRole('button', { name: 'Save Settings' }))
    expect(saveSettings).toHaveBeenCalledWith(expect.objectContaining({ scripts: { preRequest: 'edited()', postResponse: '', disabled: [] }, inheritScripts: false }))
  })

  it('shows only the general settings without a scripts editor', async () => {
    render(<ItemSettingsDialog itemId="f1" onClose={vi.fn()} />)
    expect(await screen.findByRole('textbox', { name: 'Description' })).toBeInTheDocument()
    expect(screen.queryByRole('tab', { name: 'Scripts' })).not.toBeInTheDocument()
  })
})
