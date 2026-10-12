import { describe, expect, it } from 'vitest'
import type { ApiOperation } from '@/platform/mappers/documentation'
import { sectionOperations } from './operation-groups'

function operation(requestId: string, tag: string): ApiOperation {
  return {
    requestId,
    tag,
    method: 'GET',
    path: `/${requestId}`,
    server: '',
    summary: requestId,
    description: '',
    parameters: [],
    body: null,
    responses: [],
    security: [],
    variant: false,
  }
}

const operations = [operation('list', 'Products'), operation('ping', ''), operation('create', 'Products / Admin'), operation('login', 'Auth'), operation('stray', 'Stray')]

function shape(sections: ReturnType<typeof sectionOperations>) {
  return sections.map((section) => ({ group: section.group, tags: section.tags.map((group) => [group.tag, group.operations.map((op) => op.requestId)]) }))
}

describe('sectionOperations', () => {
  it('keeps plain folder groups in one unnamed section when the document has no x-tagGroups', () => {
    expect(shape(sectionOperations(operations, []))).toEqual([
      {
        group: '',
        tags: [
          ['Products', ['list']],
          ['', ['ping']],
          ['Products / Admin', ['create']],
          ['Auth', ['login']],
          ['Stray', ['stray']],
        ],
      },
    ])
  })

  it('puts each x-tagGroups group in its own section, in its tag order, after the requests no group holds', () => {
    const groups = [
      { name: 'Products', tags: ['Products / Admin', 'Products', 'Empty'] },
      { name: 'Auth', tags: ['Auth'] },
      { name: 'Nothing', tags: ['None'] },
    ]
    expect(shape(sectionOperations(operations, groups))).toEqual([
      {
        group: '',
        tags: [
          ['', ['ping']],
          ['Stray', ['stray']],
        ],
      },
      {
        group: 'Products',
        tags: [
          ['Products / Admin', ['create']],
          ['Products', ['list']],
        ],
      },
      { group: 'Auth', tags: [['Auth', ['login']]] },
    ])
  })

  it('has no unnamed section when every operation is in a group', () => {
    const sections = sectionOperations([operation('login', 'Auth')], [{ name: 'Auth', tags: ['Auth'] }])
    expect(shape(sections)).toEqual([{ group: 'Auth', tags: [['Auth', ['login']]] }])
  })
})
