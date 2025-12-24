/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        up: '#00c087',
        down: '#ff3b30',
        card: '#1e1e1e',
        background: '#121212',
      }
    },
  },
  plugins: [],
}
