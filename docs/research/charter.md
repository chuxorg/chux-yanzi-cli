# CAP-200 Research Charter

## Operational Supervision for Autonomous and Semi-Autonomous Work

**Status:** Draft Research Charter

**Purpose**

This charter defines the research objectives, methodology, scope, and success criteria for CAP-200. It intentionally precedes architectural design and implementation. Its purpose is to ensure that Yanzi evolves from evidence and observation rather than assumptions or market enthusiasm.

---

# Problem Statement

The emergence of large language models and autonomous software agents has dramatically reduced the cost of executing many forms of knowledge work. Writing code, generating documentation, performing analysis, producing designs, and interacting with external systems can now be delegated to software at unprecedented speed.

Execution, however, is only one component of productive work.

As delegated work increases, new operational challenges emerge:

* Understanding what work is currently being performed.
* Determining why a worker made a decision.
* Intervening when work begins to diverge.
* Recovering after interruption or failure.
* Preserving evidence and accountability.
* Coordinating humans and automated workers.
* Maintaining governance across long-running operations.

The central hypothesis of CAP-200 is that **supervision**, rather than execution, is becoming the primary operational constraint.

---

# Research Hypothesis

**Primary Hypothesis**

> As autonomous and semi-autonomous systems become capable of performing increasingly valuable work, the scarce capability becomes operational supervision rather than execution.

Execution continues to become less expensive.

Supervision becomes more valuable.

---

# Research Questions

CAP-200 seeks to answer the following questions:

1. What is operational supervision?
2. How do successful organizations supervise delegated work today?
3. Which supervision concepts remain constant across industries?
4. What operational responsibilities cannot be delegated?
5. Which existing software categories solve portions of this problem?
6. Which gaps remain unsolved?
7. Which standards define relevant operational concepts?
8. What responsibilities should belong to an operational supervision platform?
9. Which responsibilities should explicitly remain outside such a platform?
10. Does this represent a durable software category?

---

# Scope

The research is intentionally broader than software engineering.

Software engineering serves as the initial credibility domain because it is experiencing rapid adoption of AI-assisted work.

The resulting operational model should be applicable to any domain involving delegated work, including but not limited to:

* Software Engineering
* IT Operations
* Cybersecurity
* Manufacturing
* Healthcare
* Scientific Research
* Financial Operations
* Logistics
* Government
* Military Command and Control

---

# Research Methodology

The project will favor observation over assumption.

Research activities include:

* Literature review
* Industry standards review
* Product capability analysis
* Historical technology comparisons
* Cross-industry operational studies
* Terminology analysis
* Architectural pattern analysis

Conclusions must be supported by multiple independent observations whenever possible.

---

# Research Principles

The following principles govern CAP-200.

## 1. Assume We Are Wrong

Research exists to invalidate assumptions before implementation.

Contradictory evidence is welcomed.

---

## 2. Observe Before Designing

Architecture follows understanding.

Understanding does not follow architecture.

---

## 3. Borrow Before Inventing

When mature terminology or concepts already exist, prefer adopting them over creating new vocabulary.

---

## 4. Separate Observations From Conclusions

Raw observations should be recorded independently from interpretation.

Patterns emerge through accumulation.

---

## 5. Study Operations, Not Technology

The research focuses on operational behavior rather than specific AI models or programming technologies.

Technologies change.

Operational responsibilities change much more slowly.

---

## 6. Generalize Carefully

Generalization should occur only after multiple domains demonstrate common patterns.

---

# Categories of Research

## Operational Disciplines

Examples include:

* Industrial automation
* Manufacturing execution
* Network operations
* Security operations
* Air traffic control
* Incident command
* Military command and control
* Site Reliability Engineering

---

## Existing Software Categories

Research should examine:

* AI coding assistants
* Agent frameworks
* Workflow orchestration
* Observability platforms
* Knowledge management
* Project management
* DevOps platforms
* Operational dashboards

The objective is not to compare features but to identify responsibilities and boundaries.

---

## Standards

Research should evaluate relevant standards and specifications including:

* ISA-95
* ISO/IEC 42001
* NIST AI Risk Management Framework
* OpenTelemetry
* OpenAPI
* OAuth 2.0 / OpenID Connect
* Model Context Protocol (MCP)

Additional standards may be added as research progresses.

---

# Non-Goals

CAP-200 is not intended to:

* Design a workflow engine.
* Design an AI model.
* Create an orchestration framework.
* Define prompt engineering techniques.
* Replace existing project management systems.
* Produce implementation plans prematurely.

---

# Expected Deliverables

The research effort is expected to produce:

1. Position Paper
2. Market Landscape Analysis
3. Standards Mapping
4. Operational Vocabulary and Glossary
5. Capability Matrix
6. Responsibility Model
7. Reference Architecture
8. Product Position Statement

---

# Success Criteria

CAP-200 is successful if it can answer the following questions with evidence:

* Does a meaningful operational problem exist?
* Is the problem distinct from existing software categories?
* Are organizations already assembling partial solutions?
* Are common operational principles emerging?
* Can Yanzi define a coherent responsibility within this ecosystem?
* Can the proposed architecture remain valuable even as AI models evolve?

---

# Initial Thesis

The history of computing demonstrates that every major technological shift eventually requires new operational disciplines.

The widespread adoption of AI-assisted work appears to represent another such shift.

If execution becomes abundant while accountability, coordination, governance, recovery, and supervision remain scarce, then operational supervision becomes an independent and valuable discipline.

CAP-200 exists to determine whether that hypothesis is true and, if so, to define the principles upon which such a discipline should be built.
