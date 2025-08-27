// Go Coffee Theme Configuration
// Centralized theme management for consistent branding

export const theme = {
  // Brand Colors
  brand: {
    primary: '#d4a574', // coffee-500
    secondary: '#c4956c', // coffee-600
    accent: '#ffd700', // brand-gold
    amber: '#ffbf00', // brand-amber
    copper: '#b87333', // brand-copper
    bronze: '#cd7f32', // brand-bronze
  },

  // Coffee Color Palette
  coffee: {
    50: '#fefcfb',
    100: '#fdf6f0',
    200: '#f9e6d3',
    300: '#f4d0a7',
    400: '#e8b86d',
    500: '#d4a574',
    600: '#c4956c',
    700: '#a67c5a',
    800: '#8b6914',
    900: '#6b4423',
    950: '#3d2512',
  },

  // Crypto Colors
  crypto: {
    bitcoin: '#f7931a',
    ethereum: '#627eea',
    usdc: '#2775ca',
    usdt: '#26a17b',
    solana: '#9945ff',
    cardano: '#0033ad',
    polygon: '#8247e5',
    binance: '#f3ba2f',
  },

  // Status Colors
  status: {
    success: '#10b981',
    warning: '#f59e0b',
    error: '#ef4444',
    info: '#3b82f6',
  },

  // Typography
  typography: {
    fontFamily: {
      display: ['Poppins', 'Inter', 'system-ui', 'sans-serif'],
      body: ['Inter', 'system-ui', 'sans-serif'],
      mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
    },
    fontSize: {
      xs: '0.75rem',
      sm: '0.875rem',
      base: '1rem',
      lg: '1.125rem',
      xl: '1.25rem',
      '2xl': '1.5rem',
      '3xl': '1.875rem',
      '4xl': '2.25rem',
      '5xl': '3rem',
      '6xl': '3.75rem',
      '7xl': '4.5rem',
      '8xl': '6rem',
      '9xl': '8rem',
    },
    fontWeight: {
      light: '300',
      normal: '400',
      medium: '500',
      semibold: '600',
      bold: '700',
      extrabold: '800',
    },
  },

  // Spacing
  spacing: {
    xs: '0.25rem',
    sm: '0.5rem',
    md: '1rem',
    lg: '1.5rem',
    xl: '2rem',
    '2xl': '3rem',
    '3xl': '4rem',
    '4xl': '6rem',
    '5xl': '8rem',
  },

  // Border Radius
  borderRadius: {
    sm: '0.375rem',
    md: '0.5rem',
    lg: '0.75rem',
    xl: '1rem',
    '2xl': '1.5rem',
    full: '9999px',
  },

  // Shadows
  shadows: {
    sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
    md: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
    lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1)',
    xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1)',
    '2xl': '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
    glow: '0 0 20px rgba(212, 165, 116, 0.3)',
    'glow-lg': '0 0 40px rgba(212, 165, 116, 0.4)',
    coffee: '0 4px 20px rgba(139, 69, 19, 0.3)',
    crypto: '0 4px 20px rgba(102, 126, 234, 0.3)',
  },

  // Animations
  animations: {
    duration: {
      fast: '150ms',
      normal: '300ms',
      slow: '500ms',
    },
    easing: {
      default: 'cubic-bezier(0.4, 0, 0.2, 1)',
      in: 'cubic-bezier(0.4, 0, 1, 1)',
      out: 'cubic-bezier(0, 0, 0.2, 1)',
      inOut: 'cubic-bezier(0.4, 0, 0.2, 1)',
    },
  },

  // Breakpoints
  breakpoints: {
    sm: '640px',
    md: '768px',
    lg: '1024px',
    xl: '1280px',
    '2xl': '1536px',
  },

  // Component Variants
  components: {
    button: {
      sizes: {
        sm: 'h-9 px-3 text-sm',
        md: 'h-10 px-4 text-sm',
        lg: 'h-11 px-8 text-base',
        xl: 'h-12 px-10 text-lg',
      },
      variants: {
        primary: 'bg-gradient-to-r from-coffee-500 to-coffee-600 text-white hover:from-coffee-600 hover:to-coffee-700',
        secondary: 'bg-gradient-to-r from-slate-700 to-slate-800 text-white hover:from-slate-600 hover:to-slate-700',
        ghost: 'bg-transparent hover:bg-white/10 text-white border border-white/20',
        coffee: 'bg-gradient-to-r from-coffee-500 to-brand-amber text-white shadow-coffee',
        crypto: 'bg-gradient-to-r from-crypto-bitcoin to-crypto-ethereum text-white crypto-glow',
        success: 'bg-gradient-to-r from-status-success to-green-600 text-white',
        warning: 'bg-gradient-to-r from-status-warning to-yellow-600 text-white',
        error: 'bg-gradient-to-r from-status-error to-red-600 text-white',
      },
    },
    card: {
      variants: {
        default: 'bg-card border border-border rounded-xl shadow-sm',
        glass: 'glass-card',
        feature: 'feature-card',
        stats: 'stats-card',
        metric: 'metric-card',
        coffee: 'bg-gradient-to-br from-coffee-900/20 to-coffee-800/20 border-coffee-700/30',
        crypto: 'crypto-card',
        glow: 'shadow-glow border-coffee-500/30',
      },
    },
  },
}

