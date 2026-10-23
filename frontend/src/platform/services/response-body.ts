import { Clipboard } from '@wailsio/runtime'
import { AppError, toAppError } from '../errors'

const BODY_UNAVAILABLE = 'SND-03404'

export async function fetchResponseText(bodyUrl: string): Promise<string> {
  try {
    const response = await fetch(bodyUrl)
    if (!response.ok) {
      throw new AppError(BODY_UNAVAILABLE, 'The response body is no longer available')
    }
    return await response.text()
  } catch (err) {
    throw toAppError(err)
  }
}

/** The webview's own clipboard API is refused inside Wails, so the copy goes through the native clipboard. */
export async function copyResponseText(text: string): Promise<boolean> {
  try {
    await Clipboard.SetText(text)
    return true
  } catch (err) {
    throw toAppError(err)
  }
}
