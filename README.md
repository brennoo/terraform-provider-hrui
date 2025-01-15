[![HRUI Provider](docs/project-logo.png)](#)
<p align="center" style="font-size: 1.5em;">
    <em>Terraform provider for HRUI switches</em>
</p>

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/2af02dd9c60141b2b9142291693fde15)](https://app.codacy.com/gh/brennoo/terraform-provider-hrui/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/2af02dd9c60141b2b9142291693fde15)](https://app.codacy.com/gh/brennoo/terraform-provider-hrui/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)
[![CodeQL](https://github.com/brennoo/terraform-provider-hrui/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/brennoo/terraform-provider-hrui/actions/workflows/github-code-scanning/codeql)
[![Release](https://github.com/brennoo/terraform-provider-hrui/actions/workflows/release.yml/badge.svg)](https://github.com/brennoo/terraform-provider-hrui/actions/workflows/release.yml)
---

This is a hobby project to create a Terraform provider for [HRUI](www.hruitech.com) switches that are web-managed. It relies on [goquery](https://github.com/PuerkitoBio/goquery) for interfacing with the web UI.

Check the documentation at:

- Terraform: [HRUI Provider](https://registry.terraform.io/providers/brennoo/hrui)
- OpenToFu: [HRUI Provider](https://search.opentofu.org/provider/brennoo/hrui)

## HRUI ODM/OEM

This provider is developed using a Horaco (ZX-SWTG124AS) switch. Other brands that are likely to work with this provider, as they seem to be the same hardware, include:

* HRUI
* Horaco
* Sodola
* XikeStor
* AmpCom

> [!NOTE]
> Using firmware v1.9.

## Getting Started

1.  Configure the provider in your Terraform configuration:

    ```terraform
    terraform {
      required_providers {
        hrui = {
          source  = "brennoo/hrui"
          version = "= 0.1.0-alpha.1"
        }
      }
    }

    provider "hrui" {
      url      = "http://192.168.2.1"
      username = "admin"
      # The password can be set in HRUI_PASSWORD environment variable
    }
    ```

2.  Refer to the [examples](examples) folder for usage of available resources and data sources.

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) file for guidelines.

## License

This provider is licensed under the MPL-2.0 License.
