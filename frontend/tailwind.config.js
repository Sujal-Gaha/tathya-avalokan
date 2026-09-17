/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        workbench: {
          sidebar: '#18181b',
          editor: '#09090b',
          panel: '#121215',
          border: '#27272a',
          accent: '#3b82f6',
          active: '#27272a',
          hover: '#1e1e24',
        }
      }
    },
  },
  plugins: [],
}
