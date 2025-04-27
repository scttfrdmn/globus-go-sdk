# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Documentation Restructuring Plan

This document outlines a comprehensive plan to restructure the Globus Go SDK documentation for improved cohesiveness, clarity, and flow.

## Current Documentation Issues

Based on analysis of the current documentation structure, the following issues have been identified:

1. **Inconsistent Naming Conventions**: Mix of ALL_CAPS (ROADMAP.md) and kebab-case (user-guide.md)
2. **Duplicated Content**: Multiple versions of CONTRIBUTING.md and fragmented security documentation
3. **Fragmented Information**: Related topics spread across multiple files
4. **Unclear Navigation**: No central entry point or clear hierarchy for different user types
5. **Inconsistent Formatting**: Varying header styles, inconsistent SPDX license headers
6. **Documentation Gaps**: Missing or incomplete sections on important topics
7. **Poor Cross-Linking**: Limited connections between related documents

## Proposed Documentation Structure

The new documentation structure will be organized into logical categories with clear navigation paths for different user personas.

```
/
├── README.md                  # Project overview, quickstart
├── CONTRIBUTING.md            # Consolidated contribution guide
├── LICENSE                    # License information
├── doc/
│   ├── README.md              # Documentation index and navigation guide
│   ├── guides/                # User-focused guides
│   │   ├── getting-started.md # First steps with the SDK
│   │   ├── authentication.md  # Authentication guide
│   │   ├── transfer.md        # File transfer guide
│   │   ├── search.md          # Search functionality guide
│   │   ├── groups.md          # Groups management guide
│   │   ├── flows.md           # Automation flows guide
│   │   ├── compute.md         # Compute functionality guide
│   │   └── timers.md          # Timers functionality guide
│   ├── topics/                # Conceptual & feature documentation
│   │   ├── token-storage.md   # Token storage options
│   │   ├── error-handling.md  # Error handling best practices
│   │   ├── rate-limiting.md   # Rate limiting behavior
│   │   ├── logging.md         # Logging and tracing
│   │   ├── performance.md     # Performance optimization
│   │   └── data-schemas.md    # Data models & schemas
│   ├── advanced/              # Advanced use cases
│   │   ├── recursive-transfers.md # Recursive transfers
│   │   ├── resumable-transfers.md # Resumable transfers
│   │   ├── connection-pooling.md  # Connection pooling
│   │   ├── extending.md       # Extending the SDK
│   │   └── mfa.md             # Multi-factor authentication
│   ├── development/           # Developer-focused documentation
│   │   ├── architecture.md    # Design patterns & architecture
│   │   ├── contributing.md    # How to contribute
│   │   ├── testing.md         # Testing guide (unit & integration)
│   │   ├── security.md        # Security best practices & guidelines
│   │   └── benchmarking.md    # Performance benchmarking
│   ├── examples/              # Extended examples (beyond README)
│   │   ├── authentication.md  # Authentication examples
│   │   ├── transfer.md        # Transfer operation examples
│   │   ├── search.md          # Search operation examples
│   │   ├── groups.md          # Groups operation examples
│   │   ├── flows.md           # Flows operation examples
│   │   └── cli.md             # CLI usage examples
│   ├── reference/             # Technical reference
│   │   ├── api-overview.md    # API structure overview
│   │   ├── configuration.md   # Configuration options
│   │   ├── environment.md     # Environment variables
│   │   ├── error-codes.md     # Error codes & troubleshooting
│   │   └── glossary.md        # Terminology glossary
│   ├── project/               # Project metadata
│   │   ├── roadmap.md         # Development roadmap
│   │   ├── status.md          # Implementation status
│   │   ├── changelog.md       # Version history
│   │   └── alignment.md       # SDK alignment with other Globus SDKs
```

## Document Consolidation Plan

The following documents will be consolidated to reduce duplication and improve coherence:

1. **Security Documentation**
   - Merge: SECURITY_GUIDELINES.md, SECURITY_AUDIT_PLAN.md, SECURITY_TESTING.md, SECURITY_TOOLING.md
   - Into: development/security.md

2. **Testing Documentation**
   - Merge: INTEGRATION_TESTING.md, INTEGRATION_TESTING_SETUP.md, shell-testing.md
   - Into: development/testing.md

3. **Performance Documentation**
   - Merge: performance-benchmarking.md, memory-optimization.md, connection-pooling.md
   - Into: topics/performance.md (with specific aspects in advanced/)

4. **Contribution Documentation**
   - Merge: /CONTRIBUTING.md, /doc/CONTRIBUTING.md
   - Into: /CONTRIBUTING.md (primary) with links from development/contributing.md

## Implementation Strategy

The restructuring will be implemented in phases:

### Phase 1: Create Structure and Index
- Create the folder structure outlined above
- Develop doc/README.md as a central index
- Establish document templates with consistent headers

### Phase 2: Consolidate and Migrate Content
- Consolidate duplicated documents per the plan
- Move existing documents to appropriate locations
- Update internal links

### Phase 3: Standardize Formatting
- Ensure consistent SPDX headers
- Standardize document structure with uniform sections
- Implement consistent document metadata

### Phase 4: Enhance Cross-Linking
- Add "Related Topics" sections to each document
- Create bidirectional links between related documents
- Add breadcrumb navigation where appropriate

### Phase 5: Fill Documentation Gaps
- Complete missing documentation 
- Add comprehensive examples
- Expand advanced use case documentation

## Style Guidelines

To ensure consistency across all documentation:

1. **Naming Conventions**
   - All documentation files: kebab-case.md
   - Exception: Root-level project files remain ALL_CAPS.md

2. **File Structure**
   - SPDX license header at top
   - Title heading (# Title)
   - Brief description
   - Table of contents for longer documents
   - Consistent heading levels (# → ## → ###)

3. **Content Guidelines**
   - Active voice preferred
   - Code examples for all functional descriptions
   - "See Also" section at the end of documents
   - Version information where appropriate

## Tracking Progress

Implementation progress will be tracked in the GitHub project board with the following states:
- To Do
- In Progress
- Review
- Completed

## Next Steps

1. Create the folder structure and document templates
2. Develop the central index document
3. Begin consolidating duplicated content
4. Update the README.md documentation links