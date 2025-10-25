import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  // output: 'standalone', // Disabled for development hot reloading
  
  // Environment variables that should be available to the client
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081',
  },
  
  // Allow warnings during build instead of failing
  eslint: {
    ignoreDuringBuilds: true,
  },
  typescript: {
    ignoreBuildErrors: false,
  },
  
  // Enable experimental features if needed
  experimental: {
    // Enable if you need server actions or other experimental features
  },
};

export default nextConfig;
