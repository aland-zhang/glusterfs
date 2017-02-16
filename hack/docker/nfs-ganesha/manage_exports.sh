#!/bin/bash

#This script is called by glusterd when the user
#tries to export a volume via NFS-Ganesha.
#An export file specific to a volume
#is created in GANESHA_DIR/export.d .
# Adapted from https://github.com/gluster/glusterfs/tree/master/extras/ganesha/scripts
# and https://github.com/nfs-ganesha/nfs-ganesha/issues/94

OPTION=$1        # add | remove
if [ "$OPTION" = "add" ]; then
    GLUSTER_HOST=$2  # GlusterFS Host IP
    VOL=$3           # Volume Name
fi
if [ "$OPTION" = "remove" ]; then
    VOL=$2           # Volume Name
fi

GANESHA_DIR=/etc/ganesha
CONF=$GANESHA_DIR/ganesha.conf
EXPORT_CONF=$GANESHA_DIR/export.d/$VOL.conf
declare -i EXPORT_ID

function check_cmd_status()
{
	if [ "$1" != "0" ]; then
		rm -rf $EXPORT_CONF
		exit 1
	fi
}

#This function keeps track of export IDs and increments it with every new entry
function dynamic_export_add()
{
    count=$(ls -l $GANESHA_DIR/export.d/*.conf | wc -l)
    if [ $count -eq 0 ] ; then
        EXPORT_ID=2
    else
		EXPORT_ID=`cat $GANESHA_DIR/.export_added`
		check_cmd_status `echo $?`
		EXPORT_ID=EXPORT_ID+1
    fi
    echo $EXPORT_ID > $GANESHA_DIR/.export_added
    check_cmd_status `echo $?`

	cat >$EXPORT_CONF <<EOL
EXPORT {
    Export_Id = 2;    # Export ID unique to each export
    Path = "/$VOL";   # DOCKER BUG: will export only "/" if some of the commented options below uncommented

    FSAL {
        Name = "GLUSTER";
        Hostname = "$GLUSTER_HOST";  # IP of a node in the GLUSTER trusted pool
        Volume = "$VOL";             # Volume name. Eg: "test_volume"
    }
    Access_type = RW;         # Access permissions
    Squash = No_Root_Squash;  # To enable/disable root squashing
    Disable_ACL = true;       # To enable/disable ACL
    Pseudo = "/$VOL";         # NFSv4 pseudo path for this export. Eg: "/test_volume_pseudo"
    #Protocols = "3,4" ;      # NFS protocols supported
    #Transports = "UDP,TCP" ; # Transport protocols supported
    SecType = "sys";          # Security flavors supported
}
EOL
    sed -i s/Export_Id.*/"Export_Id= $EXPORT_ID ;"/ $EXPORT_CONF
    check_cmd_status `echo $?`
    cmd="dbus-send --print-reply --system \
		--dest=org.ganesha.nfsd /org/ganesha/nfsd/ExportMgr \
		org.ganesha.nfsd.exportmgr.AddExport string:$EXPORT_CONF \
		string:EXPORT(Path=/$VOL)"
	echo $cmd; $cmd
    check_cmd_status `echo $?`
}

#This function removes an export dynamically(uses the export_id of the export)
function dynamic_export_remove()
{
    removed_id=$(cat $EXPORT_CONF | grep Export_Id | cut -d ' ' -f6)
    echo "removed_id = $removed_id"
    check_cmd_status `echo $?`
    cmd="dbus-send --print-reply --system \
		--dest=org.ganesha.nfsd /org/ganesha/nfsd/ExportMgr \
		org.ganesha.nfsd.exportmgr.RemoveExport uint16:$removed_id"
	echo $cmd; $cmd
    check_cmd_status `echo $?`
    rm -rf $EXPORT_CONF
}

RETVAL=0
case "$OPTION" in
	add)
		dynamic_export_add $@
		;;
	remove)
		dynamic_export_remove $@
		;;
	*)	(10)
        echo $"Usage: $0 add gluster_host vol"
        echo $"Usage: $0 remove vol"
		RETVAL=1
esac
exit $RETVAL
