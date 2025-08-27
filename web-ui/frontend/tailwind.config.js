/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: ["class"],
  content: [
    './pages/**/*.{ts,tsx}',
    './components/**/*.{ts,tsx}',
    './app/**/*.{ts,tsx}',
    './src/**/*.{ts,tsx}',
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        // Enhanced Coffee Color Palette
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
        // Enhanced Crypto Colors
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
        // Modern Accent Colors
        brand: {
          gold: '#ffd700',
          copper: '#b87333',
          bronze: '#cd7f32',
          amber: '#ffbf00',
        },
        // Status Colors
        status: {
          success: '#10b981',
          warning: '#f59e0b',
          error: '#ef4444',
          info: '#3b82f6',
        },
        // Gradient Colors
        gradient: {
          from: '#1e293b',
          via: '#334155',
          to: '#475569',
        }
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      keyframes: {
        "accordion-down": {
          from: { height: 0 },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: 0 },
        },
        "fade-in": {
          from: { opacity: 0, transform: "translateY(10px)" },
          to: { opacity: 1, transform: "translateY(0)" },
        },
        "fade-in-up": {
          from: { opacity: 0, transform: "translateY(30px)" },
          to: { opacity: 1, transform: "translateY(0)" },
        },
        "slide-in": {
          from: { transform: "translateX(-100%)" },
          to: { transform: "translateX(0)" },
        },
        "slide-in-right": {
          from: { transform: "translateX(100%)" },
          to: { transform: "translateX(0)" },
        },
        "pulse-glow": {
          "0%, 100%": { boxShadow: "0 0 5px rgba(191, 160, 148, 0.5)" },
          "50%": { boxShadow: "0 0 20px rgba(191, 160, 148, 0.8)" },
        },
        "coffee-steam": {
          "0%": { transform: "translateY(0) rotate(0deg)", opacity: 1 },
          "50%": { transform: "translateY(-10px) rotate(5deg)", opacity: 0.8 },
          "100%": { transform: "translateY(-20px) rotate(-5deg)", opacity: 0 },
        },
        "float": {
          "0%, 100%": { transform: "translateY(0px)" },
          "50%": { transform: "translateY(-10px)" },
        },
        "bounce-gentle": {
          "0%, 100%": { transform: "translateY(0)" },
          "50%": { transform: "translateY(-5px)" },
        },
        "shimmer": {
          "0%": { backgroundPosition: "-200px 0" },
          "100%": { backgroundPosition: "calc(200px + 100%) 0" },
        },
        "gradient-shift": {
          "0%, 100%": { backgroundPosition: "0% 50%" },
          "50%": { backgroundPosition: "100% 50%" },
        },
        "scale-in": {
          from: { transform: "scale(0.9)", opacity: 0 },
          to: { transform: "scale(1)", opacity: 1 },
        },
        "wiggle": {
          "0%, 100%": { transform: "rotate(-3deg)" },
          "50%": { transform: "rotate(3deg)" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
        "fade-in": "fade-in 0.5s ease-out",
        "fade-in-up": "fade-in-up 0.6s ease-out",
        "slide-in": "slide-in 0.3s ease-out",
        "slide-in-right": "slide-in-right 0.3s ease-out",
        "pulse-glow": "pulse-glow 2s ease-in-out infinite",
        "coffee-steam": "coffee-steam 3s ease-in-out infinite",
        "float": "float 3s ease-in-out infinite",
        "bounce-gentle": "bounce-gentle 2s ease-in-out infinite",
        "shimmer": "shimmer 2s linear infinite",
        "gradient-shift": "gradient-shift 3s ease-in-out infinite",
        "scale-in": "scale-in 0.3s ease-out",
        "wiggle": "wiggle 1s ease-in-out infinite",
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
        display: ['Poppins', 'Inter', 'system-ui', 'sans-serif'],
        body: ['Inter', 'system-ui', 'sans-serif'],
      },
      fontSize: {
        'xs': ['0.75rem', { lineHeight: '1rem' }],
        'sm': ['0.875rem', { lineHeight: '1.25rem' }],
        'base': ['1rem', { lineHeight: '1.5rem' }],
        'lg': ['1.125rem', { lineHeight: '1.75rem' }],
        'xl': ['1.25rem', { lineHeight: '1.75rem' }],
        '2xl': ['1.5rem', { lineHeight: '2rem' }],
        '3xl': ['1.875rem', { lineHeight: '2.25rem' }],
        '4xl': ['2.25rem', { lineHeight: '2.5rem' }],
        '5xl': ['3rem', { lineHeight: '1' }],
        '6xl': ['3.75rem', { lineHeight: '1' }],
        '7xl': ['4.5rem', { lineHeight: '1' }],
        '8xl': ['6rem', { lineHeight: '1' }],
        '9xl': ['8rem', { lineHeight: '1' }],
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-conic': 'conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))',
        'gradient-coffee': 'linear-gradient(135deg, #8b4513 0%, #d2691e 50%, #cd853f 100%)',
        'gradient-crypto': 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        'gradient-gold': 'linear-gradient(135deg, #ffd700 0%, #ffb347 100%)',
        'coffee-pattern': "url('/images/coffee-pattern.svg')",
        'noise': "url('data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.65' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E')",
      },
      boxShadow: {
        'glow': '0 0 20px rgba(245, 158, 11, 0.3)',
        'glow-lg': '0 0 40px rgba(245, 158, 11, 0.4)',
        'coffee': '0 4px 20px rgba(139, 69, 19, 0.3)',
        'crypto': '0 4px 20px rgba(102, 126, 234, 0.3)',
        'inner-glow': 'inset 0 0 20px rgba(245, 158, 11, 0.1)',
      },
      backdropBlur: {
        'xs': '2px',
      },
    },
  },
  plugins: [
    require("@tailwindcss/forms"),
    require("@tailwindcss/typography"),
  ],
}
