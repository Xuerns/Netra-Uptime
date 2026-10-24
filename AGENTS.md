# Agent Rule
## Always Validate
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
