## Integrate Beacon into an Ansible Playbook

This example demonstrates integrating Beacon into an Ansible Playbook.  The Playbook will:
- instantiate a BIG-IP
- provision the BIG-IP via definitions within the [host_var](https://github.com/f5devcentral/f5-beacon/tree/master/examples/ansible-integration/host_vars) folder
- deploy [AS3 appplications](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/host_vars/localhost/apps.json) to the BIG-IP
- update Beacon dashboards based on the AS3 apps deloyed

#### Requirements

- Python env with the dependencies noted in [py_requirements.txt](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/py_requirements.txt) (install all via `pip install -r py_requirements.txt`)
- AWS SSH key, Security Group, VPC
  - these are specified in [all.yaml](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/group_vars/all.yaml)
- AWS EIP
  - specified in [vars.yaml](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/host_vars/localhost/vars.yaml)

#### Output

High level Playbook flow:

- If BIG-IP does not exist in AWS based on tags
  - Create EC2 BIG-IP instance
  - Associate the EIP in host_vars to the new EC2 instance
  - Install Automation Toolchain ([DO](https://github.com/F5Networks/f5-declarative-onboarding), [AS3](https://github.com/F5Networks/f5-appsvcs-extension), [TS](https://github.com/F5Networks/f5-telemetry-streaming))
  - Deploy [DO declaration](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/host_vars/localhost/do.json)
- When the existing/new BIG-IP is online
  - Deploy [AS3 Declaration](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/host_vars/localhost/apps.json)
  - If host_var **bcon_enabled** is set **true**, then run Beacon Ansible Role
    - If a Beacon token for the BIG-IP does not exist, create one
    - Send [TS declaration](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/host_vars/localhost/ts.json) to BIG-IP with the Beacon token
    - Wait ~60 seconds for TS poller to send data to Beacon
    - Compare AS3 declaration and Beacon Applications/Components and true-up (apps defined based on constants class in AS3)

#### Setup

The ansible playbook has a hosts file with a group called **bigips**. The host listed under this group should have a corresponding folder under `host_vars` which holds the definitions of the BIG-IP.

- Update the git_repo var in `group_vars/all.yaml` to point to your customized/forked version. It will update from this as source-of-truth on every run and overwrite local changes. Make sure the host running this playbook has access to pull from this repo.
- Update the python interpreter inside the **hosts** file to the appropriate location.
  - ie: `localhost ansible_python_interpreter=/sc/venvs/np/bin/python`
- Add vault file `../vaults/creds.yaml` with the following secrets and encrypt with your password
  - aws_access_key: `****`
  - aws_secret_key: `****`
  - bcon_un: `**Beacon Username**`
  - bcon_pw: `**Beacon Password**`
  - vpass: `**Desired BIG-IP admin password**`
- **ansible.cfg** by default looks up the vault password at `../vaults/.vault_pass`. This is not required and the vault password may be added in a way which best fits your environment.
- **ansible.cfg** update **log_path** var to point to correct place to log ansible output.
- For each BIG-IP `host_var` update the following:
  - **do.json** with desired Onboard settings (note AWS will reset the hostname based on dhclient).
  - Update **apps.json** with desired AS3 Applications
- Update `group_vars/main.yaml` with AWS env information:
  ```
  ssh_keyfile: "~/.ssh/privatekey.pem"
  ami: ami-07b978ff1e31c60f9
  aws_ssh_key: 4est2
  aws_instance_type: t2.large
  aws_sg: 4est-bip
  aws_vpc_subnet: subnet-0ec6b63e44ba17aad
  region: us-east-2
  ```

#### Beacon Setup

- Beacon role is only used when **bcon_enabled: true** is set within host_vars
- Beacon apps to be updated by the playbook should be defined by the var `bcon_apps`. By default this is set within `group_vars/main.yaml` similar to the below example.
  ```
  bcon_apps:
      App1:
          labels:
          costcenter: yours
          department: Marvel
          location: AZURE
          owner: The Avengers
          region: US East
  ```
- Each AS3 app must have constants set similar to below referencing which Beacon app to associate with them.
  ```
  "App1": {
  "class": "Application",
  "constants": {
      "class": "Constants",
      "beacon": {
          "app_dependency": ["Accounting","HR","App1"]
      }
  },
  ```
- For the playbook to update an App within the Beacon portal, the App must be referenced in the **bcon_apps** var as well as the **AS3** declaration. If it is not referenced in **AS3**, it will be treated as an App with no dependencies (no-op). If it is not referenced in **bcon_apps** it will assume you are not wanting ansible to update that App.

### Assumptions

- App Names in Beacon are unique. Ansible is not currently tracking App IDs, it references based on name.
- Ansible will not manipulate Beacon App dependencies that do not belong to the current BIG-IP the playbook is running against.
- If an Beacon Component is already a dependency for an App, it will not be added a second time.
- There is only 1 health source per Component/App Dependency
- If an OverwBeaconatch App defined in Ansible has no dependencies left, it will remove the App in Beacon

## Run the Playbook

```shell
ansible-playbook deploy.yaml
```
