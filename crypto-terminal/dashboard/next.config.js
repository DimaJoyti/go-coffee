/** @type {import('next').NextConfig} */
const nextConfig = {
  // Image optimization
  images: {
    domains: ['localhost'],
  },

  // API rewrites for backend integration
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: 'http://localhost:8090/api/v1/:path*',
      },
    ]
  },

  // Enhanced webpack configuration
  webpack: (config) => {
    // WebSocket externals
    config.externals.push({
      'utf-8-validate': 'commonjs utf-8-validate',
      'bufferutil': 'commonjs bufferutil',
    })

    return config
  },

  // Enable compression
  compress: true,

  // Use SWC minifier for better performance
  swcMinify: true,
}

module.exports = nextConfig
