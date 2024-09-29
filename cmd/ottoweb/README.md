# OTTOWEB

## TailwindUI

Flakey set-up, but it seems to work.

1. Install TailwindCSS somewhere outside of this project (like the `tailwinder` directory).
2. Copy the `tailwind.config.js` there and update the paths to the `ottoweb` directory.
2. Run the watch command with the relative paths.

```bash
npx tailwindcss -i ../ottomap/cmd/ottoweb/assets/css/tailwind-input.css -o ../ottomap/cmd/ottoweb/assets/css/tailwind.css --watch
```

## Build for DO

```bash
GOOS=linux GOARCH=amd64 go build -o ottoweb.exe
```
