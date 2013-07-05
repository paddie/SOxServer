#!/usr/bin/env python
# encoding: utf-8

import os, plistlib, time, sys, socket
from datetime import datetime, date
from uuid import getnode as get_mac
import platform
import subprocess
import tempfile
import httplib
import json

# sophos antivirus log is in binary format => convert to xml1
def plistFromPath(plist_path):
    # convertPlist(plist_path, 'xml1')
    if os.system('plutil -convert xml1 '+ plist_path) != 0:
        print 'failed to convert plist from path: ', plist_path
        sys.exit(1)
    try:
    	return plistlib.plistFromPath(plist_path)
    except AttributeError: # there was an AttributeError, we may need to use the older method for reading the plist
    	try:
    		return plistlib.Plist.fromFile(plist_path)
    	except:
    		print 'failed to read plist'
    		sys.exit(5)

def convertToXML(path):
    # convertPlist(plist_path, 'xml1')
	tmp_path = os.path.join("/var/tmp", "com.application_walking_tmp.plist")
    # tmp_path = "/Library/AdPeople/com.application_walking_tmp.plist"
	subprocess.call(['cp', path, tmp_path])
	if os.system('plutil -convert xml1 '+ tmp_path) != 0:
		raise Exception("Could not convert binary plist to xml1")
	
	plist = plistlib.readPlist(tmp_path)
	subprocess.call(['rm', tmp_path])
	return plist

