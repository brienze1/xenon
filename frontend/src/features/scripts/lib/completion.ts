import type { ScriptMember } from '@/platform/mappers/scripts'
import type { EditorCompletions } from '@/shared/ui/code-editor/editor-extensions'

const VARIABLE_CALL = /pm\.(?:environment|globals|collectionVariables|variables)\.(?:get|set|unset|has)\(\s*["'`]([^"'`\n]*)$/
const MEMBER_CHAIN = /(?:^|[^\w$.])((?:[A-Za-z_$][\w$]*\.)*)([A-Za-z_$][\w$]*)?$/

export interface VariableNameQuery {
  from: number
  prefix: string
}

export interface VariableOption {
  name: string
  detail: string
}

/** The variable name being typed inside `pm.<scope>.get/set/unset/has("…`, or `null`. */
export function variableNameQuery(text: string, position: number): VariableNameQuery | null {
  const match = VARIABLE_CALL.exec(text.slice(0, position))
  if (!match) {
    return null
  }
  const prefix = match[1] ?? ''
  return { from: position - prefix.length, prefix }
}

export function variableCompletions(query: VariableNameQuery, options: readonly VariableOption[]): EditorCompletions {
  return { from: query.from, options: options.map((option) => ({ label: option.name, detail: option.detail, apply: option.name })) }
}

function parentOf(path: string): string {
  const dot = path.lastIndexOf('.')
  return dot < 0 ? '' : path.slice(0, dot)
}

function nameOf(path: string): string {
  return path.slice(path.lastIndexOf('.') + 1)
}

/** Members of the object before the cursor (`pm.re` → pm.request, pm.response…), or globals for a bare name. */
export function memberCompletions(members: readonly ScriptMember[], text: string, position: number): EditorCompletions | null {
  const match = MEMBER_CHAIN.exec(text.slice(0, position))
  if (!match) {
    return null
  }
  const parent = (match[1] ?? '').replace(/\.$/, '')
  const partial = match[2] ?? ''
  if (parent === '' && partial === '') {
    return null
  }
  const lowered = partial.toLowerCase()
  const options = members
    .filter((member) => parentOf(member.path) === parent && nameOf(member.path).toLowerCase().startsWith(lowered))
    .map((member) => ({ label: nameOf(member.path), detail: member.description, apply: nameOf(member.path) }))
  return options.length === 0 ? null : { from: position - partial.length, options }
}
