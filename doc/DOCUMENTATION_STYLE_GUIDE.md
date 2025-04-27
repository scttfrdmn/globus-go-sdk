# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Documentation Style Guide

_Last Updated: April 27, 2025_

This style guide establishes standards for the Globus Go SDK documentation to ensure consistency, clarity, and usability.

> **DISCLAIMER**: The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus, the University of Chicago, or their affiliated organizations.

## Table of Contents

- [Document Structure](#document-structure)
- [Document Types](#document-types)
- [Formatting Conventions](#formatting-conventions)
- [Code Examples](#code-examples)
- [Cross-References](#cross-references)
- [Images and Diagrams](#images-and-diagrams)
- [Version Information](#version-information)
- [Templates](#templates)

## Document Structure

All documentation files should follow this basic structure:

```markdown
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Document Title

_Last Updated: [Date]_
_Compatible with SDK versions: vX.Y and above_

> **DISCLAIMER**: The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus, the University of Chicago, or their affiliated organizations.

[Brief introduction/overview]

## Table of Contents

- [Section 1](#section-1)
- [Section 2](#section-2)
- ...

## Section 1

[Content]

## Section 2

[Content]

## Related Topics

- [Link to related topic 1](path/to/topic1.md)
- [Link to related topic 2](path/to/topic2.md)
```

## Document Types

### Guides
- Purpose: Step-by-step instructions for performing tasks
- Location: `doc/guides/`
- Sections: Overview, Prerequisites, Step-by-step instructions, Examples, Troubleshooting

### Topics
- Purpose: Conceptual explanations of features and patterns
- Location: `doc/topics/`
- Sections: Overview, Explanation, Examples, Best Practices

### Advanced
- Purpose: In-depth coverage of advanced features
- Location: `doc/advanced/`
- Sections: Overview, Prerequisites, Technical details, Example use cases

### Development
- Purpose: Information for SDK contributors
- Location: `doc/development/`
- Sections: Purpose, Requirements, Process, Guidelines

### Reference
- Purpose: Technical details and specifications
- Location: `doc/reference/`
- Sections: Overview, Specifications, Examples

### Project
- Purpose: Project-level information
- Location: `doc/project/`
- Sections: Varies by document type

## Formatting Conventions

### Headings

- Use title case for the main document title (H1)
- Use sentence case for all other headings (H2-H6)
- Limit heading nesting to 4 levels (H1-H4)
- Include one blank line before and after each heading

### Text Formatting

- Use **bold** for emphasis of important terms or warnings
- Use *italics* for introducing new terms or slight emphasis
- Use `code formatting` for code elements, file names, and technical terms

### Lists

- Use bulleted lists for unordered items
- Use numbered lists for sequential steps or prioritized items
- Maintain consistent capitalization and punctuation in lists
- Indent sublists consistently (4 spaces)

### Blockquotes and Notes

- Use blockquotes for callouts, notes, warnings, and tips
- Use consistent formatting for different types of notes:

```markdown
> **NOTE**: This is a general note.

> **WARNING**: This is a warning or critical information.

> **TIP**: This is a helpful suggestion.
```

## Code Examples

### Inline Code

- Use single backticks for inline code elements: `variable`, `method()`, `package/module`
- Use inline code formatting for command-line commands: `go test ./...`

### Code Blocks

- Use triple backticks with language specifier for code blocks
- Include language identifier for syntax highlighting (e.g., ```go, ```bash)
- Keep code examples concise and focused
- Include comments for complex or non-obvious code
- Include error handling in examples

Example:

~~~markdown
```go
// Create a transfer client
client := pkg.NewTransferClient(ctx, "your-access-token")

// Submit a transfer request
result, err := client.SubmitTransfer(
    ctx,
    sourceEndpoint,
    destinationEndpoint,
    &pkg.TransferOptions{
        Label: "Example Transfer",
    },
)
if err != nil {
    log.Fatalf("Transfer failed: %v", err)
}
```
~~~

### Example Patterns

- Start with simple examples and progress to more complex ones
- Include complete working examples when possible
- Include error handling in all examples
- Highlight important lines or concepts in comments

## Cross-References

### Internal Links

- Use relative paths for linking to other documentation files
- Include descriptive link text: [Authentication Guide](guides/authentication.md) instead of [click here](guides/authentication.md)
- Link to specific sections when appropriate: [Error Handling](topics/errors.md#handling-rate-limit-errors)

### Related Topics Section

- Include a "Related Topics" section at the end of each document
- List 2-5 most relevant related documents
- Use descriptive link text

Example:

```markdown
## Related Topics

- [Rate Limiting](topics/rate-limiting.md)
- [Error Handling](topics/error-handling.md)
- [Performance Optimization](topics/performance.md)
```

## Images and Diagrams

- Store images in `doc/images/` directory
- Use descriptive file names: `auth-flow-diagram.png` instead of `diagram1.png`
- Include alt text for all images
- Keep images under 1MB in size
- Use SVG format for diagrams when possible
- Include a caption below the image

Example:

```markdown
![Authentication flow diagram](../images/auth-flow-diagram.png)
*Figure 1: OAuth 2.0 Authorization Code Flow*
```

## Version Information

- Include SDK version compatibility information near the top of the document
- Note when features were introduced in specific versions
- Use consistent version notation: vX.Y.Z

Example:

```markdown
_Compatible with SDK versions: v0.5.0 and above_

> **NOTE**: The recursive transfer feature was introduced in v0.7.0
```

## Templates

### Service Guide Template

```markdown
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# [Service] Service Guide

_Last Updated: [Date]_
_Compatible with SDK versions: vX.Y and above_

> **DISCLAIMER**: The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus, the University of Chicago, or their affiliated organizations.

[Brief introduction to the service and its purpose]

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Basic Operations](#basic-operations)
- [Advanced Features](#advanced-features)
- [Error Handling](#error-handling)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)
- [Related Topics](#related-topics)

## Overview

[Description of the service and its capabilities]

## Authentication

[Required scopes and authentication process]

## Basic Operations

[Core functionality with examples]

## Advanced Features

[Advanced usage patterns]

## Error Handling

[Service-specific error handling]

## Examples

[Complete code examples]

## Troubleshooting

[Common issues and solutions]

## Related Topics

- [Link to related topic 1](path/to/topic1.md)
- [Link to related topic 2](path/to/topic2.md)
```

### Topic Template

```markdown
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# [Topic Name]

_Last Updated: [Date]_
_Compatible with SDK versions: vX.Y and above_

> **DISCLAIMER**: The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus, the University of Chicago, or their affiliated organizations.

[Brief introduction to the topic]

## Table of Contents

- [Overview](#overview)
- [Key Concepts](#key-concepts)
- [Implementation](#implementation)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Related Topics](#related-topics)

## Overview

[General explanation of the topic]

## Key Concepts

[Explanation of important concepts]

## Implementation

[How to implement the feature/concept]

## Examples

[Code examples demonstrating the topic]

## Best Practices

[Recommendations for using the feature effectively]

## Related Topics

- [Link to related topic 1](path/to/topic1.md)
- [Link to related topic 2](path/to/topic2.md)
```