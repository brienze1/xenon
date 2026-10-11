import type {
  GitAccount as GitAccountDto,
  GitHubDeviceCode as GitHubDeviceCodeDto,
  GitHubSignInResult as GitHubSignInResultDto,
  GitRemoteInfo as GitRemoteInfoDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export type GitProvider = 'github' | 'gitlab' | 'bitbucket' | 'azure_devops' | 'other'
export type SshMode = 'agent' | 'key_file' | 'generate'
export type SignInStatus = 'pending' | 'done' | 'expired' | 'denied'

/** A Git account as Go lists it; secrets never cross the bridge. */
export interface GitAccount {
  id: string
  kind: 'github' | 'token' | 'ssh'
  provider: string
  sshMode: string
  displayName: string
  host: string
  username: string
  commitName: string
  commitEmail: string
  avatarUrl: string
  publicKey: string
  hasPassphrase: boolean
}

export interface TokenAccountInput {
  provider: GitProvider
  host: string
  username: string
  token: string
  repository: string
  commitName: string
  commitEmail: string
}

export interface SshAccountInput {
  mode: SshMode
  host: string
  keyPath: string
  passphrase: string
  commitName: string
  commitEmail: string
}

export interface GitHubDeviceCode {
  sessionId: string
  userCode: string
  verificationUri: string
  expiresIn: number
  interval: number
}

export interface GitHubSignInResult {
  status: SignInStatus
  interval: number
  account: GitAccount | null
}

export interface RemoteInfo {
  defaultBranch: string
  branches: string[]
}

export function toGitAccount(dto: GitAccountDto): GitAccount {
  return {
    id: dto.id,
    kind: dto.kind as GitAccount['kind'],
    provider: dto.provider,
    sshMode: dto.ssh_mode,
    displayName: dto.display_name,
    host: dto.host,
    username: dto.username,
    commitName: dto.commit_name,
    commitEmail: dto.commit_email,
    avatarUrl: dto.avatar_url,
    publicKey: dto.public_key,
    hasPassphrase: dto.has_passphrase,
  }
}

export function toGitHubDeviceCode(dto: GitHubDeviceCodeDto): GitHubDeviceCode {
  return { sessionId: dto.session_id, userCode: dto.user_code, verificationUri: dto.verification_uri, expiresIn: dto.expires_in, interval: dto.interval }
}

export function toGitHubSignInResult(dto: GitHubSignInResultDto): GitHubSignInResult {
  return { status: dto.status as SignInStatus, interval: dto.interval, account: dto.account ? toGitAccount(dto.account) : null }
}

export function toRemoteInfo(dto: GitRemoteInfoDto): RemoteInfo {
  return { defaultBranch: dto.default_branch, branches: dto.branches ?? [] }
}
