import * as esbuild from 'esbuild'

await esbuild.build({
  entryPoints: [
    './assets/js/*.js'
  ],
  bundle: true,
  outdir: "./static/js",
})