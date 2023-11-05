# Example Go Fiber App, with Templ and Goth Auth

This project is a simple Go example project which uses:

- [Fiber](https://github.com/gofiber/fiber) - a web framework
- [Templ](https://github.com/a-h/templ) - a HTML templating engine
- [Goth](https://github.com/markbates/goth) - a multi-provider authentication library for Go
- [Goth Fiber](https://github.com/shareed2k/goth_fiber) - a wrapper for Goth to use with Fiber

The example uses AAD based authentication.

## Running with Air

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Set environment variables
export AUTH_TARGET_APPLICATION_ID="YOUR_AAD_APPLICATION_ID"
export AUTH_CLIENT_SECRET="YOUR_AAD_CLIENT_SECRET"

# Run with Air
air
```
