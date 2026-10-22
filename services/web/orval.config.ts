import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig, defineTransformer } from 'orval';

const directory = path.dirname(fileURLToPath(import.meta.url));
const input =
  process.env.CAFE_ORVAL_INPUT ??
  path.resolve(directory, '../api/adaptors/openapi/generated/api.openapi.json');
const output = process.env.CAFE_ORVAL_OUTPUT ?? path.resolve(directory, 'src/adaptors/generated/http');

function normaliseSchemas(value: unknown): void {
  if (Array.isArray(value)) {
    value.forEach(normaliseSchemas);
    return;
  }
  if (typeof value !== 'object' || value === null) return;
  const object = value as Record<string, unknown>;
  if (Array.isArray(object.required) && object.required.length === 0) delete object.required;
  Object.values(object).forEach(normaliseSchemas);
}

const browserContract = defineTransformer((document) => {
  // Preserve historical OpenAPI bytes while satisfying current JSON Schema tooling.
  normaliseSchemas(document);
  const methods = ['get', 'post', 'put', 'patch', 'delete', 'head', 'options'] as const;
  for (const pathItem of Object.values(document.paths ?? {})) {
    if (!pathItem || '$ref' in pathItem) continue;
    for (const method of methods) {
      const operation = pathItem[method];
      if (!operation?.parameters) continue;
      operation.parameters = operation.parameters.filter(
        (parameter) =>
          '$ref' in parameter || parameter.in !== 'header' || parameter.name.toLowerCase() !== 'origin',
      );
    }
  }
  return document;
});

export default defineConfig({
  cafe: {
    input: {
      target: input,
      override: {
        transformer: browserContract,
      },
    },
    output: {
      target: path.join(output, 'client.ts'),
      schemas: {
        path: path.join(output, 'model'),
        splitByTags: true,
      },
      operationManifest: {
        path: path.join(output, 'operations.ts'),
      },
      client: 'react-query',
      httpClient: 'fetch',
      mode: 'tags-split',
      headers: true,
      indexFiles: true,
      tagsSplitDeduplication: true,
      clean: true,
      mock: {
        indexMockFiles: true,
        fixtures: true,
        generators: [
          { type: 'msw', useExamples: true },
          { type: 'faker', useExamples: true },
        ],
      },
      override: {
        requestOptions: {
          cache: 'no-store',
        },
        mock: {
          exactOptional: true,
        },
        fetch: {
          forceSuccessResponse: true,
          includeHttpResponseReturnType: true,
        },
        query: {
          version: 5,
          signal: true,
          useInfinite: true,
          useInfiniteQueryParam: 'cursor',
          useGetQueryData: true,
          useSetQueryData: true,
        },
      },
    },
  },
});
