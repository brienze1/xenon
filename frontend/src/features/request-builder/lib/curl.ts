const CURL_COMMAND = /^\s*curl(\.exe)?\s/i

/** Pasted text that starts like a cURL command (`curl …` or PowerShell's `curl.exe …`). */
export function isCurlCommand(text: string): boolean {
  return CURL_COMMAND.test(text)
}
