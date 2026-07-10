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
