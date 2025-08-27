import type { Metadata } from 'next'
import { RealTimeDataProvider } from '@/contexts/RealTimeDataContext'
import './globals.css'

export const metadata: Metadata = {
  title: 'Go Coffee - Web3 Coffee Ecosystem',
  description: 'Revolutionary Web3 ecosystem combining traditional coffee ordering with DeFi trading, AI automation, and enterprise-grade infrastructure. Experience the future of coffee commerce.',
  keywords: 'coffee, web3, defi, trading, cryptocurrency, AI, automation, blockchain, ethereum, solana',
  authors: [{ name: 'Go Coffee Team' }],
  creator: 'Go Coffee',
  publisher: 'Go Coffee',
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  metadataBase: new URL('https://gocoffee.io'),
  openGraph: {
    title: 'Go Coffee - Web3 Coffee Ecosystem',
    description: 'Revolutionary Web3 ecosystem combining traditional coffee ordering with DeFi trading and AI automation.',
    url: 'https://gocoffee.io',
    siteName: 'Go Coffee',
    images: [
      {
        url: '/og-image.png',
        width: 1200,
        height: 630,
        alt: 'Go Coffee - Web3 Coffee Ecosystem',
      },
    ],
    locale: 'en_US',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Go Coffee - Web3 Coffee Ecosystem',
    description: 'Revolutionary Web3 ecosystem combining traditional coffee ordering with DeFi trading and AI automation.',
    images: ['/og-image.png'],
    creator: '@GoCoffeeWeb3',
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
}

interface RootLayoutProps {
  children: React.ReactNode
}

export default function RootLayout({ children }: RootLayoutProps) {
  return (
    <html lang="en">
      <head>
        <meta name="theme-color" content="#d4a574" />
        <meta name="apple-mobile-web-app-capable" content="yes" />
        <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
        <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no" />
        <link rel="icon" href="/favicon.ico" />
        <link rel="apple-touch-icon" href="/apple-touch-icon.png" />
        <link rel="manifest" href="/manifest.json" />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=Poppins:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet" />
      </head>
      <body className="font-body bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-white min-h-screen antialiased">
        <RealTimeDataProvider>
          {children}
        </RealTimeDataProvider>
      </body>
    </html>
  )
}
