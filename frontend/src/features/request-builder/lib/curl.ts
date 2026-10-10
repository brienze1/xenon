import type { RequestDraft } from '@/platform/mappers/request'

const CURL_COMMAND = /^\s*curl(\.exe)?\s/i

/** Pasted text that starts like a cURL command (`curl …` or PowerShell's `curl.exe …`). */
export function isCurlCommand(text: string): boolean {
  return CURL_COMMAND.test(text)
}

/** The current draft with everything a cURL command describes replaced; its id, name, description and turned-off auto headers stay. */
export function withPastedRequest(current: RequestDraft, pasted: RequestDraft): RequestDraft {
  return {
    ...current,
    method: pasted.method,
    url: pasted.url,
    params: pasted.params,
    pathVariables: pasted.pathVariables,
    headers: pasted.headers,
    auth: pasted.auth,
    body: pasted.body,
    settings: pasted.settings,
  }
}
