from pymongo import Connection
from gridfs import GridFS as gfs
import sys
import os

if __name__ == '__main__':
	db = Connection("152.146.38.56").sox_scripts
	fs = gfs(db)
	path = os.path.join(sys.path[0], "client_script.py") # script path
	if not os.path.isfile(path):
		raise Error("No file named sox_sophos.py at path %s" % sys.path[0])
		sys.exit(2)
	
	soph = open(path, "r")
	
	try:
		file_id = fs.put(soph, filename="script")
		# file_id.close()
	except:
		print "error!"
		sys.exit(2)
		
	print "Sucess!"
	soph.close()