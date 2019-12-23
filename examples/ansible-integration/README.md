## Integrate Beacon into an Ansible playbook

This example demonstrates integrating Beacon into an Ansible playbook.  The playbook will:
- instantiate a BIG-IP
- provision the BIG-IP via definitions within the [host_vars](host_vars) folder
- deploy [AS3 appplications](host_vars/localhost/apps.json) to the BIG-IP
- update Beacon dashboards based on the AS3 apps deloyed

#### Requirements

- Python env with the dependencies noted in [py_requirements.txt](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/py_requirements.txt) (install all via `pip install -r py_requirements.txt`)
- AWS SSH key, Security Group, VPC
  - specified in [all.yaml](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/group_vars/all.yaml)
- AWS EIP
  - specified in [vars.yaml](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/host_vars/localhost/vars.yaml)

#### Output

High level playbook flow:

- If BIG-IP does not exist in AWS based on tags
  - Create EC2 BIG-IP instance
  - Associate the EIP in host_vars to the new EC2 instance
  - Install Automation Toolchain ([DO](https://github.com/F5Networks/f5-declarative-onboarding), [AS3](https://github.com/F5Networks/f5-appsvcs-extension), [TS](https://github.com/F5Networks/f5-telemetry-streaming))
  - Deploy [DO declaration](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/host_vars/localhost/do.json)
- When the existing/new BIG-IP is online
  - Deploy [AS3 Declaration](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/host_vars/localhost/apps.json)
  - If host_var **bcon_enabled** is **true**, then run Beacon Ansible Role
    - If a Beacon token for the BIG-IP does not exist, create one
    - Send [TS declaration](https://github.com/f5devcentral/f5-beacon/blob/master/examples/ansible-integration/host_vars/localhost/ts.json) to BIG-IP with the Beacon token
    - Wait ~60 seconds for TS poller to send data to Beacon
    - Compare AS3 declaration and Beacon Applications/Components and true-up (apps defined based on constants class in AS3)

#### Setup

The Ansible playbook has a hosts file with a group called **bigips**.  The host listed under this group should have a corresponding folder under `host_vars` which holds the definition for the BIG-IP.

- Update the git_repo var in `group_vars/all.yaml` to point to your customized/forked version.  It will update from the specified repo as the source-of-truth on every run and overwrite local changes.  Make sure the host running this playbook has access to pull from the specified repo.
- Update the python interpreter inside the **hosts** file to the appropriate location.
  - e.g.: `localhost ansible_python_interpreter=/sc/venvs/np/bin/python`
- Add vault file `../vaults/creds.yaml` with the following secrets:
  - aws_access_key: `****`
  - aws_secret_key: `****`
  - bcon_username: `**Beacon Username**`
  - bcon_password: `**Beacon Password**`
  - vpass: `**desired BIG-IP admin password**`
- by default, **ansible.cfg** looks up the vault password at `../vaults/.vault_pass`.  This is not required and the vault password may be added in a way which best fits your environment.
- update **log_path** in **ansible.cfg** to point to correct place to log Ansible output
- for each BIG-IP `host_var` update the following:
  - **do.json** with desired onboarding settings (note: AWS will reset the hostname based on dhclient)
  - **apps.json** with desired AS3 applications
- Update `group_vars/all.yaml` with AWS env information:
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

- Beacon role is only used when **bcon_enabled** is **true** within host_vars
- For Beacon apps to be updated, they must be defined in `bcon_apps` within `group_vars/all.yaml` similar to the below example.
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
- Each AS3 app must have constants set similar to below referencing the associated Beacon app(s).
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
- For the playbook to update an app within Beacon, the app must be referenced in both **bcon_apps** and the **AS3** declaration.  If it is not referenced in the **AS3** declaration, it will be treated as an app with no Beacon dependencies (no-op).  If it is not referenced in **bcon_apps**, it will assume you are not wanting Ansible to update that app.

### Assumptions

- App names in Beacon are unique.  Ansible is not currently tracking app IDs -- apps are referenced by name.
- Ansible will not manipulate Beacon app dependencies that do not belong to the BIG-IP the playbook is running against.
- If a Beacon component is already a dependency for an app, it will not be added a second time.
- There is only one health source per component/app dependency.
- If a Beacon app defined in Ansible has no dependencies left, it will remove the app in Beacon.

## Run the Playbook

```shell
ansible-playbook deploy.yaml
```
