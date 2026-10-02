import type { Metadata } from "next";
export const metadata: Metadata = { title: "FlowPilot", description: "Self-hosted workflow automation" };
export default function Layout({ children }: {children: React.ReactNode}) { return <html lang="en"><body>{children}</body></html>; }
