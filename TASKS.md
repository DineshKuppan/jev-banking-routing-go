# Tasks

## Active

- [ ] **Phase 1: Build Go banking-routing project** - Implement and test all source files, verify unit tests pass
  - [ ] Create directory structure and go.mod
  - [ ] Implement domain/transaction.go
  - [ ] Implement jev/client.go (HTTP contract with TypeSafe)
  - [ ] Implement routing/policy.go (deterministic decisions)
  - [ ] Implement service/router.go with fallback logic
  - [ ] Implement cmd/api/main.go (HTTP API)
  - [ ] Add testdata/scenarios.json and router_test.go
  - [ ] Add Dockerfile, Makefile, .env.example
  - [ ] Run tests: `go test ./...` (verify 4 policy tests pass)
  - [ ] Verify compilation and structure
- [ ] **Phase 2: Draft integrated narrative** - Combine Jev concepts with Go code into blog structure
  - [ ] Extract conceptual sections from PLAN-DRAFT-BLOG.md
  - [ ] Map Go code snippets to each narrative section
  - [ ] Weave real-world applications (5 domains)
  - [ ] Target 3,200–3,800 words
- [ ] **Phase 3: Polish for Medium** - Format for broad audience, add visual aids
  - [ ] Break up paragraphs (Medium ~3 sentence preference)
  - [ ] Add frequent subheadings
  - [ ] Verify all code examples compile and are accurate
  - [ ] Add flow diagrams and tables
  - [ ] Include curl command with expected output
- [ ] **Phase 4: Graph & Publish** - Create knowledge graph of Go project, then publish
  - [ ] Run `/graphify jev-banking-routing-go/` to visualize architecture
  - [ ] Create Medium account/workspace if needed
  - [ ] Prepare headline, subtitle, tags (AI, Machine Learning, TypeSafe, Go, Banking)
  - [ ] Publish to Medium with GitHub repo link in CTA

## Waiting On

## Someday

## Done
