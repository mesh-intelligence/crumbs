# Crumbs Vision

## Executive Summary

Crumbs is a general-purpose storage system for work items with built-in support for exploratory work sessions. We provide a command-line tool and Go library for tracking work with trails—sequences of crumbs you can complete (merge) or abandon (backtrack). We are not a workflow engine, coordination framework, or message queue.

## The Problem

Coding agents need storage for work items that supports how they actually work—exploring solutions, hitting dead ends, and backtracking. When an agent explores a solution approach, it creates tasks and subtasks. Sometimes the approach works and those tasks become permanent work. Sometimes the agent realizes the approach will not work and needs to abandon the entire exploration without polluting the main task list.

Current task storage systems lack support for this exploratory workflow. They couple directly to a specific database or workflow engine, making it difficult to switch backends. More importantly, they have no concept of tentative work sessions. Agents must either commit failed exploration tasks to the permanent record or manually track and clean up abandoned items. Neither is acceptable—the first pollutes the task history with dead ends, the second is error-prone and complex.

## What This Does

Crumbs solves this by providing storage with first-class support for trails. You drop crumbs (work items) as you explore. Each crumb can belong to a trail—a work session or exploration path. Trails can grow anywhere, deviate from the main path, and either complete (crumbs become permanent) or be abandoned (backtracking—the entire trail is cleaned up).

We use the breadcrumb metaphor (Hansel and Gretel) because it naturally captures how work flows. The **cupboard** holds all crumbs and trails. You **drop crumbs** as you work. You **follow the trail** to complete dependencies. You **deviate** to explore a new path. If the trail leads nowhere, you **backtrack**—abandon it and start fresh. When a trail succeeds, you **sweep up**—complete it and merge crumbs into the permanent record.

The storage system supports multiple backends (local JSON files, Dolt for version control, DynamoDB for cloud scale) with a pluggable architecture. All identifiers use UUID v7 for time-ordered, sortable IDs. Properties are first-class entities with extensible schemas—you define new properties at runtime. Metadata tables (comments, attachments, logs) can be added without changing the core schema.

We provide both a command-line tool and a Go library. The primary use case is coding agents—the first implementation is a VS Code coding agent that uses trails to explore implementation approaches, complete successful paths, and abandon dead ends atomically. The library and CLI also support personal task tracking and other agent workflows.

## What Success Looks Like

We measure success along three dimensions.

**Performance and Scale**: Operations complete with low latency as crumb counts and concurrent trails grow. We establish performance baselines as the codebase expands and refine targets based on real usage patterns.

**Developer Experience**: Developers integrate the Go library quickly. The API is asynchronous, type-safe, and self-explanatory. Adding a new backend takes hours, not days. Defining new properties or metadata tables requires no schema migrations.

**Agent Workflow**: Coding agents create trails for exploration, drop crumbs as they plan implementation steps, and abandon dead-end approaches without manual cleanup. Completed trails merge seamlessly into the permanent task list. The VS Code agent demonstrates that trail-based exploration feels natural and improves code quality by encouraging agents to explore alternatives.

## What This Is NOT

We are not building a workflow engine. Coordination semantics (claiming work, timeouts, announcements) belong in layers above this storage—frameworks like Task Fountain that build on Crumbs.

We are not building a message queue. Crumbs stores work items; it does not route messages or provide pub/sub.

We are not building an HTTP/RPC API. Applications using Crumbs define their own APIs. The command-line tool provides a local interface; distributed coordination is out of scope.

We are not building replication or multi-region support. Backends may provide these features natively (DynamoDB global tables, Dolt remotes), but replication is not a core Crumbs concern.

We are not building a general-purpose database. Crumbs is purpose-built for work item storage with trails, properties, and metadata.
