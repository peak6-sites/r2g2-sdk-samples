Loader Configuration Examples
=============================

This directory contains example YAML configuration files to configure RAG Store loaders.

You can use the `r2g2` tool to manage the loader configurations for a store.

Currently, the supported loader types are:
- confluence
- readme
- zendesk

### Examples

Retrieve the current configuration:

```bash
r2g2 rag loaders config get STORE_1234 confluence
```

Update a configuration:

```bash
r2g2 rag loaders config set -f confluence.yaml STORE_1234
```

```bash
r2g2 rag loaders config set STORE_1234 < confluence.yaml 
```

Delete a configuration:

```bash
r2g2 rag loaders config delete STORE_1234 confluence
```