import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";

export default tseslint.config(
  {
    // Build output and generated files. api-contract.ts comes from
    // openapi-typescript, SourceRegistry.generated.ts and src/data-sources/
    // from sync-data-sources.mjs — linting them reports on a generator's
    // output, which nobody can act on without editing a file marked
    // "do not edit".
    ignores: [
      "dist",
      "src/shared/api-contract.ts",
      "src/components/sources/SourceRegistry.generated.ts",
      "src/data-sources",
    ],
  },

  // Application code: browser globals, React rules.
  {
    files: ["src/**/*.{ts,tsx}"],
    extends: [js.configs.recommended, tseslint.configs.recommended],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
    },
    plugins: {
      "react-hooks": reactHooks,
      "react-refresh": reactRefresh,
    },
    rules: {
      ...reactHooks.configs["recommended-latest"].rules,
      ...reactRefresh.configs.vite.rules,
      // Unused args are often there to document a signature; the leading
      // underscore is the opt-out.
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
    },
  },

  // Build and sync scripts: Node globals, no React, and they are allowed to
  // log — that output is the point of running them.
  {
    files: ["scripts/**/*.mjs"],
    extends: [js.configs.recommended],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.node,
    },
    rules: {
      // Same underscore opt-out as application code: test stubs keep the
      // signature they are standing in for, even when unused.
      "no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
    },
  },
);
