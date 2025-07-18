/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    appDir: true,
  },
  transpilePackages: ['@developer-dao/shared'],
  images: {
    domains: ['localhost', 'api.developer-dao.com'],
  },
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
    NEXT_PUBLIC_MARKETPLACE_API_URL: process.env.NEXT_PUBLIC_MARKETPLACE_API_URL || 'http://localhost:8081',
    NEXT_PUBLIC_METRICS_API_URL: process.env.NEXT_PUBLIC_METRICS_API_URL || 'http://localhost:8082',
  },
}

module.exports = nextConfig
