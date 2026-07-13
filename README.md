# Neural Junkie — Software development pack

Official domain pack for [Neural Junkie](https://github.com/camronwood/neural-junkie).

Adds engineering specialists (BackendEngineer, FrontendEngineer, …) with pack-owned `sd-mcp-server` and Qwen 3.5 model defaults.

**IDE features** (layout, Git SCM, LSP, composer) are in the separate **[IDE pack](https://github.com/camronwood/neural-junkie-pack-ide)** — independent of this pack.

Install via desktop **Settings → Domain packs → Pack store**, or sideload `dist/software-development-<version>.zip`.

## Develop

```bash
make verify
make pack-zip   # dist/software-development-<version>.zip
```

Tag `v2.1.0` and push to publish the release zip to GitHub.
