import { AppError } from './errors'

const BODY_UNAVAILABLE = 'SND-03404'

async function fetchBody(bodyUrl: string): Promise<Response> {
  const response = await fetch(bodyUrl)
  if (!response.ok) {
    throw new AppError(BODY_UNAVAILABLE, 'The response body is no longer available')
  }
  return response
}

export async function fetchResponseText(bodyUrl: string): Promise<string> {
  return (await fetchBody(bodyUrl)).text()
}

export async function fetchResponseBlob(bodyUrl: string): Promise<Blob> {
  return (await fetchBody(bodyUrl)).blob()
}
