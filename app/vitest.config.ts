import { defineConfig } from "vitest/config";

export default defineConfig({
  esbuild: {
    treeShaking: true,
    target: "es2023", // this is needed for the newest typescript features
  },
  build: {
    rollupOptions: {
      treeshake: true,
    },
  },
  test: {
    pool: "forks",
    include: ["**/*.vitest.ts"],
    testTimeout: 60000,
  },
  plugins: [],
});
