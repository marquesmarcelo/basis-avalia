import type { KnipConfig } from "knip"

const config: KnipConfig = {
  entry: ["src/app/**/{page,layout,loading,error,not-found,route,template}.{ts,tsx}"],
  project: ["src/**/*.{ts,tsx}"],
  next: {
    entry: ["src/app/**/{page,layout,loading,error,route}.{ts,tsx}", "src/proxy.ts"],
  },
}

export default config
