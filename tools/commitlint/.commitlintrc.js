module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    // Treat as warning until Dependabot supports commitlint.
    // https://github.com/dependabot/dependabot-core/issues/2445
    "body-max-line-length": [1, "always", 100],
  }
};
