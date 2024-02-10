/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./*.go", "./*.templ"],
  theme: {
    theme: {
      screens: {
        'sm': {'min': '1px', 'max': '768px'},
        'md': {'min': '768px', 'max': '1023px'},
        'lg': {'min': '1024px', 'max': '1279px'},
        'xl': {'min': '1280px'},
      },
    },
    extend: {},
  },
  plugins: [],
}

