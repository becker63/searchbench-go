import { defineConfig } from "vitepress";

export default defineConfig({
  title: "SearchBench",
  description: "Release evidence for agent and tool candidates",
  base: "/searchbench-go/",
  // Plain Markdown links to repo-root files (AGENTS.md, ../README) stay valid on GitHub.
  ignoreDeadLinks: true,
  srcExclude: [
    "**/archive/blog-diff-candidates/full-diffs/**",
    "**/archive/blog-diff-candidates/_embed_patches.py",
  ],
  themeConfig: {
    nav: [
      { text: "Start", link: "/start-here" },
      { text: "Concepts", link: "/concepts" },
      { text: "Architecture", link: "/architecture" },
      { text: "Development", link: "/development" },
    ],
    sidebar: [
      {
        text: "Guide",
        items: [
          { text: "Start here", link: "/start-here" },
          { text: "Concepts", link: "/concepts" },
          { text: "Architecture", link: "/architecture" },
          { text: "Development", link: "/development" },
          { text: "Workspace seeds", link: "/workspace-seeds" },
          { text: "Docs index", link: "/README" },
        ],
      },
      {
        text: "Reference",
        items: [
          { text: "Package boundaries", link: "/reference/package-boundaries" },
          { text: "Architecture (full)", link: "/reference/architecture-full" },
          { text: "Integration shape", link: "/reference/integration-shape" },
          { text: "Build system", link: "/reference/build-system" },
          { text: "Pkl round manifests", link: "/reference/pkl-round-manifests" },
          { text: "Pkl scoring", link: "/reference/pkl-scoring-interface" },
          { text: "Optimizer policy validation", link: "/reference/optimizer-policy-validation" },
          { text: "LangSmith", link: "/reference/langsmith-integration" },
        ],
      },
      {
        text: "Archive",
        items: [{ text: "Archive index", link: "/archive/README" }],
      },
    ],
    socialLinks: [
      { icon: "github", link: "https://github.com/becker63/searchbench-go" },
    ],
  },
});
