#!/bin/bash
export > /etc/envvars

ip=$(hostname -i)
if [ ! -f "$PVROOT/.installed" ] || [ $(cat "$PVROOT/.installed") != "$ip" ]; then
	mkdir -p $PVROOT/volumes
	rm -rf $PVROOT/glusterd
	mv /var/lib/glusterd $PVROOT
	echo $ip > "$PVROOT/.installed"
fi

rm -rf /var/lib/glusterd
ln -sv -t /var/lib $PVROOT/glusterd

echo "Starting runit..."
exec /usr/sbin/runsvdir-start
