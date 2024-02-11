/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./*.go", "./*.templ", "./*.html"],
  theme: {
    theme: {
      screens: {
        'sm': {'min': '1px', 'max': '600px'},
        'md': {'min': '640px', 'max': '1023px'},
        'lg': {'min': '1024px', 'max': '1279px'},
        'xl': {'min': '1280px'},
      },
    },
    extend: {},
  },
  plugins: [],
}

