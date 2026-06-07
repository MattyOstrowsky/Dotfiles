---
description: Architecture review of infrastructure design
agent: architect
subtask: true
---
Perform an architectural review of the infrastructure in $ARGUMENTS (or current directory).

## Review Scope
1. **Component Analysis:** Map all infrastructure components and their relationships
2. **Failure Modes:** For each component, identify single points of failure
3. **Scalability:** Can each component scale horizontally? What are the bottlenecks?
4. **Security Boundaries:** Are network segments properly isolated? Is least privilege enforced?
5. **Data Flow:** How does data move between components? Where are the trust boundaries?
6. **Cost:** Are there over-provisioned or idle resources?

## Output: ADR Format
```
# ADR: Infrastructure Review — [project name]

## Status: [Review/Accepted/Deprecated]

## Context
[What is the current state of the infrastructure]

## Assessment
[Findings organized by: Architecture, Security, Reliability, Cost]

## Recommendations
[Prioritized list of changes: P0 Critical, P1 High, P2 Medium]

## Consequences
[Impact of implementing vs not implementing recommendations]
```

## Mandatory Questions (if not answered by the code)
- What is the expected RTO/RPO?
- What is the blast radius of each component failure?
- What is the monthly cost budget?
- How many 9s of availability are required?
