const config = {
  branches: [
    { name: "main" },
    { name: "1.x", range: "1.x", channel: "1.x" },
  ],
  verifyConditions: [
    // Verifies the conditions for the plugins used below
    // For example verifying a GITHUB_TOKEN environment variable has been provided
    "@semantic-release/changelog",
    "@semantic-release/git",
    "@semantic-release/github",
  ],
  plugins: [
    [
      "@semantic-release/commit-analyzer",
      { preset: "conventionalcommits" },
    ],
    [
      "@semantic-release/release-notes-generator",
      { preset: "conventionalcommits" },
    ],
    ["@semantic-release/github"],
    [
      "@semantic-release/exec",
      {
        prepareCmd:
          "chmod +x ./scripts/release.sh && ./scripts/release.sh ${nextRelease.version}",
      },
    ],
    [
      "@semantic-release/git",
      {
        assets: [
          "go.mod",
          ".coverage/coverage.svg",
          ".coverage/main.breakdown",
        ],
      },
    ],
  ],
};

module.exports = config;