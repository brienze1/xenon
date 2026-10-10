import * as ExampleService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/exampleservice'
import { AppError, toAppError } from '../errors'
import { toExample, toExampleEditDto, type Example } from '../mappers/example'

const EMPTY_RESPONSE = 'UNK-00204'

function required<T>(value: T | null): T {
  if (value === null) {
    throw new AppError(EMPTY_RESPONSE, 'The backend returned no data')
  }
  return value
}

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

export function getExample(localId: string, requestId: string, exampleId: string): Promise<Example> {
  return call(async () => toExample(required(await ExampleService.Get({ local_id: localId, request_id: requestId, example_id: exampleId }))))
}

export function createExample(localId: string, requestId: string, name: string): Promise<Example> {
  return call(async () => toExample(required(await ExampleService.Create({ local_id: localId, request_id: requestId, name }))))
}

export function createExampleFromResponse(localId: string, requestId: string, tabId: string, name: string): Promise<Example> {
  return call(async () =>
    toExample(required(await ExampleService.CreateFromResponse({ local_id: localId, request_id: requestId, tab_id: tabId, name }))),
  )
}

export function saveExample(localId: string, requestId: string, example: Example, baseRevision: string, force: boolean): Promise<Example> {
  return call(async () =>
    toExample(
      required(
        await ExampleService.Save({ local_id: localId, request_id: requestId, base_revision: baseRevision, force, example: toExampleEditDto(example) }),
      ),
    ),
  )
}

export function renameExample(localId: string, requestId: string, exampleId: string, name: string): Promise<Example> {
  return call(async () =>
    toExample(required(await ExampleService.Rename({ local_id: localId, request_id: requestId, example_id: exampleId, name }))),
  )
}

export function duplicateExample(localId: string, requestId: string, exampleId: string): Promise<Example> {
  return call(async () => toExample(required(await ExampleService.Duplicate({ local_id: localId, request_id: requestId, example_id: exampleId }))))
}

export function deleteExample(localId: string, requestId: string, exampleId: string): Promise<void> {
  return call(() => ExampleService.Delete({ local_id: localId, request_id: requestId, example_id: exampleId }))
}
