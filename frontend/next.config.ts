import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'export', // Enable static HTML export
  images: {
    unoptimized: true, // Required for static export (no Image Optimization API)
  },
  trailingSlash: true, // S3 compatibility for client-side routing
  eslint: {
    ignoreDuringBuilds: true, // Ignore lint errors during build
  },
  typescript: {
    ignoreBuildErrors: false, // Keep TypeScript checks
  },
};

export default nextConfig;
