
import json, sys
as3 = json.loads(sys.argv[1])
bcon_apps = json.loads(sys.argv[2])
bcon_comp = json.loads(sys.argv[3])
env_hostname = sys.argv[4]
as3dict = {}
compdict = {}
appdict = {}
as3_service_apps = {}

beacon_app_declare_deploy = {
  "action": "deploy",
  "declaration": []
}
beacon_app_declare_remove = {
  "action": "remove",
  "declaration": []
}

#####
##### TAKES AS3 DECLARATION AND LOOKS FOR CONSTANTS WITHIN AN AS3 APPLICATION.
##### RETURNS A DICT OF EACH SERVICE AS A DEPENDENCY OF THE APP DEFINED IN CONSTANTS TO PROVIDE TO BEACON
#####

# # Parses AS3 constants class and returns Beacon app association
def as3parse(d):
  for t, tv in d.iteritems():
    if 'class' in tv and tv['class'] == 'Tenant':
      for a, av in tv.iteritems():
        if 'class' in av and av['class'] == 'Application':
          app_dependency = av['constants']['beacon']['app_dependency']
          for s, sv in av.iteritems():
            if 'class' in sv and 'Service_' in sv['class']:
              for bconapp in app_dependency:
                if bconapp not in appdict:
                  appdict[bconapp] = {}
                  appdict[bconapp]['body'] = {}
                  appdict[bconapp]['body']['dependencies'] = []
                  appdict[bconapp]['state'] = "new"
                  appdict[bconapp]['as3_dependencies'] = []
                appdict[bconapp]['as3_dependencies'].append(compdict["/%s/%s/%s" % (t,a,s)])
                if bconapp not in as3_service_apps:
                  as3_service_apps[bconapp] =[]
                as3_service_apps[bconapp].append(compdict["/%s/%s/%s" % (t,a,s)])
                

# Provides Name to ID for Beacon components              
def comp_id_dict(l):
  for c in l:
    compdict[c['name']] = c['id']

# Provides Existing Apps from Beacon
def app_update(l):
  for a in l:
    appdict[a['json']['name']] = {}
    appdict[a['json']['name']]['body'] = a['json']
    appdict[a['json']['name']]['state'] = "existing"
    appdict[a['json']['name']]['as3_dependencies'] = []

# ### Gather data from functions above
comp_id_dict(bcon_comp)
reverse_compdict = {v: k for k, v in compdict.items()}
app_update(bcon_apps)
as3parse(as3['declaration'])

### Prevents adding Components that are already in the Beacon app
### Removes current Components in Beacon apps that are no longer in AS3
for k, v in appdict.iteritems():
  remove_dep = []
  for ind, dep in enumerate(v['body']['dependencies']):
    if dep['healthSourceSettings']:
      for hs in dep['healthSourceSettings']['metrics']:
        cd = hs['tags']['name']
        if cd in appdict[k]['as3_dependencies']:
          appdict[k]['as3_dependencies'].remove(cd)
      # Remove Components in Beacon that are no longer referenced in AS3
        if k in as3_service_apps:
          if cd not in as3_service_apps[k]:
            remove_dep.append(dep)
        else:
          appdict[k]['body']['dependencies'] = []
  for dep in remove_dep:
    appdict[k]['body']['dependencies'].remove(dep)

#Update dependencies in JSON
for app, v in appdict.iteritems():
  for dep in v['as3_dependencies']:
    if appdict[app]['state'] == "existing":
      appdict[app]['body']['dependencies'].append({"name": str(reverse_compdict[dep]),"healthSourceSettings": {"metrics":[{ "measurementName":"BeaconHealth","tags":{"name":str(reverse_compdict[dep]),"source": env_hostname}}]}})
    if appdict[app]['state'] == "new":
      if "dependencies" not in appdict[app]['body']:
        appdict[app]['body']['dependencies'] = []
      appdict[app]['body']['dependencies'].append({"name": str(reverse_compdict[dep]),"healthSourceSettings": {"metrics":[{ "measurementName":"BeaconHealth","tags":{"name":str(reverse_compdict[dep]),"source":env_hostname}}]}})

def transform_body(body, app):
  new_body = {"name": app}
  for key in body:
    if key != body[key]:
      new_body[key] = body[key]
  return new_body

for app in appdict:
  if appdict[app]['body']['dependencies'] == []:
    beacon_app_declare_remove['declaration'].append({
      "application": {
        "name": app
    }})
  else: 
    beacon_app_declare_deploy['declaration'].append({
      "metadata": {
        "version": "v1"
      }, 
      "application": transform_body(appdict[app]['body'], app)
    })

print {
  "deploy_body": beacon_app_declare_deploy,
  "remove_body": beacon_app_declare_remove
}