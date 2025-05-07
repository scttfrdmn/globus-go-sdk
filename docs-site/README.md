<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK Documentation Site

This directory contains the documentation site for the Globus Go SDK, built using [Hugo](https://gohugo.io/) with the [hugo-book](https://github.com/alex-shpak/hugo-book) theme.

## Development

### Prerequisites

- [Hugo](https://gohugo.io/getting-started/installing/) (Extended version recommended)
- Git

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/scttfrdmn/globus-go-sdk.git
   cd globus-go-sdk/docs-site
   ```

2. Install the hugo-book theme:
   ```bash
   git clone https://github.com/alex-shpak/hugo-book.git themes/hugo-book
   ```

3. Start the local development server:
   ```bash
   hugo serve
   ```

4. Visit http://localhost:1313 to view the site.

### Directory Structure

- `content/` - Markdown content for the site
  - `docs/` - Documentation pages
    - `guides/` - How-to guides
    - `reference/` - API reference
    - `examples/` - Example applications
    - `faq/` - Frequently asked questions
  - `_index.md` - Homepage content
- `static/` - Static assets (images, etc.)
- `layouts/` - Custom Hugo layouts
- `assets/` - CSS, JS, and other assets

### Adding Content

1. Create a new page:
   ```bash
   hugo new docs/guides/my-guide.md
   ```

2. Edit the page in Markdown format.

### Building the Site

Run the build script to generate the site:

```bash
./build.sh
```

This will create the static site in the `public/` directory.

## Deployment

The site is automatically deployed to GitHub Pages when changes are pushed to the main branch or when a new version tag is created. The deployment is handled by the GitHub Actions workflow in `.github/workflows/pages.yml`.

### Manual Deployment

If needed, you can manually trigger the deployment by running:

```bash
cd ..  # Return to repo root
git tag vX.Y.Z  # Tag a new version
git push origin vX.Y.Z  # Push the tag
```

Or by manually triggering the workflow in the GitHub Actions tab.

## Versioning

The documentation site supports versioned documentation. When a new version tag is pushed, a new version of the documentation is created and accessible via the version selector in the UI.

Configuration for versions is in `config/_default/config.toml` under the `[params]` section.

## Adding Documentation for New Features

1. Add content to the appropriate section in `content/docs/`.
2. Update the example code to showcase the new feature.
3. Ensure API reference documentation is generated from code comments.
4. Add any necessary images or diagrams to `static/images/`.
5. Update the navigation in `content/menu/index.md` if needed.