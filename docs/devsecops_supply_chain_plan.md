# DevSecOps Supply-Chain Upgrade Plan

This document maps the thesis requirements onto the current `stock-trading-be` user-service and defines the concrete technical work needed to turn the service into a demonstrable, enforceable supply-chain-security baseline.

## 1. Context & Scope
- **Service under test**: existing Golang user-service (register/verify/login/etc.) packaged as a Docker container, deployed to Kubernetes.
- **Threat model**: dependency compromise, vulnerable base images, tampered build artifacts, and unauthorized workloads within the cluster.
- **Success criteria**: every artifact (image + metadata) is produced by an automated pipeline that emits SBOM + vulnerability report + provenance, signs the result, and Kubernetes Gatekeeper/Kyverno rejects anything that fails these checks.

## 2. Target Architecture
```
Developer -> Git push -> CI/CD (build + SBOM + scan + sign + attest) -> Registry
                                                   |
                                                   v
                                     Security metadata (SBOM, SARIF, attestations)

Kubernetes Admission Controller <- Policy engine verifies: cosign signature, SLSA provenance, CVE score, SBOM hash
```

Key components:
1. **SBOM generation**: Syft or CycloneDX CLI for `go.mod` + resulting image.
2. **Vulnerability scanning**: Grype/Trivy runs on the SBOM + image; produces SARIF/JSON for auditing.
3. **Signing and attestation**: Sigstore Cosign issues image signatures and SLSA-style provenance attestations; upload to Rekor transparency log when online.
4. **Registry**: GHCR/Artifact Registry/Harbor storing signed images plus attached SBOM/attestations (via cosign `attach sbom|attestation`).
5. **Admission enforcement**: Kyverno or OPA Gatekeeper using cosign policy controller (or Kyverno verifyImages) to block unsigned or non-compliant images, plus checks for SBOM/attestation annotations and CVE policy.
6. **Demo cluster**: Kind or Minikube with cert-manager + Kyverno + Cosign root pubkey distributed as ConfigMap/Secret.

## 3. Pipeline Blueprint (repeatable for Golang microservices)
1. **Build & Test Stage**
   - Multi-stage Dockerfile compiling Go binary with `GOOS=linux` and running as a non-root user.
   - `go test ./...` + static analyzers (optional `gosec`).
2. **SBOM Stage**
   - Run `syft packages <image>` to emit SPDX JSON.
   - Store artifact as `sbom.spdx.json`; upload to pipeline artifacts + `cosign attach sbom`.
3. **Vulnerability Scan Stage**
   - Execute `grype sbom:sbom.spdx.json --fail-on High` (configurable threshold for High/Critical).
   - Fail pipeline when findings exceed threshold; publish SARIF for observability.
4. **Signing + Attestation Stage**
   - `cosign sign --key $COSIGN_KEY_REFP` for the built image digest.
   - `cosign attest --predicate slsa-provenance.json --type slsaprovenance` capturing commit SHA, builder ID, tool versions.
   - Push both signature and attestation to registry/Rekor.
5. **Release Stage**
   - Only promote/push image if all previous stages pass.
   - Annotate Git tags/releases with SBOM + scan summary for auditing.
6. **Promotion/Deploy Stage**
   - `kubectl apply` or ArgoCD/Flux pulls signed image.
   - Admission controller re-validates cosign signature, ensures attestation presence, and checks vulnerability status via policy.

## 4. Tooling Decisions
| Requirement | Tool | Notes |
|-------------|------|-------|
| SBOM | `syft` (anchore) | Generates SPDX/CycloneDX, easy container integration. |
| Vulnerability scan | `grype` or `trivy` | Both can consume SBOM; pick one and pin version. |
| Signing/Attestation | `cosign` + Sigstore | Supports keyless (OIDC) or key-based; start with key pair stored in GitHub/Azure secret. |
| Admission policy | Kyverno `verifyImages` + custom rules | Native integration with cosign, can also inspect annotations (SBOM digests, CVE score). |
| Cluster | Kind + local registry mirror | Enables deterministic, reproducible demo. |

## 5. Implementation Phases
1. **Baseline Hardening**
   - Review Dockerfile (multi-stage build, non-root run, pinned base image digest).
   - Capture runtime config as Helm/Manifest templates with image digests instead of tags.
2. **CI/CD Enablement**
   - Add GitHub Actions/Azure DevOps pipeline definition with the stages above.
   - Store cosign keys (or configure keyless with workload identity).
   - Publish SBOM + scan artifacts per build.
3. **Security Gates**
   - Encode vulnerability threshold as pipeline guard (`FAIL_ON=High`).
   - Block merge/release if guard fails, ensuring "risky artifact never published".
4. **Registry + Metadata**
   - Choose registry (e.g., GHCR). Configure `cosign` to push signatures/attestations + `cosign triangulate` for retrieval.
   - Maintain metadata index (simple JSON) linking git SHA -> image digest -> SBOM hash.
5. **Kubernetes Enforcement**
   - Provision Kind cluster with Kyverno + cosign policy controller.
   - Distribute cosign public key via Kubernetes Secret/ConfigMap.
   - Policies:
     1. Reject unsigned images or mismatched public key.
     2. Require attestation with provenance fields (builder ID, commit SHA).
     3. Block deployment if vulnerability report annotation indicates High/Critical > 0.
     4. (Optional) Validate SBOM digest label matches `cosign attach sbom`.
6. **Demo & Evaluation Assets**
   - Write scripts (`scripts/devsecops_demo.sh|ps1`) to:
     1. Build "clean" image through pipeline, deploy successfully.
     2. Attempt to deploy tampered/unsigned image -> Admission denied.
     3. Attempt to deploy image with forced fake CVE annotation -> denied.
   - Collect CI logs + Kubernetes events + Kyverno audit logs as evidence.
7. **Documentation**
   - Update README with quickstart; add diagrams + policy samples.
   - Provide reproducible instructions for other Go services to reuse the pipeline.

## 6. Demo & Evaluation Criteria
- **Pipeline enforcement**: Provide CI logs demonstrating automatic failure when `grype` finds disallowed CVEs.
- **Trust verification**: Show `cosign verify` output + Rekor entry for released image.
- **Admission control**: Capture `kubectl describe` events showing rejection reasons (unsigned, missing SBOM, vulnerability threshold).
- **Repeatability**: Document parameterization so another Go service can plug into same workflow by setting image name + module path.

## 7. Next Steps Checklist
1. Author hardened Dockerfile + Helm/kustomize manifests.
2. Commit CI workflow skeleton with placeholder secrets (SBOM + scan + sign steps).
3. Automate Kind/Kyverno bootstrap via `scripts/devsecops_cluster_up.sh`.
4. Draft Kyverno policy set and commit under `deploy/policies/`.
5. Extend README with quick demo instructions referencing this plan.

Deliverables from these steps will satisfy the thesis objectives: automated SBOM, proactive vulnerability gating, signed artifacts with provenance, Kubernetes-level enforcement, and reproducible documentation for future services.
