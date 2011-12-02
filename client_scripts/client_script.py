import os, plistlib, time, sys, socket
from datetime import datetime, date
from pymongo.connection import Connection
from pymongo.database import Database
from uuid import getnode as get_mac
import platform
import subprocess

# import logging
# logging.basicConfig(filename=os.path.join(sys.path[0], "sox.log"),
#     level=logging.DEBUG,
#     format='%(asctime)s %(levelname)s: %(message)s',
#     datefmt='%d/%m/%Y %I:%M:%S')

import pprint
pp = pprint.PrettyPrinter(indent=4)

debug = False

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

def visit(arg, dirname, names):
	if len(names) == 0:
		return
	i = 0
	# tmp = copy.deepcopy(names)
	while i < len(names):
		dir = names[i]
		folder = "%s/%s" % (dirname,dir)
		if len(dir) > 4 and dir[-4:] == ".app":
			plist = folder + "/Contents/Info.plist"
			vs = "N/A"
			if os.path.isfile(plist):
				vs = plist_version(plist)
			arg.append({
			   "Path":dirname,
			   "Name":dir,
			   "Version":vs
			})
			
            # (dirname,dir,vs) 
			del names[i]
		else:
			i += 1

def walk(root="/Applications"):	
	args = []
	os.path.walk(root, visit, args)
	return args

def installed_apps(doc):
	apps = walk()
	doc.update( {"Apps":apps} )

def sophos_dict(doc):
    v_def, mtime = log_information()
    return doc.update({
        'Virus_version':plist_version('/Applications/Sophos Anti-Virus.app/Contents/Info.plist'),
        'Virus_def':v_def,
        'Virus_last_run':mtime,
    })

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
        'Firewall':state,
        # 'signed_apps':apps
    })

# Simple check for the Recon LaunchAgent
def recon_dict(doc):
    if os.path.isfile("/Library/LaunchDaemons/com.wpp.recon.plist"):
        doc.update({
        'Recon':True
        })
    else:
        doc.update({
        'Recon':False
        })

def machine_dict(doc):
    # machine specific info
    # old_serial = "N/A"
    serial = "N/A"
    for l in subprocess.Popen(["/usr/sbin/system_profiler","SPHardwareDataType"],
            stdout=subprocess.PIPE).communicate()[0].split("\n")[4:]:
        # if debug: print l
        if "Serial Number (system)" in l:
            serial = l.split(": ")[-1]
        # if "Serial Number (system)" in l:
        #     serial = l.split(": ")[-1]
        if "Model Identifier:" in l:
            model_id = l.split(": ")[-1]
        if "Processor Name:" in l:
            cpu_model = l.split(": ")[-1]
        if "Memory:" in l:
            memory = l.split(": ")[-1]
        if "Processor Speed" in l:
            mhz = l.split(': ')[-1]
    l = subprocess.Popen(["sw_vers"],
        stdout=subprocess.PIPE).communicate()[0].split("\n")
    
    osx_vers = "OSX %s (%s)" % (l[1].split(":\t")[-1],l[2].split(":\t")[-1])
    
    if serial == "N/A":
        print "invalid serial!"
    
    doc.update({
        '_id':serial,
        # 'Old_serial':old_serial,
        'Osx':str(osx_vers),
        'Model_id':model_id,
        'Hostname':subprocess.Popen(["/usr/sbin/scutil","--get", "ComputerName"],stdout=subprocess.PIPE).communicate()[0].split("\n")[0],
        # 'os_version':platform.mac_ver()[0],
        'Cpu':cpu_model + " " + mhz,
        'Memory':memory,
        # 'mac':hex(get_mac())[:-1],
        'Ip':socket.gethostbyname(socket.gethostname())
    })

def users():
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
    if not users:
        users.append['xadmin']
    return users

def serial_number():
    # hardware = os.popen("/usr/sbin/system_profiler SPHardwareDataType | grep \"Serial Number\" | awk '{print $4}'").read()[:-1]
    for l in subprocess.Popen(["/usr/sbin/system_profiler","SPHardwareDataType"],
            stdout=subprocess.PIPE).communicate()[0].split("\n"):
        if "Serial Number" in l:
            return l.split(" ")[-1]
    print "debug: failed to read serial number"
    # sys.exit(2)

def mongo_conn(ip,db='sox'):
	try:
		return Database(Connection(ip), db)
	except:
	    print "debug: Could not connect to mongo database at ip %s" % ip
        sys.exit(2)

# Update/Insert db.collection with doc
def update_db(db, doc, coll="main"):
    if not coll:
        return
    col = db[coll]
    try:
        id = doc['_id']
    except:
        # print "update_db: 'doc' has no '_id'", doc
        print "debug: doc has no '_id'"
    	sys.exit(2)
    	
    # if doc["Old_serial"] != "N/A" and doc["Old_serial"] != doc["_id"]:
    #     print doc["Old_serial"], doc["_id"]
    #     try:
    #         dups = col.remove({"_id": doc["Old_serial"]}, safe=True)
    #     except:
    #         print "No duplicates"
    #         logging.debug("No duplicates")
    #     logging.debug("removed duplicate _id = %s" % doc["_id"] )
    #     print "test_removed duplicate doc for %s" % doc["Hostname"]
                
    old_doc = col.find_one({"_id":id})
    if not old_doc:
        try:
            col.insert(doc, safe=True)
        except:
            print "debug: Failed to insert doc"
            sys.exit(2)
    else:
        try:
            col.update({'_id':id},doc, safe=True)
        except:
        	print "Failed to update doc:", doc["Hostname"]
        	sys.exit(2)

def main():
    # static IP for the mini-server 
    server_ip = "152.146.38.56"
    # sox database is _for now_ simply "sox"
    main_db = "sox"
    collection = "dict_scripts"
    db = mongo_conn(server_ip,db=main_db)
    # 
    date = datetime.today()
    doc = {
        'Date': date,
        'Users':users(),
    }
    machine_dict(doc)
    sophos_dict(doc)
    security_dict(doc)
    installed_apps(doc)
    recon_dict(doc)
    update_db(db,doc, coll=collection)
    # print "debug: Successfully registered machine data"
    # db.drop_collection('main')

if __name__ == '__main__':
	main()