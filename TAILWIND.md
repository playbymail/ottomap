# Tailwind

## License
This project uses both
[Tailwind CSS](https://tailwindcss.com/)
and
[Tailwind UI](https://tailwindui.com/).

### Tailwind CSS license

```text
Copyright (c) Tailwind Labs, Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

### Tailwind UI license
Tailwind UI is a commercial template.
It is not available under any open-sourc license.
Files in the `templates` directory may not be copied, distributed, or used in 
any other project without first purchasing a license from Tailwind.

## Setup

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash

nvm update

nvm version

nvm install-latest-npm

# this command doesn't work; caniuse-lite is borked
npx update-browserslist-db@latest

npm install -D tailwindcss

mkdir -p templates public/css

npx tailwindcss init

# add tailwindcss/forms and install
vi tailwind.config.js
npm install -D @tailwindcss/forms

npx tailwindcss -i public/css/input.css -o public/css/tailwind.css --watch
```

Or for the HTMx server

npx tailwindcss -i internal/servers/htmx/assets/css/tailwind-input.css -o internal/servers/htmx/assets/css/tailwind.css --watch
