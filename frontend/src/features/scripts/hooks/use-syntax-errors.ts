import { useEffect, useState } from 'react'
import type { ScriptSyntaxError } from '@/platform/mappers/scripts'
import { useScriptsStore } from '@/features/scripts/store'

export const CHECK_DELAY_MS = 300

const NO_ERRORS: ScriptSyntaxError[] = []

export function useSyntaxErrors(text: string): ScriptSyntaxError[] {
  const check = useScriptsStore((state) => state.check)
  const [checked, setChecked] = useState<{ text: string; errors: ScriptSyntaxError[] }>({ text: '', errors: NO_ERRORS })
  useEffect(() => {
    let active = true
    const timer = window.setTimeout(() => {
      void check(text).then((errors) => {
        if (active) {
          setChecked({ text, errors })
        }
      })
    }, CHECK_DELAY_MS)
    return () => {
      active = false
      window.clearTimeout(timer)
    }
  }, [check, text])
  return checked.text === text ? checked.errors : NO_ERRORS
}
