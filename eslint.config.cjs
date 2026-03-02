const tseslint = require("@typescript-eslint/eslint-plugin");
const tsparser = require("@typescript-eslint/parser");

module.exports = [
  {
    ignores: ["**/*.d.ts", "lib/*.js", "bin/*.js", "test/*.js", "node_modules/", "cdk.out/"],
  },
  {
    files: ["lib/**/*.ts", "bin/**/*.ts", "test/**/*.ts"],
    languageOptions: {
      parser: tsparser,
    },
    plugins: {
      "@typescript-eslint": tseslint,
    },
    rules: {
      ...tseslint.configs.recommended.rules,
    },
  },
];
