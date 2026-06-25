import { defineConfig } from 'orval'

export default defineConfig({
  dotsBeacon: {
    input: {
      // Reads the live spec. If the API isn't running during generation,
      // save the spec to ./openapi.json and point `target` at it instead.
      target: 'http://localhost:8080/openapi.json',
    },
    output: {
      mode: 'tags-split',
      target: './src/lib/api/generated',
      schemas: './src/lib/api/generated/model',
      client: 'react-query',
      httpClient: 'axios',
      clean: true,
      override: {
        mutator: {
          path: './src/lib/api/axios-instance.ts',
          name: 'customInstance',
        },
      },
    },
  },
})
