import * as ScriptService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/scriptservice'
import { toAppError } from '../errors'
import { toScriptMember, toScriptSnippet, toSyntaxErrors, type ScriptMember, type ScriptSnippet, type ScriptSyntaxError } from '../mappers/scripts'

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

/** Syntax errors of a script, compiled by Go without running it. */
export function checkScript(text: string): Promise<ScriptSyntaxError[]> {
  return call(async () => toSyntaxErrors(await ScriptService.Check({ text })))
}

export function scriptReference(): Promise<ScriptMember[]> {
  return call(async () => ((await ScriptService.Reference()).members ?? []).map(toScriptMember))
}

export function scriptSnippets(): Promise<ScriptSnippet[]> {
  return call(async () => ((await ScriptService.Snippets()) ?? []).map(toScriptSnippet))
}
