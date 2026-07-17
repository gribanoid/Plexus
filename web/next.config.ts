import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  output: 'standalone',
  transpilePackages: ['@plexus/api', '@plexus/ui', '@plexus/features'],
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: '**.plexus.app' },
      { protocol: 'http', hostname: 'localhost' },
    ],
  },
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'X-Frame-Options', value: 'DENY' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
          {
            key: 'Permissions-Policy',
            value: 'camera=(), microphone=(), geolocation=()',
          },
          {
            key: 'Content-Security-Policy',
            value: [
              "default-src 'self'",
              "script-src 'self' 'unsafe-inline' 'unsafe-eval'",
              "style-src 'self' 'unsafe-inline'",
              "img-src 'self' data: https: blob:",
              "font-src 'self' data:",
              "connect-src 'self' http://localhost:* http://127.0.0.1:* https: ws: wss:",
              "frame-ancestors 'none'",
              "base-uri 'self'",
              "form-action 'self'",
            ].join('; '),
          },
        ],
      },
    ]
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'}/api/:path*`,
      },
    ]
  },
  async redirects() {
    return [
      {
        source: '/:orgSlug/:projectKey/board',
        destination: '/orgs/:orgSlug/:projectKey/board',
        permanent: false,
      },
      {
        source: '/:orgSlug/:projectKey/backlog',
        destination: '/orgs/:orgSlug/:projectKey/backlog',
        permanent: false,
      },
      {
        source: '/:orgSlug/:projectKey/settings',
        destination: '/orgs/:orgSlug/:projectKey/settings',
        permanent: false,
      },
      {
        source: '/:orgSlug/:projectKey/issues/:issueNumber',
        destination: '/orgs/:orgSlug/:projectKey/issues/:issueNumber',
        permanent: false,
      },
    ]
  },
}

export default nextConfig
