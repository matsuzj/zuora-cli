# Security Policy

## Supported versions

`zuora-cli` ships as a single statically-linked binary via Homebrew and GitHub
Releases. Only the **latest release** receives security fixes — please upgrade
(`brew upgrade zr`) and confirm with `zr version` before reporting.

| Version        | Supported          |
| -------------- | ------------------ |
| Latest release | :white_check_mark: |
| Older releases | :x:                |

## Reporting a vulnerability

**Please do not open a public issue for security vulnerabilities.**

Use GitHub's private vulnerability reporting:

1. Open the [Security tab](https://github.com/matsuzj/zuora-cli/security).
2. Click **Report a vulnerability**.
3. Include the affected version (`zr version`), reproduction steps, and impact.

We aim to acknowledge reports within a few business days and will keep you
updated as we investigate and ship a fix. Reporters who wish to be named will be
credited once a fix is released.

## Scope and hardening notes

`zuora-cli` is a command-line client for the Zuora API. The most relevant
concerns are credential handling, output that could leak secrets, and the
release supply chain:

- **Credentials** are read from environment variables (`ZR_CLIENT_ID` /
  `ZR_CLIENT_SECRET`) or a no-echo prompt. Passing `--client-secret` on the
  command line is discouraged because it lands in shell history (the CLI warns
  about this).
- **Verbose output** (`-v` / `-vv`) is designed never to echo the client secret
  or issued tokens; a regression test pins this.
- **Release artifacts** carry a signed [build-provenance attestation][prov].
  Verify a download with:

  ```
  gh attestation verify <archive> --repo matsuzj/zuora-cli
  ```

[prov]: https://github.com/matsuzj/zuora-cli/attestations
