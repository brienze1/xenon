import type {
  ScriptCheck as ScriptCheckDto,
  ScriptMember as ScriptMemberDto,
  ScriptSnippet as ScriptSnippetDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export type ScriptMemberKind = 'object' | 'function' | 'property'
export type SnippetPhase = 'any' | 'preRequest' | 'postResponse'

/** One entry of the sandbox surface for autocomplete, e.g. `pm.environment.get`. */
export interface ScriptMember {
  path: string
  signature: string
  description: string
  kind: ScriptMemberKind
}

export interface ScriptSnippet {
  id: string
  title: string
  phase: SnippetPhase
  text: string
}

export interface ScriptSyntaxError {
  message: string
  line: number
  column: number
}

const KINDS: ReadonlySet<string> = new Set<ScriptMemberKind>(['object', 'function', 'property'])
const PHASES: ReadonlySet<string> = new Set<SnippetPhase>(['any', 'preRequest', 'postResponse'])

export function toScriptMember(dto: ScriptMemberDto): ScriptMember {
  return { path: dto.path, signature: dto.signature, description: dto.description, kind: KINDS.has(dto.kind) ? (dto.kind as ScriptMemberKind) : 'property' }
}

export function toScriptSnippet(dto: ScriptSnippetDto): ScriptSnippet {
  return { id: dto.id, title: dto.title, phase: PHASES.has(dto.phase) ? (dto.phase as SnippetPhase) : 'any', text: dto.text }
}

export function toSyntaxErrors(dto: ScriptCheckDto): ScriptSyntaxError[] {
  return (dto.errors ?? []).map(({ message, line, column }) => ({ message, line, column }))
}
