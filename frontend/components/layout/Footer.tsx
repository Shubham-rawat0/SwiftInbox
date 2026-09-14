"use client"

export function Footer() {
  return (
    <footer className="border-t border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-[#111313]">
      <div className="mx-auto flex max-w-7xl items-center justify-center px-4 py-5 sm:px-6 lg:px-8">
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Built by{" "}
          <a
            href="https://github.com/Shubham-rawat0"
            target="_blank"
            rel="noopener noreferrer"
            className="
              font-medium
              text-gray-700
              underline
              decoration-gray-300
              underline-offset-4
              transition-colors
              hover:text-pink-500
              hover:decoration-pink-500
              focus-visible:outline-none
              focus-visible:ring-2
              focus-visible:ring-pink-500
              focus-visible:ring-offset-2
              dark:text-gray-200
              dark:decoration-gray-600
              dark:hover:text-pink-400
              dark:hover:decoration-pink-400
              dark:focus-visible:ring-offset-[#111313]
            "
          >
            @Shubham Rawat
          </a>
        </p>
      </div>
    </footer>
  )
}
