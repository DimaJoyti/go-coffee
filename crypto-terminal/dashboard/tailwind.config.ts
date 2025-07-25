import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  darkMode: ["class"],
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
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))',
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))',
        },
        popover: {
          DEFAULT: 'hsl(var(--popover))',
          foreground: 'hsl(var(--popover-foreground))',
        },
        card: {
          DEFAULT: 'hsl(var(--card))',
          foreground: 'hsl(var(--card-foreground))',
        },
        // Trading specific colors
        bull: "hsl(var(--bull))",
        bear: "hsl(var(--bear))",
        neutral: "hsl(var(--neutral))",
        profit: "hsl(var(--profit))",
        loss: "hsl(var(--loss))",
        warning: "hsl(var(--warning))",
        info: "hsl(var(--info))",
        coffee: {
          50: '#fdf7f0',
          100: '#f9e8d3',
          200: '#f2d0a7',
          300: '#e8b375',
          400: '#dc9441',
          500: '#d4791f',
          600: '#c66315',
          700: '#a44d14',
          800: '#853f17',
          900: '#6d3416',
        },
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
      fontFamily: {
        sans: ["var(--font-inter)", "system-ui", "sans-serif"],
        mono: ["var(--font-mono)", "Consolas", "monospace"],
      },
      fontSize: {
        '2xs': ['0.625rem', { lineHeight: '0.75rem' }],
      },
      spacing: {
        '18': '4.5rem',
        '88': '22rem',
        '128': '32rem',
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: "0" },
        },
        "fade-in": {
          "0%": { opacity: "0" },
          "100%": { opacity: "1" },
        },
        "slide-in-up": {
          "0%": { opacity: "0", transform: "translateY(20px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        "slide-in-right": {
          "0%": { opacity: "0", transform: "translateX(20px)" },
          "100%": { opacity: "1", transform: "translateX(0)" },
        },
        "glow": {
          "0%": { boxShadow: "0 0 20px hsl(var(--primary) / 0.3)" },
          "100%": { boxShadow: "0 0 30px hsl(var(--primary) / 0.6), 0 0 40px hsl(var(--primary) / 0.3)" },
        },
        "flash-profit": {
          "0%": { backgroundColor: "transparent", transform: "scale(1)" },
          "50%": { backgroundColor: "hsl(var(--profit) / 0.2)", transform: "scale(1.02)" },
          "100%": { backgroundColor: "transparent", transform: "scale(1)" },
        },
        "flash-loss": {
          "0%": { backgroundColor: "transparent", transform: "scale(1)" },
          "50%": { backgroundColor: "hsl(var(--loss) / 0.2)", transform: "scale(1.02)" },
          "100%": { backgroundColor: "transparent", transform: "scale(1)" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
        "fade-in": "fade-in 0.5s ease-in-out",
        "slide-in-up": "slide-in-up 0.5s ease-out",
        "slide-in-right": "slide-in-right 0.5s ease-out",
        "glow": "glow 2s ease-in-out infinite alternate",
        "pulse-slow": "pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite",
        "flash-profit": "flash-profit 0.6s ease-in-out",
        "flash-loss": "flash-loss 0.6s ease-in-out",
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-conic': 'conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))',
        'gradient-epic': 'linear-gradient(135deg, hsl(var(--primary)) 0%, hsl(var(--primary) / 0.8) 100%)',
        'terminal-bg': `
          radial-gradient(circle at 20% 80%, hsl(var(--primary) / 0.1) 0%, transparent 50%),
          radial-gradient(circle at 80% 20%, hsl(var(--accent) / 0.1) 0%, transparent 50%),
          hsl(var(--background))
        `,
      },
      boxShadow: {
        'epic': '0 8px 32px 0 rgba(31, 38, 135, 0.37)',
        'glass': '0 8px 32px 0 rgba(31, 38, 135, 0.37)',
        'glow': '0 0 20px hsl(var(--primary) / 0.3)',
        'glow-lg': '0 0 30px hsl(var(--primary) / 0.6), 0 0 40px hsl(var(--primary) / 0.3)',
      },
    },
  },
  plugins: [require("@tailwindcss/typography")],
}

export default config