// Theme utilities
export const getThemeColor = (path: string) => {
  const keys = path.split('.')
  let value: any = theme
  
  for (const key of keys) {
    value = value?.[key]
    if (value === undefined) return null
  }
  
  return value
}

// CSS custom properties generator
export const generateCSSVariables = () => {
  const cssVars: Record<string, string> = {}
  
  // Brand colors
  Object.entries(theme.brand).forEach(([key, value]) => {
    cssVars[`--brand-${key}`] = value
  })
  
  // Coffee colors
  Object.entries(theme.coffee).forEach(([key, value]) => {
    cssVars[`--coffee-${key}`] = value
  })
  
  // Crypto colors
  Object.entries(theme.crypto).forEach(([key, value]) => {
    cssVars[`--crypto-${key}`] = value
  })
  
  // Status colors
  Object.entries(theme.status).forEach(([key, value]) => {
    cssVars[`--status-${key}`] = value
  })
  
  return cssVars
}

// Responsive utilities
export const responsive = {
  mobile: `@media (max-width: ${theme.breakpoints.sm})`,
  tablet: `@media (min-width: ${theme.breakpoints.sm}) and (max-width: ${theme.breakpoints.lg})`,
  desktop: `@media (min-width: ${theme.breakpoints.lg})`,
  
  // Mobile-first approach
  sm: `@media (min-width: ${theme.breakpoints.sm})`,
  md: `@media (min-width: ${theme.breakpoints.md})`,
  lg: `@media (min-width: ${theme.breakpoints.lg})`,
  xl: `@media (min-width: ${theme.breakpoints.xl})`,
  '2xl': `@media (min-width: ${theme.breakpoints['2xl']})`,
}

// Animation presets
export const animations = {
  fadeIn: {
    initial: { opacity: 0 },
    animate: { opacity: 1 },
    exit: { opacity: 0 },
  },
  slideUp: {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    exit: { opacity: 0, y: -20 },
  },
  slideRight: {
    initial: { opacity: 0, x: -20 },
    animate: { opacity: 1, x: 0 },
    exit: { opacity: 0, x: 20 },
  },
  scale: {
    initial: { opacity: 0, scale: 0.9 },
    animate: { opacity: 1, scale: 1 },
    exit: { opacity: 0, scale: 0.9 },
  },
  bounce: {
    animate: {
      y: [0, -10, 0],
      transition: {
        duration: 2,
        repeat: Infinity,
        ease: 'easeInOut',
      },
    },
  },
}

export default theme
