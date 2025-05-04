# Globus Go SDK Documentation Site

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This directory contains the source for the Globus Go SDK documentation site, which is built using [Hugo](https://gohugo.io/) and the [Hugo Book theme](https://github.com/alex-shpak/hugo-book).

## Prerequisites

To work with the documentation site locally, you need to install:

1. [Hugo](https://gohugo.io/getting-started/installing/) (Extended version recommended)
2. [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

## Setup

To set up the documentation site for development:

1. Install Hugo:
   - On macOS: `brew install hugo`
   - On Linux: `sudo apt install hugo` or follow the [official instructions](https://gohugo.io/getting-started/installing/)
   - On Windows: `choco install hugo-extended` or download from the [Hugo releases page](https://github.com/gohugoio/hugo/releases)

2. Clone the hugo-book theme into the themes directory:

```bash
git clone https://github.com/alex-shpak/hugo-book.git themes/hugo-book
```

## Local Development

To run the site locally:

1. Navigate to the docs-site directory:

```bash
cd docs-site
```

2. Start the Hugo development server:

```bash
hugo server -D
```

3. Open your browser and go to [http://localhost:1313/globus-go-sdk/](http://localhost:1313/globus-go-sdk/)

The site will automatically refresh when you make changes to the content.

## Adding Content

- Content is stored in the `content/` directory
- Reference documentation is in `content/docs/reference/`
- Guides and tutorials are in `content/docs/guides/`
- Example code and demos are in `content/docs/examples/`

## Building for Production

To build the site for production:

```bash
hugo --minify
```

This will generate the static site in the `public/` directory, which can be deployed to GitHub Pages.

## Deploying to GitHub Pages

The site is automatically deployed to GitHub Pages when changes are pushed to the main branch, using GitHub Actions. The workflow is defined in `.github/workflows/github_workflow_docs.yml`.

## Site Structure

- `config.toml`: Hugo configuration file
- `content/`: Markdown content for the site
- `static/`: Static assets like images and CSS
- `themes/hugo-book/`: The Hugo Book theme
- `public/`: Generated static site (not tracked by Git)

## Contributing

1. Create or edit Markdown files in the `content/` directory
2. Test your changes locally with `hugo server -D`
3. Commit and push your changes
4. GitHub Actions will automatically build and deploy the site