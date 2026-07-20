# AppservR Documentation Website

Hugo source for the [documentation website](https://appservR.github.io), built with the
[Doks](https://github.com/h-enk/doks) theme.

## Developing locally

Hugo itself is a standalone binary, not an npm package. `npm install` here doesn't install Hugo
as a JS dependency — its `postinstall` script (`hugo-installer`) downloads the pinned Hugo binary
into `node_modules/.bin/hugo`, and the npm scripts below just wrap calls to it (Doks also uses
Hugo's asset pipeline to run PostCSS/Babel on SCSS/JS, which do come from npm).

From the repository root:

```sh
cd docs
npm install
npm run start   # runs `hugo server`, with live reload
npm run build   # runs `hugo --gc --minify`, production build to docs/public
```

## Publishing

Pushes to `main` that touch `docs/**` are built and published automatically by
[`.github/workflows/deploy-docs.yml`](../.github/workflows/deploy-docs.yml), which pushes the
built site to the `gh-pages` branch of [appservR/appservR.github.io](https://github.com/appservR/appservR.github.io)
(a machine-generated repository — do not edit it directly).
