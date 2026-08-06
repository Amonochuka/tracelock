import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  // @ts-ignore - Bypass NextConfig type checking for allowedDevOrigins if missing
  allowedDevOrigins: ['192.168.89.64'],
};

export default nextConfig;
