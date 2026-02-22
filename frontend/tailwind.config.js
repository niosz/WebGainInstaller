/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        gh: {
          'bg':        '#0d1117',
          'surface':   '#161b22',
          'overlay':   '#1c2128',
          'border':    '#30363d',
          'border-m':  '#21262d',
          'text':      '#e6edf3',
          'text-sec':  '#8b949e',
          'text-muted':'#484f58',
          'blue':      '#58a6ff',
          'green':     '#3fb950',
          'red':       '#f85149',
          'yellow':    '#d29922',
          'purple':    '#bc8cff',
          'progress':  '#238636',
          'progress-bg':'#21262d',
        },
      },
      fontFamily: {
        mono: ["'MesloLGS NF'", "'Cascadia Code'", "'Consolas'", "monospace"],
      },
    },
  },
  plugins: [],
}
