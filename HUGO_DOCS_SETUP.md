# Hugo Documentation Site Setup

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This document describes the setup and maintenance of the Hugo-based documentation site for the Globus Go SDK.

## Overview

We've set up a documentation site using [Hugo](https://gohugo.io/) with the [Hugo Book theme](https://github.com/alex-shpak/hugo-book). The site is structured to provide comprehensive documentation for all services in the SDK, including API reference, guides, and examples.

## Directory Structure

- `/docs-site/` - Contains the Hugo site source files
  - `/config.toml` - Main Hugo configuration file
  - `/content/` - Markdown content for the site
    - `/_index.md` - Home page
    - `/docs/` - Main documentation section
      - `/reference/` - API reference documentation
      - `/guides/` - How-to guides
      - `/examples/` - Example code and applications
  - `/static/` - Static assets (images, CSS, etc.)
  - `/themes/hugo-book/` - Hugo Book theme (not tracked in Git)

## Setup Steps

1. Created the basic Hugo site structure in `/docs-site/`
2. Configured Hugo in `config.toml` with appropriate settings for the Book theme
3. Created content directory structure for all documentation sections
4. Created index pages for the site, docs section, reference section, and all services
5. Created a sync script (`sync-docs.sh`) to copy reference documentation from `/doc/reference/` to Hugo content directories
6. Created a build script (`build.sh`) to build the site for deployment
7. Created a GitHub Actions workflow (`.github/workflows/deploy-docs.yml`) to automatically deploy the site to GitHub Pages when changes are pushed to the main branch
8. Updated the main README.md to include links to the online documentation site
9. Created a comprehensive guide (`doc/DOCUMENTATION_SITE.md`) for maintaining the site

## Development Workflow

1. **Local Development**:
   - Clone the Hugo Book theme: `git clone https://github.com/alex-shpak/hugo-book.git docs-site/themes/hugo-book`
   - Run the sync script: `./docs-site/sync-docs.sh`
   - Run the Hugo server: `cd docs-site && hugo server -D`
   - View the site at http://localhost:1313/globus-go-sdk/

2. **Building for Deployment**:
   - Run the build script: `./docs-site/build.sh`
   - The built site will be in the `docs-site/public` directory

3. **Automated Deployment**:
   - Push changes to the main branch
   - GitHub Actions will automatically build and deploy the site to GitHub Pages

## File Structure

- **Configuration Files**:
  - `/docs-site/config.toml` - Hugo configuration
  - `/docs-site/.gitignore` - Git ignore rules for Hugo files
  - `/.github/workflows/deploy-docs.yml` - GitHub Actions workflow

- **Scripts**:
  - `/docs-site/sync-docs.sh` - Syncs reference documentation to Hugo content
  - `/docs-site/build.sh` - Builds the site for deployment

- **Documentation**:
  - `/doc/DOCUMENTATION_SITE.md` - Guide for maintaining the site
  - `/docs-site/README.md` - Information about the docs site

## GitHub Pages Setup

The site is configured to be deployed to GitHub Pages using GitHub Actions. The workflow:

1. Checks out the code
2. Sets up Hugo
3. Installs the Hugo Book theme
4. Syncs reference documentation
5. Builds the site
6. Deploys to GitHub Pages

## Next Steps

1. **Content Creation**:
   - Add more guides and examples
   - Expand API documentation

2. **Site Enhancements**:
   - Add search functionality
   - Improve navigation
   - Add version switching

3. **Integration**:
   - Add links to the documentation site throughout the codebase
   - Update CI/CD to verify documentation builds correctly

## Resources

- [Hugo Documentation](https://gohugo.io/documentation/)
- [Hugo Book Theme Documentation](https://github.com/alex-shpak/hugo-book)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)