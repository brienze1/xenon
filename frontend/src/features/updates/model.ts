/**
 * The "ready to install" toast's life in one app run: `unseen` until an automatic download is
 * verified, `open` while it shows, `closed` after Restart now, Later or × — it never reopens.
 */
export type ReadyToastPhase = 'unseen' | 'open' | 'closed'
