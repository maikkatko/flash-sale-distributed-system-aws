const fs = require('fs');

// Professional styling
const tableBorder = { style: BorderStyle.SINGLE, size: 1, color: "CCCCCC" };
const cellBorders = { top: tableBorder, bottom: tableBorder, left: tableBorder, right: tableBorder };
const headerShading = { fill: "2B579A", type: ShadingType.CLEAR };
const altRowShading = { fill: "F5F5F5", type: ShadingType.CLEAR };

const doc = new Document({
  styles: {
    default: { document: { run: { font: "Arial", size: 22 } } },
    paragraphStyles: [
      { id: "Title", name: "Title", basedOn: "Normal",
        run: { size: 48, bold: true, color: "2B579A", font: "Arial" },
        paragraph: { spacing: { before: 0, after: 200 }, alignment: AlignmentType.CENTER } },
      { id: "Heading1", name: "Heading 1", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 32, bold: true, color: "2B579A", font: "Arial" },
        paragraph: { spacing: { before: 400, afterasdf : 200 }, outlineLevel: 0 } },
      { id: "Heading2", name: "Heading 2", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 26, bold: true, color: "404040", font: "Arial" },
        paragraph: { spacing: { before: 300, after: 150 }, outlineLevel: 1 } },
      { id: "Heading3", name: "Heading 3", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 24, bold: true, color: "595959", font: "Arial" },
        paragraph: { spacing: { before: 200, after: 100 }, outlineLevel: 2 } },
    ]
  },
  numbering: {
    config: [
      { reference: "main-bullets",
        levels: [
          { level: 0, format: LevelFormat.BULLET, text: "•", alignment: AlignmentType.LEFT,
            style: { paragraph: { indent: { left: 720, hanging: 360 } } } },
          { level: 1, format: LevelFormat.BULLET, text: "○", alignment: AlignmentType.LEFT,
            style: { paragraph: { indent: { left: 1440, hanging: 360 } } } }
        ] },
      { reference: "numbered-list",
        levels: [{ level: 0, format: LevelFormat.DECIMAL, text: "%1.", alignment: AlignmentType.LEFT,
          style: { paragraph: { indent: { left: 720, hanging: 360 } } } }] },
      { reference: "phase-list",
        levels: [{ level: 0, format: LevelFormat.DECIMAL, text: "%1.", alignment: AlignmentType.LEFT,
          style: { paragraph: { indent: { left: 720, hanging: 360 } } } }] },
      { reference: "component-bullets",
        levels: [{ level: 0, format: LevelFormat.BULLET, text: "•", alignment: AlignmentType.LEFT,
          style: { paragraph: { indent: { left: 720, hanging: 360 } } } }] },
      { reference: "config-bullets",
        levels: [{ level: 0, format: LevelFormat.BULLET, text: "•", alignment: AlignmentType.LEFT,
          style: { paragraph: { indent: { left: 720, hanging: 360 } } } }] },
      { reference: "redis-bullets",
        levels: [{ level: 0, format: LevelFormat.BULLET, text: "•", alignment: AlignmentType.LEFT,
          style: { paragraph: { indent: { left: 720, hanging: 360 } } } }] },
      { reference: "sqs-bullets",
        levels: [{ level: 0, format: LevelFormat.BULLET, text: "•", alignment: AlignmentType.LEFT,
          style: { paragraph: { indent: { left: 720, hanging: 360 } } } }] },
      { reference: "cb-bullets",
        levels: [{ level: 0, format: LevelFormat.BULLET, text: "•", alignment: AlignmentType.LEFT,
          style: { paragraph: { indent: { left: 720, hanging: 360 } } } }] },
      { reference: "scale-bullets",
        levels: [{ level: 0, format: LevelFormat.BULLET, text: "•", alignment: AlignmentType.LEFT,
          style: { paragraph: { indent: { left: 720, hanging: 360 } } } }] },
      { reference: "failover-bullets",
        levels: [{ level: 0, format: LevelFormat.BULLET, text: "•", alignment: AlignmentType.LEFT,
          style: { paragraph: { indent: { left: 720, hanging: 360 } } } }] },
      { reference: "monitoring-bullets",
        levels: [{ level: 0, format: LevelFormat.BULLET, text: "•", alignment: AlignmentType.LEFT,
          style: { paragraph: { indent: { left: 720, hanging: 360 } } } }] },
    ]
  },
  sections: [{
    properties: {
      page: { margin: { top: 1440, right: 1440, bottom: 1440, left: 1440 } }
    },
    headers: {
      default: new Header({ children: [new Paragraph({ 
        alignment: AlignmentType.RIGHT,
        children: [new TextRun({ text: "CS6650 Flash Sale Architecture", italics: true, size: 20, color: "666666" })]
      })] })
    },
    footers: {
      default: new Footer({ children: [new Paragraph({ 
        alignment: AlignmentType.CENTER,
        children: [new TextRun({ text: "Page ", size: 20 }), new TextRun({ children: [PageNumber.CURRENT], size: 20 }), new TextRun({ text: " of ", size: 20 }), new TextRun({ children: [PageNumber.TOTAL_PAGES], size: 20 })]
      })] })
    },
    children: [
      // Title
      new Paragraph({ heading: HeadingLevel.TITLE, children: [new TextRun("Flash Sale System Architecture")] }),
      new Paragraph({ alignment: AlignmentType.CENTER, spacing: { after: 400 },
        children: [new TextRun({ text: "High-Contention E-Commerce Platform Design", size: 24, color: "666666" })] }),

      // Executive Summary
      new Paragraph({ heading: HeadingLevel.HEADING_1, children: [new TextRun("Executive Summary")] }),
      new Paragraph({ spacing: { after: 200 }, children: [
        new TextRun("This document details the architecture for a flash sale system designed to handle 500+ concurrent users competing for limited inventory (10-100 items) with zero overselling. The architecture prioritizes consistency over availability (CP in CAP theorem) while maintaining sub-100ms P95 latency during normal operations.")
      ]}),

      // Architecture Overview
      new Paragraph({ heading: HeadingLevel.HEADING_1, children: [new TextRun("Architecture Overview")] }),
      
      // ASCII Diagram
      new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun("System Flow Diagram")] }),
      new Paragraph({ spacing: { after: 100 }, children: [
        new TextRun({ text: "The following diagram illustrates the complete request flow from user to database:", size: 22 })
      ]}),

      // Diagram as monospace text block
      new Paragraph({ spacing: { before: 200, after: 200 }, children: [
        new TextRun({                          text: "┌────────────────────────────────────────────────────────────────────────────────┐", font: "Courier New", size: 18 })]}),
      new Paragraph({ children: [new TextRun({ text: "│                           FLASH SALE ARCHITECTURE                              │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "├────────────────────────────────────────────────────────────────────────────────┤", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                                                                │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│    ┌──────────┐         ┌─────────────┐         ┌────────────────────┐         │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│    │  Users   │─────────│     ALB     │─────────│   Flash Sale API   │         │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│    │ (Locust) │         │             │         │   (Go - ECS)       │         │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│    └──────────┘         └─────────────┘         └─────────┬──────────┘         │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                                           │                    │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                              ┌────────────────────────────┼────────────┐       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                              │                            │            │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                              ▼                            ▼            │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                   ┌──────────────────┐        ┌───────────────────┐    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                   │  Redis Cluster   │        │    SQS Queues     │    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                   │  (ElastiCache)   │        │                   │    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                   │                  │        │  ┌─────────────┐  │    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                   │ • Product Cache  │        │  │ Order Queue │  │    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                   │ • Inventory Cntr │        │  └──────┬──────┘  │    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                   │ • Dist. Locks    │        │         │         │    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                   │ • Reservations   │        │  ┌──────┴──────┐  │    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                   └──────────────────┘        │  │  DLQ Queue  │  │    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                               │  └─────────────┘  │    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                               └─────────┬─────────┘    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                                         │              │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                                         ▼              │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                               ┌───────────────────┐    │       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                               │  Order Processor  │────┘       │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                               │    (Go - ECS)     │            │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                               └─────────┬─────────┘            │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                                         │                      │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                                         ▼                      │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                               ┌───────────────────┐            │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                               │   RDS PostgreSQL  │            │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                               │    (Multi-AZ)     │            │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                               └───────────────────┘            │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "│                                                                                │", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "└────────────────────────────────────────────────────────────────────────────────┘", font: "Courier New", size: 18 })] }),
      
      // Request Flow
      new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun("Request Flow: Checkout Operation")] }),
      new Paragraph({ spacing: { after: 100 }, children: [
        new TextRun("The checkout operation is the most critical path. Here's the two-phase commit flow:")
      ]}),
      
      // Phase 1
      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Phase 1: Synchronous Reservation (< 50ms target)")] }),
      new Paragraph({ numbering: { reference: "phase-list", level: 0 }, children: [new TextRun("User submits checkout request via ALB to Flash Sale API")] }),
      new Paragraph({ numbering: { reference: "phase-list", level: 0 }, children: [new TextRun("API validates request and generates idempotency key")] }),
      new Paragraph({ numbering: { reference: "phase-list", level: 0 }, children: [new TextRun("Redis Lua script atomically: CHECK inventory > 0 → DECR inventory → SET reservation with TTL")] }),
      new Paragraph({ numbering: { reference: "phase-list", level: 0 }, children: [new TextRun("If successful, publish order message to SQS")] }),
      new Paragraph({ numbering: { reference: "phase-list", level: 0 }, children: [new TextRun("Return 202 Accepted with order tracking ID")] }),

      // Phase 2
      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Phase 2: Asynchronous Confirmation")] }),
      new Paragraph({ numbering: { reference: "numbered-list", level: 0 }, children: [new TextRun("Order Processor polls SQS for messages")] }),
      new Paragraph({ numbering: { reference: "numbered-list", level: 0 }, children: [new TextRun("Simulates payment processing (configurable delay)")] }),
      new Paragraph({ numbering: { reference: "numbered-list", level: 0 }, children: [new TextRun("On success: persist order to RDS, delete reservation from Redis")] }),
      new Paragraph({ numbering: { reference: "numbered-list", level: 0 }, children: [new TextRun("On failure: release reservation (INCR inventory counter), move to DLQ")] }),

      // Component Details
      new Paragraph({ heading: HeadingLevel.HEADING_1, children: [new TextRun("Component Specifications")] }),

      // Flash Sale API
      new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun("Flash Sale API Service")] }),
      new Paragraph({ spacing: { after: 100 }, children: [new TextRun("The primary entry point for all user-facing operations.")] }),
      
      new Table({
        columnWidths: [2800, 6560],
        rows: [
          new TableRow({ tableHeader: true, children: [
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Property", bold: true, color: "FFFFFF" })] })] }),
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Specification", bold: true, color: "FFFFFF" })] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun({ text: "Language", bold: true })] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Go 1.21+ (goroutines for concurrency)")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun({ text: "Framework", bold: true })] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("Gin or Echo (high-performance HTTP router)")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun({ text: "Deployment", bold: true })] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("ECS Fargate (256 CPU, 512MB RAM per task)")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun({ text: "Scaling", bold: true })] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("Min 2, Max 10 tasks; Step scaling on ALB RequestCountPerTarget")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun({ text: "Endpoints", bold: true })] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("GET /products, GET /products/:id, POST /checkout, GET /orders/:id")] })] })
          ]}),
        ]
      }),

      new Paragraph({ spacing: { before: 200 }, heading: HeadingLevel.HEADING_3, children: [new TextRun("Key Implementation Details")] }),
      new Paragraph({ numbering: { reference: "component-bullets", level: 0 }, children: [new TextRun({ text: "Connection Pooling: ", bold: true }), new TextRun("Use go-redis pool with max 50 connections, min idle 10")] }),
      new Paragraph({ numbering: { reference: "component-bullets", level: 0 }, children: [new TextRun({ text: "Timeouts: ", bold: true }), new TextRun("Redis operations 50ms, SQS publish 100ms, total request 500ms")] }),
      new Paragraph({ numbering: { reference: "component-bullets", level: 0 }, children: [new TextRun({ text: "Idempotency: ", bold: true }), new TextRun("Store idempotency keys in Redis with 24h TTL")] }),
      new Paragraph({ numbering: { reference: "component-bullets", level: 0 }, children: [new TextRun({ text: "Correlation IDs: ", bold: true }), new TextRun("Generate UUID per request, propagate through all services")] }),

      // Redis Section
      new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun("Redis (ElastiCache)")] }),
      new Paragraph({ spacing: { after: 100 }, children: [new TextRun("The heart of the system - handles all synchronous inventory operations.")] }),

      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Data Structures")] }),
      
      new Table({
        columnWidths: [2200, 2200, 4960],
        rows: [
          new TableRow({ tableHeader: true, children: [
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Key Pattern", bold: true, color: "FFFFFF" })] })] }),
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Type", bold: true, color: "FFFFFF" })] })] }),
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Purpose", bold: true, color: "FFFFFF" })] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun({ text: "inv:{product_id}", font: "Courier New", size: 20 })] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Integer")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Current available inventory count")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun({ text: "res:{order_id}", font: "Courier New", size: 20 })] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("Hash")] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("Reservation details (product_id, qty, user_id, timestamp)")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun({ text: "lock:{product_id}", font: "Courier New", size: 20 })] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("String")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Distributed lock (for pessimistic strategy)")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun({ text: "idem:{key}", font: "Courier New", size: 20 })] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("String")] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("Idempotency key with order result")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun({ text: "prod:{product_id}", font: "Courier New", size: 20 })] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Hash")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Product cache (name, price, description)")] })] })
          ]}),
        ]
      }),

      new Paragraph({ spacing: { before: 200 }, heading: HeadingLevel.HEADING_3, children: [new TextRun("Critical: Atomic Reservation Lua Script")] }),
      new Paragraph({ spacing: { after: 100 }, children: [new TextRun("This script ensures no overselling by making check-and-decrement atomic:")] }),

      // Lua script as code block
      new Paragraph({ spacing: { before: 100 }, children: [new TextRun({ text: "-- reserve_inventory.lua", font: "Courier New", size: 18, color: "666666" })] }),
      new Paragraph({ children: [new TextRun({ text: "local inv_key = KEYS[1]           -- inv:{product_id}", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "local res_key = KEYS[2]           -- res:{order_id}", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "local quantity = tonumber(ARGV[1])", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "local order_id = ARGV[2]", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "local user_id = ARGV[3]", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "local ttl = tonumber(ARGV[4])     -- reservation TTL (300s)", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "local current = tonumber(redis.call('GET', inv_key) or 0)", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "if current < quantity then", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    return {0, 'INSUFFICIENT_INVENTORY', current}", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "end", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "redis.call('DECRBY', inv_key, quantity)", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "redis.call('HSET', res_key, 'qty', quantity, 'user', user_id, 'ts', redis.call('TIME')[1])", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "redis.call('EXPIRE', res_key, ttl)", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "return {1, 'RESERVED', current - quantity}", font: "Courier New", size: 18 })] }),

      new Paragraph({ spacing: { before: 200 }, heading: HeadingLevel.HEADING_3, children: [new TextRun("Configuration")] }),
      new Paragraph({ numbering: { reference: "redis-bullets", level: 0 }, children: [new TextRun({ text: "Instance: ", bold: true }), new TextRun("cache.r6g.large (2 vCPU, 13GB RAM)")] }),
      new Paragraph({ numbering: { reference: "redis-bullets", level: 0 }, children: [new TextRun({ text: "Cluster Mode: ", bold: true }), new TextRun("Disabled for simplicity (single-node sufficient for 500 users)")] }),
      new Paragraph({ numbering: { reference: "redis-bullets", level: 0 }, children: [new TextRun({ text: "Reservation TTL: ", bold: true }), new TextRun("300 seconds (5 min) - auto-releases on payment timeout")] }),
      new Paragraph({ numbering: { reference: "redis-bullets", level: 0 }, children: [new TextRun({ text: "Persistence: ", bold: true }), new TextRun("AOF with appendfsync everysec")] }),

      // SQS Section
      new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun("SQS Message Queue")] }),
      new Paragraph({ spacing: { after: 100 }, children: [new TextRun("Decouples the fast checkout response from slow order processing.")] }),

      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Queue Configuration")] }),
      
      new Table({
        columnWidths: [3200, 3200, 2960],
        rows: [
          new TableRow({ tableHeader: true, children: [
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Queue", bold: true, color: "FFFFFF" })] })] }),
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Type", bold: true, color: "FFFFFF" })] })] }),
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Purpose", bold: true, color: "FFFFFF" })] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("flash-sale-orders")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Standard (high throughput)")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Primary order processing")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("flash-sale-orders-dlq")] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("Dead Letter Queue")] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("Failed orders for manual review")] })] })
          ]}),
        ]
      }),

      new Paragraph({ spacing: { before: 200 }, heading: HeadingLevel.HEADING_3, children: [new TextRun("Message Schema")] }),
      new Paragraph({ children: [new TextRun({ text: "{", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: '  "order_id": "uuid-v4",', font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: '  "product_id": "string",', font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: '  "user_id": "string",', font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: '  "quantity": 1,', font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: '  "correlation_id": "uuid-v4",', font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: '  "timestamp": "ISO-8601",', font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: '  "idempotency_key": "string"', font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "}", font: "Courier New", size: 18 })] }),

      new Paragraph({ spacing: { before: 200 }, children: [new TextRun({ text: "Why SQS over RabbitMQ: ", bold: true }), new TextRun("For this project scope, SQS provides simpler setup (no broker management), native AWS integration, and sufficient throughput. RabbitMQ would add complexity without proportional benefit at 500 user scale.")] }),

      // Order Processor Section
      new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun("Order Processor Service")] }),
      new Paragraph({ spacing: { after: 100 }, children: [new TextRun("Consumes orders from SQS, processes payments, and persists to RDS.")] }),

      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Processing Logic")] }),
      new Paragraph({ children: [new TextRun({ text: "func processOrder(msg OrderMessage) error {", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    // 1. Check idempotency", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: '    if exists := redis.Get("idem:" + msg.IdempotencyKey); exists {', font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "        return nil // Already processed", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    }", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    // 2. Simulate payment (configurable 100-500ms)", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    if err := simulatePayment(msg); err != nil {", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "        releaseReservation(msg.OrderID)", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "        return err", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    }", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    // 3. Persist to RDS (with retry)", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    tx := db.Begin()", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    tx.Create(&Order{...})", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: '    tx.Exec("UPDATE products SET stock = stock - ? WHERE id = ?", ...)', font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    tx.Commit()", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    // 4. Cleanup reservation & set idempotency", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: '    redis.Del("res:" + msg.OrderID)', font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: '    redis.SetEx("idem:" + msg.IdempotencyKey, "completed", 24*time.Hour)', font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    return nil", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "}", font: "Courier New", size: 18 })] }),

      // Circuit Breaker Section
      new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun("Circuit Breaker Pattern")] }),
      new Paragraph({ spacing: { after: 100 }, children: [new TextRun("Prevents cascade failures when downstream services are unhealthy.")] }),

      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Circuit Breaker Placement")] }),
      
      new Table({
        columnWidths: [2400, 2400, 4560],
        rows: [
          new TableRow({ tableHeader: true, children: [
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Circuit", bold: true, color: "FFFFFF" })] })] }),
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Threshold", bold: true, color: "FFFFFF" })] })] }),
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Behavior When Open", bold: true, color: "FFFFFF" })] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("API → Redis")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("5 failures in 10s")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Return 503, queue fails immediately")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("API → SQS")] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("3 failures in 10s")] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("Rollback Redis reservation, return 503")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Processor → RDS")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("5 failures in 30s")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Stop polling SQS, wait for half-open")] })] })
          ]}),
        ]
      }),

      new Paragraph({ spacing: { before: 200 }, children: [new TextRun({ text: "Go Implementation: ", bold: true }), new TextRun("Use sony/gobreaker or afex/hystrix-go library.")] }),

      // Auto-scaling Section
      new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun("Auto-Scaling Configuration")] }),
      new Paragraph({ spacing: { after: 100 }, children: [new TextRun("Step scaling is preferred over target tracking for flash sale traffic patterns.")] }),

      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Step Scaling Policy (Recommended)")] }),
      new Paragraph({ numbering: { reference: "scale-bullets", level: 0 }, children: [new TextRun({ text: "Metric: ", bold: true }), new TextRun("ALBRequestCountPerTarget (not CPU - traffic spikes faster than CPU)")] }),
      new Paragraph({ numbering: { reference: "scale-bullets", level: 0 }, children: [new TextRun({ text: "Scale Out: ", bold: true }), new TextRun("100 req/target → +2 tasks, 200 req/target → +4 tasks")] }),
      new Paragraph({ numbering: { reference: "scale-bullets", level: 0 }, children: [new TextRun({ text: "Scale In: ", bold: true }), new TextRun("< 50 req/target for 5 min → -1 task")] }),
      new Paragraph({ numbering: { reference: "scale-bullets", level: 0 }, children: [new TextRun({ text: "Cooldown: ", bold: true }), new TextRun("Scale out 30s, scale in 300s")] }),

      new Paragraph({ spacing: { before: 200 }, children: [new TextRun({ text: "Why Step > Target Tracking: ", bold: true }), new TextRun("Target tracking reacts to sustained load. Step scaling can add multiple tasks on first spike detection, critical for thundering herd scenarios.")] }),

      // Database Section
      new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun("RDS PostgreSQL")] }),
      new Paragraph({ spacing: { after: 100 }, children: [new TextRun("Source of truth for orders and final inventory reconciliation.")] }),

      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Schema")] }),
      new Paragraph({ children: [new TextRun({ text: "CREATE TABLE products (", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    name VARCHAR(255) NOT NULL,", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    price DECIMAL(10,2) NOT NULL,", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    stock INTEGER NOT NULL CHECK (stock >= 0),", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    version INTEGER DEFAULT 1  -- for optimistic locking", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: ");", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "CREATE TABLE orders (", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    id UUID PRIMARY KEY,", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    user_id VARCHAR(255) NOT NULL,", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    product_id UUID REFERENCES products(id),", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    quantity INTEGER NOT NULL,", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    status VARCHAR(20) DEFAULT 'completed',", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    created_at TIMESTAMP DEFAULT NOW(),", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "    correlation_id UUID", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: ");", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "CREATE INDEX idx_orders_user ON orders(user_id);", font: "Courier New", size: 18 })] }),
      new Paragraph({ children: [new TextRun({ text: "CREATE INDEX idx_orders_product ON orders(product_id);", font: "Courier New", size: 18 })] }),

      new Paragraph({ spacing: { before: 200 }, heading: HeadingLevel.HEADING_3, children: [new TextRun("Configuration")] }),
      new Paragraph({ numbering: { reference: "config-bullets", level: 0 }, children: [new TextRun({ text: "Instance: ", bold: true }), new TextRun("db.r6g.large (2 vCPU, 16GB RAM)")] }),
      new Paragraph({ numbering: { reference: "config-bullets", level: 0 }, children: [new TextRun({ text: "Multi-AZ: ", bold: true }), new TextRun("Enabled for Experiment 3 failover testing")] }),
      new Paragraph({ numbering: { reference: "config-bullets", level: 0 }, children: [new TextRun({ text: "Connection Pool: ", bold: true }), new TextRun("Max 100 connections (use pgbouncer if needed)")] }),

      // Failure Recovery Section
      new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun("Failure Recovery Mechanisms")] }),

      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Scenario: Redis Failure")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("Circuit breaker opens after 5 consecutive failures")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("All checkout requests return 503 immediately")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("Half-open state tests every 30s")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("On recovery: sync inventory from RDS → Redis")] }),

      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Scenario: RDS Failover")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("Order Processor circuit opens, stops SQS polling")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("Messages remain in SQS (visibility timeout 5 min)")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("Multi-AZ failover completes (~60-120s)")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("Circuit half-opens, resumes processing")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("Idempotency keys prevent duplicate orders")] }),

      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Scenario: Reservation Timeout")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("Reservation TTL expires (300s)")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("Background job detects expired reservations")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("Inventory counter incremented (INCR)")] }),
      new Paragraph({ numbering: { reference: "failover-bullets", level: 0 }, children: [new TextRun("Order marked as 'expired' if still in SQS")] }),

      // Monitoring Section
      new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun("Observability & Monitoring")] }),

      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Key Metrics (CloudWatch)")] }),
      new Paragraph({ numbering: { reference: "monitoring-bullets", level: 0 }, children: [new TextRun({ text: "Business: ", bold: true }), new TextRun("Orders completed, orders failed, inventory levels, checkout latency P50/P95/P99")] }),
      new Paragraph({ numbering: { reference: "monitoring-bullets", level: 0 }, children: [new TextRun({ text: "Infrastructure: ", bold: true }), new TextRun("ECS task count, CPU/memory utilization, ALB request count, 5xx rate")] }),
      new Paragraph({ numbering: { reference: "monitoring-bullets", level: 0 }, children: [new TextRun({ text: "Redis: ", bold: true }), new TextRun("Operations/sec, cache hit rate, memory usage, evictions")] }),
      new Paragraph({ numbering: { reference: "monitoring-bullets", level: 0 }, children: [new TextRun({ text: "SQS: ", bold: true }), new TextRun("Messages visible, messages in-flight, age of oldest message")] }),
      new Paragraph({ numbering: { reference: "monitoring-bullets", level: 0 }, children: [new TextRun({ text: "Circuit Breakers: ", bold: true }), new TextRun("State (closed/open/half-open), trip count")] }),

      new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun("Correlation ID Flow")] }),
      new Paragraph({ spacing: { after: 200 }, children: [new TextRun("Every request generates a UUID correlation ID at ALB, propagated via X-Correlation-ID header through all services and into CloudWatch logs. This enables end-to-end request tracing.")] }),

      // Summary Section
      new Paragraph({ heading: HeadingLevel.HEADING_1, children: [new TextRun("Implementation Priority")] }),
      new Paragraph({ spacing: { after: 100 }, children: [new TextRun("Recommended implementation order to achieve project objectives:")] }),

      new Table({
        columnWidths: [1200, 3600, 4560],
        rows: [
          new TableRow({ tableHeader: true, children: [
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Priority", bold: true, color: "FFFFFF" })] })] }),
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Component", bold: true, color: "FFFFFF" })] })] }),
            new TableCell({ borders: cellBorders, shading: headerShading, children: [new Paragraph({ children: [new TextRun({ text: "Enables Experiment", bold: true, color: "FFFFFF" })] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ alignment: AlignmentType.CENTER, children: [new TextRun({ text: "1", bold: true })] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Redis Lua atomic reservation")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Exp 1: High-Contention Locking")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ alignment: AlignmentType.CENTER, children: [new TextRun({ text: "2", bold: true })] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("SQS integration + Order Processor")] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("Decouples checkout from processing")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ alignment: AlignmentType.CENTER, children: [new TextRun({ text: "3", bold: true })] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Step scaling on RequestCount")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Exp 2: Autoscaling Under Thundering Herd")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ alignment: AlignmentType.CENTER, children: [new TextRun({ text: "4", bold: true })] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("Circuit breakers")] })] }),
            new TableCell({ borders: cellBorders, shading: altRowShading, children: [new Paragraph({ children: [new TextRun("Exp 3: Cascading Failure Recovery")] })] })
          ]}),
          new TableRow({ children: [
            new TableCell({ borders: cellBorders, children: [new Paragraph({ alignment: AlignmentType.CENTER, children: [new TextRun({ text: "5", bold: true })] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Idempotency + Correlation IDs")] })] }),
            new TableCell({ borders: cellBorders, children: [new Paragraph({ children: [new TextRun("Data consistency verification")] })] })
          ]}),
        ]
      }),

      new Paragraph({ spacing: { before: 300 }, children: [
        new TextRun({ text: "This architecture targets your stated goals: 500+ concurrent users, zero overselling, sub-100ms P95 latency for the critical checkout path, and resilience under chaos engineering scenarios.", italics: true })
      ]}),
    ]
  }]
});