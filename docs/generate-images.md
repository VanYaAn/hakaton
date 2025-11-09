# Generate Images from Mermaid Diagrams

## Option 1: Use Mermaid Live Editor

Visit: https://mermaid.live/

Copy-paste any diagram from the documentation files and export as:
- PNG
- SVG
- PDF

## Option 2: Use Command Line Tool

Install mermaid-cli:
```bash
npm install -g @mermaid-js/mermaid-cli
```

Generate images:
```bash
# From architecture.md
mmdc -i docs/architecture.md -o docs/images/architecture.png

# Or extract specific diagrams to separate files
```

## Option 3: VS Code Extension

Install: "Markdown Preview Mermaid Support"

Then right-click on diagram → Export to PNG

## Option 4: Automated Script

Create a script to extract and convert all diagrams:

```bash
#!/bin/bash
mkdir -p docs/images

# This would extract mermaid blocks and convert them
# Requires mermaid-cli installed
```

## Quick Links

### Architecture Diagrams
- System Overview: Copy from `architecture.md` lines 7-71
- Deployment: Copy from `architecture.md` lines 95-118
- Configuration: Copy from `architecture.md` lines 123-140

### Database Diagrams
- ERD: Copy from `database-schema.md` lines 7-61
- Data Flow: Copy from `database-schema.md` lines 211-233

### Workflow Diagrams
- Meeting Creation: Copy from `workflows.md` lines 7-72
- Voting: Copy from `workflows.md` lines 78-136
- Close Voting: Copy from `workflows.md` lines 142-193

### Connection Diagrams
- Component Dependencies: Copy from `connections.md` lines 7-85
- Data Flow: Copy from `connections.md` lines 90-140
- Network Topology: Copy from `connections.md` lines 269-299
