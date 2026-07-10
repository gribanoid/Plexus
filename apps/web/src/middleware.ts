import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

const RESERVED = new Set(['login', 'register', 'orgs', 'api', '_next'])

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Legacy: /plexus -> /orgs/plexus (single-segment org slug only)
  const segments = pathname.split('/').filter(Boolean)
  if (segments.length === 1 && !RESERVED.has(segments[0])) {
    const url = request.nextUrl.clone()
    url.pathname = `/orgs/${segments[0]}`
    return NextResponse.redirect(url)
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
}
