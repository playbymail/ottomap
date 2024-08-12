/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.{gohtml,html,js}"],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}

