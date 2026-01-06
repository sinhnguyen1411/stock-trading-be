<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <title>SCS Architecture Diagram</title>

    <!-- Mermaid JS (Latest 10.x) -->
    <script type="module">
      import mermaid from "https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs";
      mermaid.initialize({ startOnLoad: true });
    </script>

    <style>
        body {
            background: #f5f6fa;
            padding: 20px;
            font-family: Arial, sans-serif;
        }
        h2 {
            font-size: 24px;
            font-weight: bold;
        }
        .container {
            background: white;
            padding: 20px;
            border-radius: 12px;
            box-shadow: 0 4px 10px rgba(0,0,0,0.1);
        }
    </style>
</head>

<body>
    <h2>Supply Chain Security Architecture – Automated Pipeline</h2>

    <div class="container">
        <div class="mermaid">
flowchart LR

    %% ====== LAYER 1: DEVELOPER ======
    Dev[Developer] -->|git push| Repo[Source Code Repository]

    %% ====== LAYER 2: CI/CD & SCS ======
    Repo -->|Trigger Pipeline| CI[CI/CD Pipeline]

    CI --> Build[Build Go Binary and Docker Image]
    Build --> SBOM[Generate SBOM (Syft)]
    SBOM --> Scan[Scan Vulnerabilities (Grype)]

    Scan -->|Fail: High or Critical CVE| FailBuild[Pipeline Fails - Unsafe Image Blocked]
    Scan -->|Pass| Sign[Sign Image (Cosign)]
    Sign --> Attest[Create SLSA Provenance Attestation]

    Sign --> Registry[Secure Container Registry]
    Attest --> Registry

    %% ====== LAYER 3: KUBERNETES ======
    Dev -->|Deploy (kubectl / Helm)| K8s[Kubernetes Cluster]

    K8s --> AC[Admission Controller / Kyverno]
    Reg
