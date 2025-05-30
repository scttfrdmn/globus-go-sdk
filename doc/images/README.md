# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Documentation Images

This directory contains images used in the Globus Go SDK documentation. All images should follow these guidelines:

1. Use descriptive filenames that reflect the content (e.g., `auth-flow-diagram.png`)
2. Keep images under 1MB in size
3. Use SVG format for diagrams when possible
4. Provide alt text in documentation references

## Project Logo

The `globus-go-sdk-logo.png` is the official logo for the Globus Go SDK project. It is used in the README and other documentation.

### Logo Usage Guidelines

When using the logo:

1. Do not modify the logo colors or proportions
2. Maintain adequate space around the logo when displaying it
3. Use the logo at an appropriate size for the context
4. For web usage, the recommended width is 400px
5. For presentations, ensure the logo is clearly visible

## Naming Conventions

Use kebab-case for filenames:
- `service-name-feature.png` (e.g., `transfer-recursive-diagram.png`)
- `concept-name-diagram.png` (e.g., `token-refresh-flow.png`)
- `architecture-component.png` (e.g., `architecture-overview.png`)

## Organization

Images should be organized in subdirectories by service or concept when the number of images grows:

```
/images
  /auth
  /transfer
  /groups
  /architecture
```

## Adding New Images

When adding new images:

1. Optimize images for web viewing (compress when appropriate)
2. Use consistent styling for related diagrams
3. Include a source file (e.g., `.drawio`, `.svg`) when available
4. Update this README with any new guidelines or conventions