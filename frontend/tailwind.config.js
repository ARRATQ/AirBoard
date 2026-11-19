/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Accent de la palette active (variables injectées par stores/app.js).
        primary: {
          50: 'rgb(var(--primary-50) / <alpha-value>)',
          100: 'rgb(var(--primary-100) / <alpha-value>)',
          200: 'rgb(var(--primary-200) / <alpha-value>)',
          300: 'rgb(var(--primary-300) / <alpha-value>)',
          400: 'rgb(var(--primary-400) / <alpha-value>)',
          500: 'rgb(var(--primary-500) / <alpha-value>)',
          600: 'rgb(var(--primary-600) / <alpha-value>)',
          700: 'rgb(var(--primary-700) / <alpha-value>)',
          800: 'rgb(var(--primary-800) / <alpha-value>)',
          900: 'rgb(var(--primary-900) / <alpha-value>)',
          950: 'rgb(var(--primary-950) / <alpha-value>)',
        },
        // Gris neutres alignés sur les tokens A+ (surface, sunken, line, muted…).
        gray: {
          50: '#f4f5f6',
          100: '#eceef0',
          200: '#e0e3e6',
          300: '#d6dade',
          400: '#959da5',
          500: '#6b737b',
          600: '#596068',
          700: '#272c31',
          800: '#15181b',
          900: '#0e1012',
          950: '#050607',
        },
      },
      fontFamily: {
        sans: ['var(--f-body)'],
        mono: ['var(--f-num)'],
      },
      // Direction A+ : 3 px (badges), 4 px (boutons, champs), 6 px (cartes, panneaux).
      borderRadius: {
        sm: '3px',
        DEFAULT: '4px',
        md: '4px',
        lg: '6px',
        xl: '6px',
        '2xl': '6px',
      },
      // Pas d'ombre sur les cartes ; une seule ombre pour les éléments flottants.
      boxShadow: {
        sm: 'none',
        DEFAULT: 'none',
        md: 'none',
        lg: 'var(--shadow-float)',
        xl: 'var(--shadow-float)',
        '2xl': 'var(--shadow-float)',
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'slide-in': 'slideIn 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideIn: {
          '0%': { transform: 'translateX(-100%)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
      },
      backdropBlur: {
        xs: '2px',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}