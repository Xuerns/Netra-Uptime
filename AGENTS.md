## Rule For Agents
# Always Validate
- Run all validation in server folder if you changed server code
- Run 'go vet ./...' on server folder
- Run 'gosec ./...' on server folder
- Run 'go test ./...' on server folder
- Run 'go fmt ./...' on server folder
- Run 'go build ./...' on server folder

- Run all validation in client folder if you changed client code
- Run 'pnpm lint' on client folder
- Run 'pnpm run check' on client folder
- Run 'pnpm run format' on client folder
- Run 'pnpm run build' on client folder

# Client Rules
- Always use **pnpm** as the package manager
- Always use **SvelteKit (Svelte 5 + Runes)** as the framework
- Always use **TypeScript** as the language
- Always use **Tailwind CSS** as the styling
- Always use **lucide-svelte** as the icons
- Never use emoji in the code
- Always make the code clean and readable
- Always follow the **SvelteKit** guidelines
- Never use shadow drop
- Never use pulse animation
- Makesure the library save

# E2E Testing
- Always run e2e testing after changed client code
- Run 'pnpm test:e2e' on client folder

# Important Notes
- Never hardcode any API keys and creditials, use environment variables (.env) instead

# Documentation you can refer
- Sveltekit: https://svelte.dev/docs/kit/
- Tailwind: https://tailwindcss.com/docs/
- Lucide: https://lucide.dev/icons/