#!/bin/bash
cd "$(dirname "$0")" || exit
npx tailwindcss -c ./tailwind.config.js -i ./index-src.css -o ./index-dist.css --watch --minify
