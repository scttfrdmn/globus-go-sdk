# Globus Go SDK Documentation Site

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This document describes how to maintain and update the Globus Go SDK documentation site.

## Overview

The Globus Go SDK documentation site is built using [Hugo](https://gohugo.io/) with the [Hugo Book theme](https://github.com/alex-shpak/hugo-book). The site is automatically deployed to GitHub Pages when changes are pushed to the main branch.

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

- `/doc/reference/` - Original reference documentation
  - Source Markdown files are automatically synced to the Hugo site

## Workflow

### Updating Reference Documentation

1. Make changes to the reference documentation in `/doc/reference/`.
2. Push changes to the main branch.
3. The GitHub Actions workflow will automatically sync the changes to the Hugo site and deploy it to GitHub Pages.

### Updating Guides and Examples

1. Edit or add Markdown files in `/docs-site/content/docs/guides/` or `/docs-site/content/docs/examples/`.
2. Push changes to the main branch.
3. The GitHub Actions workflow will automatically deploy the site to GitHub Pages.

### Local Development

To work on the site locally:

1. Install Hugo (Extended version recommended):
   ```bash
   brew install hugo # macOS
   sudo apt install hugo # Linux
   ```

2. Clone the Hugo Book theme:
   ```bash
   git clone https://github.com/alex-shpak/hugo-book.git docs-site/themes/hugo-book
   ```

3. Start the Hugo development server:
   ```bash
   cd docs-site
   hugo server -D
   ```

4. Open your browser and go to [http://localhost:1313/globus-go-sdk/](http://localhost:1313/globus-go-sdk/)

### Adding New Content

#### Adding a New Guide

1. Create a new Markdown file in `/docs-site/content/docs/guides/`:
   ```bash
   touch docs-site/content/docs/guides/new-guide.md
   ```

2. Add front matter to the top of the file:
   ```yaml
   ---
   title: "New Guide Title"
   weight: 10
   ---
   ```

3. Add your content using Markdown.

#### Adding a New Example

1. Create a new Markdown file in `/docs-site/content/docs/examples/`:
   ```bash
   touch docs-site/content/docs/examples/new-example.md
   ```

2. Add front matter to the top of the file:
   ```yaml
   ---
   title: "New Example Title"
   weight: 10
   ---
   ```

3. Add your content using Markdown.

### Hugo Book Theme Features

The Hugo Book theme provides several features that you can use in your content:

#### Hints

```markdown
{{< hint info >}}
**Info**
This is an info box.
{{< /hint >}}

{{< hint warning >}}
**Warning**
This is a warning box.
{{< /hint >}}

{{< hint danger >}}
**Danger**
This is a danger box.
{{< /hint >}}
```

#### Tabs

```markdown
{{< tabs "uniqueid" >}}
{{< tab "Tab 1" >}}
Content for tab 1
{{< /tab >}}
{{< tab "Tab 2" >}}
Content for tab 2
{{< /tab >}}
{{< /tabs >}}
```

#### Code Blocks

````markdown
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, world!")
}
```
````

#### Internal Links

```markdown
[Link to another page]({{< ref "/docs/reference/auth" >}})
```

## Deployment

The site is automatically deployed to GitHub Pages when changes are pushed to the main branch. The deployment is handled by a GitHub Actions workflow defined in `.github/workflows/deploy-docs.yml`.

## Adding New Services

When adding documentation for a new service:

1. Create reference documentation in `/doc/reference/new-service/`.
2. Create an index page in `/docs-site/content/docs/reference/new-service/_index.md`.
3. Update the navigation in `/docs-site/content/docs/reference/_index.md`.
4. Push changes to the main branch.

## Troubleshooting

- If the site is not deployed correctly, check the GitHub Actions workflow logs for errors.
- If the site is not rendering correctly, check for Markdown syntax errors or Hugo configuration issues.
- If the site is not updating after a push, check that the correct branch and paths are being tracked by the GitHub Actions workflow.