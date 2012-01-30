#!/usr/bin/env python
# encoding: utf-8
"""
untitled.py

Created by Patrick Madsen on 2011-06-27.
Copyright (c) 2011 __MyCompanyName__. All rights reserved.
"""
from pymongo import Connection
from gridfs import GridFS as gfs
import sys
import os
# import logging
import subprocess
# logging.basicConfig(filename='sox.log',
#     level=logging.DEBUG,
#     format='%(asctime)s %(levelname)s: %(message)s',
#     datefmt='%d/%m/%Y %I:%M:%S')

def serial_number():
    # hardware = os.popen("/usr/sbin/system_profiler SPHardwareDataType | grep \"Serial Number\" | awk '{print $4}'").read()[:-1]
    for l in subprocess.Popen(["/usr/sbin/system_profiler","SPHardwareDataType"],
            stdout=subprocess.PIPE).communicate()[0].split("\n"):
        if "Serial Number" in l:
            return l.split(" ")[-1]
    print "error: failed to read serial number"

def main():
	db = Connection("152.146.38.56").sox_scripts
	fs = gfs(db)

	path = os.path.join(sys.path[0], "script.py") # script path
	id = os.path.join(sys.path[0], "script_id") # version path
	
	# print path
	# print id
	try:
		new_script = fs.get_last_version("script")
	except:
		print "error: Cold not connect to mongodb or no script in DB"
		sys.exit(2)	
	
	# read old script id
	if os.path.isfile(id):
		tmp = open(id, "r")
		old_id = tmp.read()
		tmp.close()
		if str(new_script._id) == old_id:
			if os.path.isfile(path):
				try:
					os.execl(sys.executable, "python", path)
				except:
					print "error: Script not executed properly, aborting.."
					sys.exit(2)
				return
	
	print "debug: New version of script, updating script.."
	
	script = open(path, "w") # overwrite script everytime
	id_file = open(id, "w")
	if not os.path.isfile(path) and not os.path.isfile(id):
		print "error: Script and/or script_id not created properly"
		sys.exit(2)
	reruns = 2
	while reruns > 0:
		try:
			s = new_script.read()
			break
		except:
			print "error: Error reading new script - retrying.. "
			if reruns == 1:
				print "error: Error reading new script - aborting"
				sys.exit(2)
				
	script.write(s)

    # os.chown(path,0,0)
    # os.chmod(path, 755)
	
    # os.chown(id, 0,0)
    # os.chmod(id,644)
	
	id_file.write(str(new_script._id))
	
	script.close()
	id_file.close()
	print "debug: Script downloaded, resetting machine data in db.."
	col = Connection("152.146.38.56").sox["main"]
	col.remove({"_id":serial_number()})
	print "debug: running script"
	
	try:
		os.execl(sys.executable, "python", path)
	except:
		print "error: Script not executed properly, aborting.."
		sys.exit(2)
	return

if __name__ == '__main__':
	main()