def runProcess(exe):    
    p = subprocess.Popen(exe, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    while(True):
        retcode = p.poll() #returns None while subprocess is running
        line = p.stdout.readline()
        yield line
        if(retcode is not None):
            break

"""
Software Update Tool
Copyright 2002-2009 Apple

Software Update found the following new or updated software:
   * iTunesX-11.0.1
    iTunes (11.0.1), 193391K [recommended]
"""
def softwareupdate(doc):
    # subprocess.check_output(*popenargs, **kwargs)
    collect_info = False
    pending_updates = ""
    for l in runProcess(["softwareupdate", "-l"]):
        print l
        if "No new software available" in l:
            doc.update({"softwareupdate":False})
            return

        if collect_info is False and "Software Update found" in l:
            collect_info = True

        if collect_info:
            pending_updates += l

    print "pending updates: ", pending_updates

    doc.update({
        "softwareupdate":True,
        "softwareoutput":pending_updates
        })
     # = subprocess.Popen(["softwareupdate", "-l"],).split("\n")


def plist_version(path):
	plist = "N/A"
	try:
		plist = plistlib.readPlist(path)
	except:
	    try:
			plist = convertToXML(path)
	    except:
	        return "N/A"
	try:
		return plist["CFBundleShortVersionString"]
	except:
		return "N/A"

def installed_apps(doc):
	# apps = walk()
    # tf = tempfile.TemporaryFile("w+b")
    apps = subprocess.Popen(["/usr/sbin/system_profiler","-xml","SPApplicationsDataType"],stdout=subprocess.PIPE).communicate()[0]
    # tf.write(apps)
    # tf.seek(0)
    plist = plistlib.readPlistFromString(apps)
    apps = plist[0]["_items"]
    for i in xrange(0,len(apps)):
        date = apps[i].get('lastModified', None)
        if date is not None:
            apps[i]['lastModified'] = date.isoformat()
        else:
            apps[i]['lastModified'] = None

    doc.update( {"apps":apps} )

def sophos_dict(doc):
    if not os.path.isfile('/Applications/Sophos Anti-Virus.app/Contents/Info.plist'):
        return doc.update({
            'virus_version':"N/A",
            'virus_def':"N/A",
            'virus_last_run':"N/A"})    

    version = plist_version('/Applications/Sophos Anti-Virus.app/Contents/Info.plist')
    v_def, mtime = log_information()
    return doc.update({
        'virus_version':version,
        'virus_def':v_def,
        'virus_last_run':mtime,
    })

def log_information(path='/Library/Logs/Sophos Anti-Virus.log'):
    if os.path.isfile(path):
        log = open(path, 'r')
        for lines in log:
            if 'com.sophos.intercheck: Version' in lines:
                vers = lines
        mtime = time.strftime("%d/%m/%y",time.localtime(os.path.getmtime(path)))
        log.close()
        
        return vers.split(": ")[1].split(",")[0], mtime
        # return vers[31:-15], mtime
    else:
        # if no log is at this position
        return "N/A", "N/A"

def firewall_state(path='/Library/Preferences/com.apple.alf.plist'):
	plist = convertToXML(path)
	apps = []
	try:
	    for app in plist['applications']:
	        try:
	            apps.append(app['bundleid'])
	        except:
	            pass
	    return plist['globalstate'], apps
	except:
	    return 0, apps

def security_dict(doc):
    firewall, apps = firewall_state()
    # ENABLE FIREWALL IFF OFF
    # if not firewall: # firewall off
    #     os.system("defaults write /Library/Preferences/com.apple.alf globalstate -int 1")
    if firewall == 0:
        state = False
    else:
        state = True
    doc.update({
        'firewall':state,
        # 'signed_apps':apps
    })

# Simple check for the Recon LaunchAgent
def recon_dict(doc):
    old = "/Library/Application Support/JAMF/scripts/submitInventory.sh"
    new = "/Library/Application Support/WPP/Inventory/scripts/submitInventory.sh"

    doc.update({'recon':os.path.isfile(new)})
    

def machine_dict(doc):
    # machine specific info
    profile = subprocess.Popen(["/usr/sbin/system_profiler","-xml","SPHardwareDataType"], stdout=subprocess.PIPE).communicate()[0]
    # read xml into plit-file, and ignore irrelevant data..
    machine = plistlib.readPlistFromString(profile)[0]["_items"][0]

    # *******************
    # OSX version and build
    # *******************
    l = subprocess.Popen(["sw_vers"],
        stdout=subprocess.PIPE).communicate()[0].split("\n")
    osx_vers = "OSX %s (%s)" % (l[1].split(":\t")[-1],l[2].split(":\t")[-1])
    
    # *******************************
    # IP - more complicated than it sounds..
    # *******************************
    ips = socket.gethostbyname_ex(socket.gethostname())[2]
    ip = ""
    for i in ips:
        # if on the work network, the third IP value is either 38 or 210 or 113
        if i.split(".")[2] in ['38', '210', '113']:
            ip = i
            break
    if ip == "":
        # if on home network, the ip might not be xx.xx.[38/210].xx
        # - simply go with the first of the ip's
        ip = socket.gethostbyname(socket.gethostname())
    
    # *****************************
    # HOSTNAME - also a bit stupid
    # *****************************
    hostname = subprocess.Popen(["/usr/sbin/scutil","--get", "ComputerName"],stdout=subprocess.PIPE).communicate()[0].split("\n")[0]

    doc.update({
        'serial':machine["serial_number"],
        # 'Old_serial':old_serial,
        'osx':str(osx_vers),
        'model':machine["machine_model"],
        'hostname':hostname,
        'cpu':"%s %s" % (machine["cpu_type"], machine["current_processor_speed"]),
        'cores':machine["number_processors"],
        'memory':machine["physical_memory"],
        'ip':ip
    })

def users():
    # lists all folders '/Users'
    # - discards: Shared and any files in that folder
    users = []
    for folder in os.listdir('/Users'):
        # ignore files
		if not folder == 'Shared' and os.path.isdir('/Users/'+folder):
		    users.append(folder)
    if os.path.isdir("/Domain/PeopleGroup.Internal/Users"):
        for folder in os.listdir('/Domain/PeopleGroup.Internal/Users'):
            # ignore files
    		if os.path.isdir('/Domain/PeopleGroup.Internal/Users/'+folder):
    		    users.append(folder)

    return users

def mongo_conn(ip,db='sox'):
	try:
		return Database(Connection(ip), db)
	except:
	    print "debug: Could not connect to mongo database at ip %s" % ip
        sys.exit(2)

def postMachineSpecs(ip, doc):
    params = json.dumps(doc)
    try:
        headers = {"Content-type": "application/x-www-form-urlencoded",
                "Accept": "text/plain"}
        conn = httplib.HTTPConnection(ip)
        conn.request("POST", "/updateMachine/", params, headers)
        print "SOX script: Success!"
    except Exception:
        print "Couldn't connect to webserver on ip: ", ip
        print "Retrying in an hour.."
    # urllib2.urlopen("localhost:6060/updateMachine", jdata)

def main():
    server_ip = "152.146.38.56:6060" # static IP for the mini-server 
    # server_ip = "localhost:6060"
    # main_db = "sox" # db name
    # collection = "machines" # collection name
    
    # db = mongo_conn(server_ip,db=main_db)
    today = datetime.now()
    doc = {
        'date': today.strftime("%m/%d/%y"),
        'datetime':int(time.time()), # iso 1970
        'time':today.strftime("%H:%M:%S"),
        'users':users(),
        "script_v" :  subprocess.Popen(["/usr/bin/git","describe"],stdout=subprocess.PIPE).communicate()[0][:-1],
        "script_d" : "Updated Recon Revision",
    }
    machine_dict(doc)
    sophos_dict(doc)
    security_dict(doc)
    # installed_apps(doc)
    recon_dict(doc)
    # softwareupdate(doc)
    print doc

    # print "softwareupdate: ", doc["softwareupdate"]
    # print "debug: Successfully registered machine data"
    # pp.pprint(doc)
    # postMachineSpecs(server_ip, doc)

if __name__ == '__main__':
	main()