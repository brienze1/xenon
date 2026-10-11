import { Clipboard } from '@wailsio/runtime'
import * as GitAccountService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/gitaccountservice'
import { AppError, toAppError } from '../errors'
import {
  toGitAccount,
  toGitHubDeviceCode,
  toGitHubSignInResult,
  toRemoteInfo,
  type GitAccount,
  type GitHubDeviceCode,
  type GitHubSignInResult,
  type RemoteInfo,
  type SshAccountInput,
  type TokenAccountInput,
} from '../mappers/git-accounts'

export const GITHUB_SIGN_IN_NOT_CONFIGURED = 'ACC-01503'
export const HOST_KEY_UNKNOWN = 'ACC-05409'
export const HOST_KEY_CHANGED = 'ACC-06409'

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

function required<T>(value: T | null): T {
  if (value === null) {
    throw new AppError('UNK-00500', 'Xenon returned no answer')
  }
  return value
}

export function listGitAccounts(): Promise<GitAccount[]> {
  return call(async () => ((await GitAccountService.List()) ?? []).map(toGitAccount))
}

export function addTokenAccount(input: TokenAccountInput): Promise<GitAccount> {
  return call(async () =>
    toGitAccount(
      required(
        await GitAccountService.AddToken({
          provider: input.provider,
          host: input.host,
          username: input.username,
          token: input.token,
          repository: input.repository,
          display_name: '',
          commit_name: input.commitName,
          commit_email: input.commitEmail,
        }),
      ),
    ),
  )
}

export function addSshAccount(input: SshAccountInput): Promise<{ account: GitAccount; publicKey: string }> {
  return call(async () => {
    const result = required(
      await GitAccountService.AddSSH({
        mode: input.mode,
        host: input.host,
        key_path: input.keyPath,
        passphrase: input.passphrase,
        display_name: '',
        commit_name: input.commitName,
        commit_email: input.commitEmail,
      }),
    )
    return { account: toGitAccount(result.account), publicKey: result.public_key }
  })
}

export function updateGitAccount(id: string, displayName: string, commitName: string, commitEmail: string): Promise<GitAccount> {
  return call(async () =>
    toGitAccount(required(await GitAccountService.Update({ id, display_name: displayName, commit_name: commitName, commit_email: commitEmail }))),
  )
}

export function removeGitAccount(id: string): Promise<void> {
  return call(() => GitAccountService.Remove({ id }))
}

export function startGitHubSignIn(): Promise<GitHubDeviceCode> {
  return call(async () => toGitHubDeviceCode(required(await GitAccountService.StartGitHubSignIn())))
}

export function pollGitHubSignIn(sessionId: string): Promise<GitHubSignInResult> {
  return call(async () => toGitHubSignInResult(required(await GitAccountService.PollGitHubSignIn({ session_id: sessionId }))))
}

export function uploadPublicKey(sshAccountId: string, gitHubAccountId: string, title: string): Promise<void> {
  return call(() => GitAccountService.UploadPublicKey({ ssh_account_id: sshAccountId, github_account_id: gitHubAccountId, title }))
}

export function testConnection(accountId: string, repository: string): Promise<RemoteInfo> {
  return call(async () => toRemoteInfo(required(await GitAccountService.TestConnection({ account_id: accountId, repository }))))
}

export function confirmHostKey(repository: string, fingerprint: string, accept: boolean): Promise<void> {
  return call(() => GitAccountService.ConfirmHostKey({ repository, fingerprint, accept }))
}

/** Public keys are not secret, so the webview may put them on the clipboard. */
export function copyPublicKey(publicKey: string): Promise<void> {
  return call(() => Clipboard.SetText(publicKey))
}
