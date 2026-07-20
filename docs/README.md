# AppservR Documentation Website

Hugo source for the [documentation website](https://appservR.github.io), built with the
[Doks](https://github.com/h-enk/doks) theme.

## Developing locally

```sh
cd docs
npm install
npm run start   # dev server with live reload
npm run build   # production build to docs/public
```

## Publishing

Pushes to `main` that touch `docs/**` are built and published automatically by
[`.github/workflows/deploy-docs.yml`](../.github/workflows/deploy-docs.yml), which pushes the
built site to the `gh-pages` branch of [appservR/appservR.github.io](https://github.com/appservR/appservR.github.io)
(a machine-generated repository — do not edit it directly).
