import { defineConfig } from "vitepress";

export default defineConfig({
  title: "SearchBench",
  description: "Release evidence for agent and tool candidates",
  base: "/searchbench-go/",
  ignoreDeadLinks: true,
  srcExclude: [
    "**/README.md",
    "architecture/**",
    "engineering/**",
  ],
  themeConfig: {
    nav: [
      { text: "Home", link: "/" },
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
        ],
      },
      {
        text: "Reference",
        items: [
          { text: "Package boundaries", link: "/reference/package-boundaries" },
          { text: "Pkl rounds", link: "/reference/pkl-rounds" },
          { text: "Pkl objectives", link: "/reference/pkl-objectives" },
          { text: "Bundles", link: "/reference/bundles" },
          { text: "Optimizer policy validation", link: "/reference/optimizer-policy-validation" },
        ],
      },
      {
        text: "Research",
        items: [
          { text: "Agent interface research", link: "/research/agent-interface-research" },
          { text: "Buck / work-graph (research)", link: "/research/bxl-meta-harness" },
        ],
      },
    ],
    socialLinks: [
      { icon: "github", link: "https://github.com/becker63/searchbench-go" },
    ],
  },
});
