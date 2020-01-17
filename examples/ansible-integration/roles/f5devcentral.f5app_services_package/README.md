# Ansible Role: Install F5 Automation Toolchain on BIG-IP

Performs a series of steps needed to download and install RPM packages on the BIG-IP that are a part of 
F5 automation toolchain. It can also remove the said packages from the BIG-IP system.

## Requirements

None

## Role Variables

Available variables are listed below. For their default values, see `defaults/main.yml`:

    provider:
      server: localhost
      server_port: 443
      user: admin
      password: secret
      validate_certs: false
      transport: rest

Establishes initial connection to the specified **BIG-IP** to install/remove RPM packages. This should be updated just as you would when using ``provider`` with a module directly.

    f5app_services_package_state: present
    
When set to ``present`` the requested RPM package will be downloaded and installed or installed on the BIG-IP.
The ``absent`` value will remove the requested RPM package from the BIG-IP.

    f5app_services_package_path

Path to the RPM package on the Ansible controller. This parameter is used when **downloading** a package or just
**installing** RPM packages that are already downloaded. When **removing** package if the full path to the package 
is provided, the filename will be cherry picked to properly remove the package.

    f5app_services_package_url

The full url of the requested RPM package to be downloaded for installation on the BIG-IP.

    f5app_services_package_checksum_url

The full url to the **SHA256** checksum file, used when checksum verification of the RPM download is required.

    f5app_services_package_checksum
    
The RPM **SHA265** checksum string, used when checksum verification of the RPM download is required. 
Mutually exclusive with ``f5app_services_package_checksum_url``.

## Dependencies

* A BIG-IP TMOS version **12.0.0** and **above**.
  
* Installed **RPM** tool on the Ansible controller.

## Example Playbooks

The role can be used in couple of ways depending on the specified variables. The choice of method will depend on you.

* Download and Install
* Install
* Remove

## Download and Install

    - name: Download and Install RPMs
      hosts: download
      any_errors_fatal: true
    
      tasks:
        - name: Download and Install AS3 RPM no sha check
          include_role:
            name: f5devcentral.f5app_services_package
          vars:
            f5app_services_package_url: "https://github.com/F5Networks/f5-appsvcs-extension/raw/master/dist/latest/f5-appsvcs-3.9.0-3.noarch.rpm"
            f5app_services_package_path: "/tmp/f5-appsvcs-3.9.0-3.noarch.rpm"
    
        - name: Download and Install DO RPM sha check - url
          include_role:
            name: f5devcentral.f5app_services_package
          vars:
            f5app_services_package_url: "https://github.com/F5Networks/f5-declarative-onboarding/raw/master/dist/f5-declarative-onboarding-1.3.0-4.noarch.rpm"
            f5app_services_package_checksum_url: "https://github.com/F5Networks/f5-declarative-onboarding/raw/master/dist/f5-declarative-onboarding-1.3.0-4.noarch.rpm.sha256"
            f5app_services_package_path: "/tmp/f5-declarative-onboarding-1.3.0-4.noarch.rpm"
    
        - name: Download and Install TS RPM sha check - no_url
          include_role:
            name: f5devcentral.f5app_services_package
          vars:
            f5app_services_package_url: "https://github.com/F5Networks/f5-telemetry-streaming/raw/master/dist/f5-telemetry-1.1.0-1.noarch.rpm"
            f5app_services_package_destination: "/tmp/f5-telemetry-1.1.0-1.noarch.rpm"
            f5app_services_package_checksum: "7fdad5ff409ca7068f75a19c38d1b06d3f4facb86afce15976af63b963c03e29"

## Install

    - name: Install RPMs
      hosts: install
      any_errors_fatal: true
    
      tasks:
        - name: Install AS3 RPM
          include_role:
            name: f5devcentral.f5app_services_package
          vars:
            f5app_services_package_path: "/tmp/f5-appsvcs-3.9.0-3.noarch.rpm"
    
        - name: Install DO RPM
          include_role:
            name: f5devcentral.f5app_services_package
          vars:
            f5app_services_package_path: "/tmp/f5-declarative-onboarding-1.3.0-4.noarch.rpm"
    
        - name: Install TS RPM
          include_role:
            name: f5devcentral.f5app_services_package
          vars:
            f5app_services_package_path: "/tmp/f5-telemetry-1.1.0-1.noarch.rpm"


## Remove

    - name: Remove RPMs
      hosts: remove
      any_errors_fatal: true
    
      tasks:
        - name: Remove AS3 RPM
          include_role:
            name: f5devcentral.f5app_services_package
          vars:
            f5app_services_package_path: "/tmp/f5-appsvcs-3.9.0-3.noarch.rpm"
    
        - name: Remove DO RPM
          include_role:
            name: f5devcentral.f5app_services_package
          vars:
            f5app_services_package_path: "/tmp/f5-declarative-onboarding-1.3.0-4.noarch.rpm"
    
        - name: Remove TS RPM
          include_role:
            name: f5devcentral.f5app_services_package
          vars:
            f5app_services_package_path: "/tmp/f5-telemetry-1.1.0-1.noarch.rpm"

## License

Apache

## Author Information

This role was created in 2019 by [Wojciech Wypior](https://github.com/wojtek0806).

[1]: https://galaxy.ansible.com/f5devcentral/f5app_services_package
