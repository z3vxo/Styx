import { defineConfig } from 'vite';

export default defineConfig({
  server: {
    proxy: {
      '/rest': {
        target: process.env.API_BASE_URL || 'http://127.0.0.1:8080',
        ws: true,
      },
    },
  },
});
