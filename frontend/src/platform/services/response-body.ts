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
