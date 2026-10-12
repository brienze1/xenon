import type { ApiDocument, Documentation, DocumentationPlan } from '@/platform/mappers/documentation'

export const generatedDocument: ApiDocument = {
  title: 'Billing',
  description: 'Invoices and payments.',
  version: '1.0.0',
  servers: ['{{host}}'],
  tags: [{ name: 'Invoices', description: '' }],
  tagGroups: [],
  operations: [
    {
      requestId: 'r1',
      tag: 'Invoices',
      method: 'GET',
      path: '/invoices',
      server: '{{host}}',
      summary: 'List invoices',
      description: 'Lists **open** invoices.',
      parameters: [{ name: 'status', in: 'query', description: 'Invoice status', example: 'open', required: false, variable: false }],
      body: null,
      responses: [{ code: 200, description: 'OK', examples: [{ id: 'e1', name: 'A page', mediaType: 'application/json', value: '[]', binary: false }] }],
      security: ['bearerAuth'],
      variant: false,
    },
    {
      requestId: 'r2',
      tag: '',
      method: 'GET',
      path: '/ping',
      server: '{{host}}',
      summary: 'Ping',
      description: '',
      parameters: [],
      body: null,
      responses: [],
      security: [],
      variant: false,
    },
  ],
  securitySchemes: [{ name: 'bearerAuth', type: 'http', scheme: 'bearer', in: '', paramName: '', flow: '', description: '', scopes: [] }],
}

export const generated: Documentation = { yaml: 'openapi: 3.1.0\ninfo:\n  title: Billing\n', document: generatedDocument }

export const editedYaml = 'openapi: 3.1.0\ninfo:\n  title: Billing v2\n'

export const plan: DocumentationPlan = {
  name: 'Billing',
  collectionChanged: true,
  updated: [{ id: 'r1', name: 'List invoices', folder: '', method: 'GET', path: '', revision: '' }],
  created: [{ id: 'n1', name: 'Archive invoice', folder: 'Invoices / Archive', method: 'DELETE', path: '/invoices/{id}', revision: '' }],
  deleted: [{ id: 'r2', name: 'Ping', folder: '', method: 'GET', path: '', revision: '' }],
}

export const emptyPlan: DocumentationPlan = { name: 'Billing', collectionChanged: false, updated: [], created: [], deleted: [] }
