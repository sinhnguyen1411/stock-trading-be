# DevSecOps CI + Admission Flow

This document explains how to run the supply-chain pipeline (SBOM → scan → sign → attest → push) and enforce it with Kyverno.

## CI Workflow (`.github/workflows/secure-supply-chain.yml`)
Pipeline steps:
1. Build Go binary via multi-stage Dockerfile and run `go test ./...`.
2. Generate SBOM with Syft (`sbom.spdx.json`).
3. Scan with Grype; fail on High/Critical; publish `grype-report.json`.
4. Sign image with Cosign keyless (OIDC) and attach SBOM.
5. Emit SLSA-style provenance predicate and attestation via Cosign.
6. Push image + artifacts to GHCR and upload SBOM/scan/Cosign bundle as workflow artifacts.
7. Compute deploy annotations from pipeline outputs:
   - `security.grype.io/high_critical`: total High+Critical from Grype (`high_critical` output).
   - `security.stock-trading.dev/sbom-digest`: SHA256 of the SBOM (`sbom_digest` output).
   The workflow renders a ready-to-use Kustomize overlay under `deploy/kubernetes/overlays/ci/` and uploads it as an artifact for deployment.

Required secrets/permissions:
- `GITHUB_TOKEN` (default) with `packages:write` and `id-token:write` permissions.
- For private registries or key-based signing, provide `COSIGN_PASSWORD` and `COSIGN_PRIVATE_KEY` (extend workflow accordingly).

Trigger manually (`workflow_dispatch`) or on push/PR to main/master.

## Admission Policies (Kyverno)
Resources under `deploy/policies/kyverno/`:
- `cosign-public-key.yaml`: ConfigMap for the Cosign public key (replace with your key or use keyless verification once supported).
- `clusterpolicy-verify-images.yaml`: Enforces Cosign signature + SLSA provenance for `ghcr.io/sinhnguyen1411/stock-trading/user-service:*`.
- `clusterpolicy-cve-threshold.yaml`: Requires annotation `security.grype.io/high_critical: "0"` on Pods (set from CI results).
- `clusterpolicy-require-sbom.yaml`: Requires annotation `security.stock-trading.dev/sbom-digest` on Pods.

Apply with:
```bash
kubectl apply -k deploy/policies/kyverno
```

Populate the Cosign key:
```bash
kubectl -n kyverno create configmap cosign-public-key --from-file=cosign.pub=./cosign.pub
```

## Local Demo (Kind)
Use `scripts/devsecops_kind_bootstrap.sh`:
```bash
COSIGN_PUB_PATH=./cosign.pub ./scripts/devsecops_kind_bootstrap.sh
```
This creates a Kind cluster, installs Kyverno, publishes the Cosign key, and applies the policies.

## Deploying the Service with Required Annotations
When deploying, ensure Pods carry:
- `security.grype.io/high_critical: "<value from CI>"` (expect 0; pipeline should fail otherwise).
- `security.stock-trading.dev/sbom-digest: "<sha256-of-sbom-or-oci-ref>"` (from CI output).

You can use the generated overlay artifact (`deploy/kubernetes/overlays/ci/`) created by the workflow:
```bash
kubectl apply -k deploy/kubernetes/overlays/ci
```
If you deploy outside the workflow context, set these annotations manually or regenerate the overlay with your values.

If using Kustomize, you can add these under `metadata.annotations` in the Deployment template.

## Verifying Enforcement
- Try to deploy an unsigned or unannotated image: Pod admission should be denied with Kyverno messages.
- Deploy the signed, scanned image with annotations: Pod should be admitted.
- Validate signatures/attestations manually:
  ```bash
  cosign verify --keyless ghcr.io/sinhnguyen1411/stock-trading/user-service:dev
  cosign verify-attestation --type slsaprovenance --keyless ghcr.io/sinhnguyen1411/stock-trading/user-service:dev
  ```

## Evidence Collection
- CI artifacts: SBOM (`sbom.spdx.json`), scan report (`grype-report.json`), Cosign bundle (`provenance.json`).
- Admission denials: `kubectl events` or Kyverno audit logs showing rejection reasons.
- Successful deploy: `kubectl get pods -n <ns>` plus Kyverno status to show policy passes.
