# Crumbs Vision

## Executive Summary

Crumbs is a general-purpose storage system for work items with built-in support for exploratory work sessions. We provide a command-line tool and Go library for tracking work with trails—sequences of crumbs you can complete (merge) or abandon (backtrack). We are not a workflow engine, coordination framework, or message queue.

## The Problem

Applications need general-purpose storage for work items that is independent of any coordination framework. Current task storage systems couple directly to a specific database or workflow engine, making it difficult to switch backends or run in different environments. Worse, they lack support for exploratory work—the ability to try a path, decide it leads nowhere, and backtrack without polluting the permanent record.

When you explore a solution, you drop crumbs as you go. Sometimes the path succeeds and those crumbs become permanent work. Sometimes you hit a dead end and need to abandon the entire trail. Existing systems force you to either commit failed work or manually clean up abandoned items.

## What This Does

Crumbs solves this by providing storage with first-class support for trails. You drop crumbs (work items) as you explore. Each crumb can belong to a trail—a work session or exploration path. Trails can grow anywhere, deviate from the main path, and either complete (crumbs become permanent) or be abandoned (backtracking—the entire trail is cleaned up).

We use the breadcrumb metaphor (Hansel and Gretel) because it naturally captures how work flows. The **cupboard** holds all crumbs and trails. You **drop crumbs** as you work. You **follow the trail** to complete dependencies. You **deviate** to explore a new path. If the trail leads nowhere, you **backtrack**—abandon it and start fresh. When a trail succeeds, you **sweep up**—complete it and merge crumbs into the permanent record.

The storage system supports multiple backends (local JSON files, Dolt for version control, DynamoDB for cloud scale) with a pluggable architecture. All identifiers use UUID v7 for time-ordered, sortable IDs. Properties are first-class entities with extensible schemas—you define new properties at runtime. Metadata tables (comments, attachments, logs) can be added without changing the core schema.

We provide both a command-line tool for personal use and a Go library for applications that need general-purpose task storage with trail support.

## What Success Looks Like

We measure success along three dimensions.

**Performance and Scale**: Operations complete with low latency as crumb counts and concurrent trails grow. We establish performance baselines as the codebase expands and refine targets based on real usage patterns.

**Developer Experience**: Developers integrate the Go library quickly. The API is asynchronous, type-safe, and self-explanatory. Adding a new backend takes hours, not days. Defining new properties or metadata tables requires no schema migrations.

**Trail Workflow**: Users create trails, drop crumbs, and abandon dead-end explorations without manual cleanup. Completed trails merge seamlessly into the permanent record. The command-line tool makes trail workflows natural and fast.

## What This Is NOT

We are not building a workflow engine. Coordination semantics (claiming work, timeouts, announcements) belong in layers above this storage—frameworks like Task Fountain that build on Crumbs.

We are not building a message queue. Crumbs stores work items; it does not route messages or provide pub/sub.

We are not building an HTTP/RPC API. Applications using Crumbs define their own APIs. The command-line tool provides a local interface; distributed coordination is out of scope.

We are not building replication or multi-region support. Backends may provide these features natively (DynamoDB global tables, Dolt remotes), but replication is not a core Crumbs concern.

We are not building a general-purpose database. Crumbs is purpose-built for work item storage with trails, properties, and metadata.
