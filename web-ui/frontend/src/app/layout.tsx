import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Go Coffee Epic UI - Web3 Coffee Ecosystem',
  description: 'Revolutionary Web3 ecosystem combining traditional coffee ordering with DeFi trading, AI automation, and multi-region infrastructure',
}

interface RootLayoutProps {
  children: React.ReactNode
}

export default function RootLayout({ children }: RootLayoutProps) {
  return (
    <html lang="en">
      <head>
        <meta name="theme-color" content="#d97706" />
        <link rel="icon" href="/favicon.ico" />
      </head>
      <body style={{
        margin: 0,
        padding: 0,
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", sans-serif',
        backgroundColor: '#0f172a',
        color: '#f8fafc',
        minHeight: '100vh'
      }}>
        {children}
      </body>
    </html>
  )
}
