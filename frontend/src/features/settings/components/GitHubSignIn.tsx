import { useEffect } from 'react'
import { withNextStep } from '@/shared/lib/next-step'
import { Button } from '@/shared/ui/button/Button'
import { useGitAccountsStore } from '@/features/settings/store'
import styles from './GitAccounts.module.css'

const MS_PER_SECOND = 1000

interface GitHubSignInProps {
  onClose: () => void
}

export function GitHubSignIn({ onClose }: GitHubSignInProps) {
  const signIn = useGitAccountsStore((state) => state.signIn)
  const startSignIn = useGitAccountsStore((state) => state.startSignIn)
  const pollSignIn = useGitAccountsStore((state) => state.pollSignIn)
  const resetSignIn = useGitAccountsStore((state) => state.resetSignIn)

  useEffect(() => {
    void startSignIn()
    return resetSignIn
  }, [startSignIn, resetSignIn])

  const { phase, interval } = signIn
  useEffect(() => {
    if (phase !== 'waiting') {
      return undefined
    }
    const timer = window.setTimeout(() => void pollSignIn(), Math.max(interval, 1) * MS_PER_SECOND)
    return () => window.clearTimeout(timer)
  }, [phase, interval, signIn.code, pollSignIn])

  return (
    <section className={styles.panel} aria-label="GitHub sign-in" aria-live="polite">
      <SignInMessage />
      <div className={styles.actions}>
        <Button size="sm" variant={phase === 'done' ? 'primary' : 'secondary'} onClick={onClose}>
          {phase === 'done' ? 'Done' : 'Cancel'}
        </Button>
      </div>
    </section>
  )
}

function SignInMessage() {
  const signIn = useGitAccountsStore((state) => state.signIn)
  switch (signIn.phase) {
    case 'waiting':
      return (
        <>
          <p className={styles.hint}>Enter this code on the GitHub page that opened in your browser ({signIn.code?.verificationUri}):</p>
          <p className={styles.code}>{signIn.code?.userCode}</p>
          <p className={styles.hint}>Waiting for you to approve Xenon on GitHub…</p>
        </>
      )
    case 'done':
      return <p className={styles.hint}>Signed in. The account is listed above.</p>
    case 'expired':
      return <p className={styles.hint}>The code expired. Close and start the sign-in again.</p>
    case 'denied':
      return <p className={styles.hint}>The sign-in was denied on GitHub.</p>
    case 'not_configured':
      return (
        <p role="alert" className={styles.hint}>
          GitHub sign-in is not configured in this build of Xenon. Add the account with a personal access token instead.
        </p>
      )
    case 'error':
      return (
        <p role="alert" className={styles.hint}>
          {withNextStep(signIn.error?.description || signIn.error?.code || '', 'Try again.')}
        </p>
      )
    default:
      return <p className={styles.hint}>Asking GitHub for a sign-in code…</p>
  }
}
