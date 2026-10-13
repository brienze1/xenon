import type {
  ApiDocument as ApiDocumentDto,
  ApiOperation as ApiOperationDto,
  Documentation as DocumentationDto,
  DocumentationApplied as DocumentationAppliedDto,
  DocumentationItem as DocumentationItemDto,
  DocumentationPlan as DocumentationPlanDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export type ApiParameterLocation = 'path' | 'query' | 'header'

export interface ApiParameter {
  name: string
  in: ApiParameterLocation
  description: string
  example: string
  required: boolean
  /** A `{{variable}}` path segment rather than a `:path` variable. */
  variable: boolean
}

export interface ApiField {
  name: string
  description: string
  example: string
  binary: boolean
}

export interface ApiBody {
  mediaType: string
  example: string
  fields: ApiField[]
  binary: boolean
}

export interface ApiExample {
  id: string
  name: string
  mediaType: string
  value: string
  binary: boolean
}

export interface ApiResponse {
  code: number
  description: string
  examples: ApiExample[]
}

export interface ApiOperation {
  /** The request the operation documents; empty for an operation the document adds. */
  requestId: string
  tag: string
  method: string
  path: string
  server: string
  summary: string
  description: string
  parameters: ApiParameter[]
  body: ApiBody | null
  responses: ApiResponse[]
  security: string[]
  /** A second request on the same method and path (listed under x-xenon-variants). */
  variant: boolean
}

export interface ApiSecurityScheme {
  name: string
  type: string
  scheme: string
  in: string
  paramName: string
  flow: string
  description: string
  scopes: string[]
}

export interface ApiTag {
  name: string
  description: string
}

/** A section of folder tags (OpenAPI's x-tagGroups): a top-level folder with its own tag and its nested folders' tags. */
export interface ApiTagGroup {
  name: string
  tags: string[]
}

export interface ApiDocument {
  title: string
  description: string
  version: string
  servers: string[]
  tags: ApiTag[]
  tagGroups: ApiTagGroup[]
  operations: ApiOperation[]
  securitySchemes: ApiSecurityScheme[]
}

export interface Documentation {
  yaml: string
  document: ApiDocument
}

export interface DocumentationItem {
  id: string
  name: string
  folder: string
  method: string
  path: string
  revision: string
}

/** What saving an edited document changes: the collection or folder and the requests it updates, creates and deletes. */
export interface DocumentationPlan {
  name: string
  collectionChanged: boolean
  updated: DocumentationItem[]
  created: DocumentationItem[]
  deleted: DocumentationItem[]
}

export interface DocumentationApplied extends DocumentationPlan {
  documentation: Documentation
}

function toLocation(value: string): ApiParameterLocation {
  return value === 'path' || value === 'header' ? value : 'query'
}

function toOperation(dto: ApiOperationDto): ApiOperation {
  return {
    requestId: dto.request_id,
    tag: dto.tag,
    method: dto.method,
    path: dto.path,
    server: dto.server,
    summary: dto.summary,
    description: dto.description,
    parameters: (dto.parameters ?? []).map((parameter) => ({ ...parameter, in: toLocation(parameter.in) })),
    body: dto.request_body
      ? { mediaType: dto.request_body.media_type, example: dto.request_body.example, fields: dto.request_body.fields ?? [], binary: dto.request_body.binary }
      : null,
    responses: (dto.responses ?? []).map((response) => ({
      code: response.code,
      description: response.description,
      examples: (response.examples ?? []).map((example) => ({ id: example.id, name: example.name, mediaType: example.media_type, value: example.value, binary: example.binary })),
    })),
    security: dto.security ?? [],
    variant: dto.variant,
  }
}

export function toApiDocument(dto: ApiDocumentDto): ApiDocument {
  return {
    title: dto.title,
    description: dto.description,
    version: dto.version,
    servers: dto.servers ?? [],
    tags: dto.tags ?? [],
    tagGroups: (dto.tag_groups ?? []).map((group) => ({ name: group.name, tags: group.tags ?? [] })),
    operations: (dto.operations ?? []).map(toOperation),
    securitySchemes: (dto.security_schemes ?? []).map((scheme) => ({
      name: scheme.name,
      type: scheme.type,
      scheme: scheme.scheme,
      in: scheme.in,
      paramName: scheme.param_name,
      flow: scheme.flow,
      description: scheme.description,
      scopes: scheme.scopes ?? [],
    })),
  }
}

export function toDocumentation(dto: DocumentationDto): Documentation {
  return { yaml: dto.yaml, document: toApiDocument(dto.document) }
}

function toItems(items: DocumentationItemDto[] | null): DocumentationItem[] {
  return (items ?? []).map((item) => ({ ...item }))
}

export function toDocumentationPlan(dto: DocumentationPlanDto): DocumentationPlan {
  return {
    name: dto.name,
    collectionChanged: dto.collection_changed,
    updated: toItems(dto.updated),
    created: toItems(dto.created),
    deleted: toItems(dto.deleted),
  }
}

export function toDocumentationApplied(dto: DocumentationAppliedDto): DocumentationApplied {
  return { ...toDocumentationPlan(dto), documentation: toDocumentation(dto) }
}
