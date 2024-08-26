/** @type {import('tailwindcss').Config} */
import daisyui from "daisyui"

module.exports = {
  content: [
    "./templates/*.html",
    "./templates/**/*.html",
    "./assets/js/*.js",
    "./config.json"
  ],
  theme: {
    extend: {},
  },
  plugins: [
    daisyui,
  ],
  safelist: [],
}

