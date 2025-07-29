/** @type {import('next').NextConfig} */
const nextConfig = {
  // Enable static export for Cloudflare Pages
  output: 'export',
  trailingSlash: true,
  
  // Image optimization - disabled for static export
  images: {
    unoptimized: true,
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
