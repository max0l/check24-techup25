/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  modularizeImports: {
    '@mui/icons-material': {
      transform: '@mui/icons-material/{{ member }}'
    },
    '@mui/material': {
      transform: '@mui/material/{{ member }}'
    }
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://api:8080/:path*'
      }
    ]
  }
}

module.exports = nextConfig
