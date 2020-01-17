# Ansible Role: Integrate F5 Beacon with AS3

Configures [Telemetry Streaming](https://clouddocs.f5.com/products/extensions/f5-telemetry-streaming/latest/) on BIG-IP and then syncs AS3 apps between BIG-IP and Beacon.

## Requirements

None

## Role Variables

Specifies the AS3 declaration in JSON format (see [example](../../host_vars/<your_desired_BIG-IP_hostname>/apps.json)):

    as3json

Specifies the apps that should be updated in Beacon (see [example](../../vars/beacon_apps.yaml)):

    bcon_apps

## Dependencies

- A BIG-IP version 12.0.0 or newer, configured with the F5 Automation Toolchain (handled via the f5devcentral.f5app_services_package role)

## Example Usage

    - name: Update Beacon
      include_role:
        name: f5devcentral.f5_as3_beacon
      when: bcon_enabled |bool

## License

Apache

## Author Information

This role was created in 2019 by [focrensh](https://github.com/focrensh).
