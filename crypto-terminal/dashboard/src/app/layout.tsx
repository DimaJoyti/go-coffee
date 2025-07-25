import type { Metadata } from 'next'
import { Inter, JetBrains_Mono } from 'next/font/google'
import './globals.css'
import { Providers } from '@/components/providers'
import { Toaster } from 'sonner'
import { cn } from '@/lib/utils'

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap',
})

const jetbrainsMono = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
  display: 'swap',
})

export const metadata: Metadata = {
  title: 'Epic Crypto Terminal | Professional Trading Dashboard',
  description: 'Advanced cryptocurrency trading terminal with real-time market data, portfolio management, and professional trading tools',
  keywords: ['crypto', 'trading', 'terminal', 'bitcoin', 'ethereum', 'defi', 'portfolio', 'charts', 'market data'],
  authors: [{ name: 'Epic Crypto Terminal' }],
  creator: 'Epic Crypto Terminal',
  publisher: 'Epic Crypto Terminal',
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  metadataBase: new URL('https://epic-crypto-terminal.com'),
  openGraph: {
    title: 'Epic Crypto Terminal',
    description: 'Professional cryptocurrency trading terminal',
    url: 'https://epic-crypto-terminal.com',
    siteName: 'Epic Crypto Terminal',
    images: [
      {
        url: '/og-image.png',
        width: 1200,
        height: 630,
        alt: 'Epic Crypto Terminal',
      },
    ],
    locale: 'en_US',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Epic Crypto Terminal',
    description: 'Professional cryptocurrency trading terminal',
    images: ['/og-image.png'],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  verification: {
    google: 'your-google-verification-code',
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" suppressHydrationWarning className="scroll-smooth">
      <head>
        <link rel="icon" href="/favicon.ico" sizes="any" />
        <link rel="icon" href="/icon.svg" type="image/svg+xml" />
        <link rel="apple-touch-icon" href="/apple-touch-icon.png" />
        <link rel="manifest" href="/manifest.json" />
        <meta name="theme-color" content="#000000" />
        <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no" />
      </head>
      <body
        className={cn(
          'min-h-screen bg-background font-sans antialiased overflow-x-hidden',
          'selection:bg-primary/20 selection:text-primary-foreground',
          inter.variable,
          jetbrainsMono.variable
        )}
      >
        <div className="relative flex min-h-screen flex-col">
          <div className="flex-1">
            <Providers>
              {children}
            </Providers>
          </div>
        </div>
        <Toaster
          position="top-right"
          expand={true}
          richColors
          closeButton
          toastOptions={{
            duration: 4000,
            className: 'group toast group-[.toaster]:bg-background group-[.toaster]:text-foreground group-[.toaster]:border-border group-[.toaster]:shadow-lg',
          }}
        />
      </body>
    </html>
  )
}
